package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	maxWorkers = 5
	queueName  = "download"
)

func failOnError(msg string, err error) {
	if err != nil {
		log.Panicf("Error: %s;%s", msg, err)
	}
}

func findMinium(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func visualizerTask() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError("No se pudo establecer la conexion", err)
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError("No se pudo crear el canal", err)
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"download",
		false,
		false,
		false,
		false,
		nil,
	)

	failOnError("No se pudo crear la cola", err)

	for {
		var wg sync.WaitGroup
		queueInspector, err := ch.QueueInspect(queueName)
		failOnError("No se pudo inspeccionar la cola", err)

		numWorkers := findMinium(queueInspector.Messages, maxWorkers)

		for task := 0; task < numWorkers; task++ {
			wg.Add(1)
			go ConsumingMessages(&wg, ch, &queue, "Processor")
		}
		wg.Wait()
	}
}

func main() {
	fmt.Println("Consumiendo...")
	visualizerTask()
}

func ConsumingMessages(wg *sync.WaitGroup, ch *amqp.Channel, queue *amqp.Queue, tag string) {
	defer wg.Done()

	// 1. Conexión a RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError("No se pudo establecer la conexión a RabbitMQ", err)
	defer conn.Close()

	ch, err = conn.Channel()
	failOnError("No se pudo abrir un canal", err)
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"download",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError("No se pudo declarar la cola", err)

	msgs, err := ch.Consume(
		q.Name,
		tag,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError("No se pudo consumir los mensajes", err)

	for msg := range msgs {
		bodyRequest := make(map[string]string)
		err := json.Unmarshal(msg.Body, &bodyRequest)
		failOnError("No se pudo deserializar", err)

		currentTime := time.Now()
		searchQuery := "ytsearch:" + bodyRequest["songname"]

		cmd := exec.Command("yt-dlp",
			"-x",
			"--geo-bypass",
			"--age-limit", "99",
			"--audio-format", "mp3",
			searchQuery,
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Error: en la función 'ConsumingMessages', no se pudo leer la salida;", err)
			fmt.Println(string(output))

			if strings.Contains(string(output), "Sign in to confirm you're not a bot") {
				fmt.Println("Se ha detectado un bloqueo por YouTube")
				fileLog, err := os.OpenFile("./logs/blocked.json", os.O_CREATE|os.O_APPEND|os.O_EXCL, os.ModeAppend)
				if err != nil {
					fmt.Println("Error: No se pudo abrir 'logs.json';", err)
					os.Mkdir("logs", os.ModeAppend)
				}
				fileLog.WriteString(fmt.Sprintf("Fecha: %s; Blocked by YouTube.\n", currentTime))
			}
			msg.Nack(false, false)
			continue
		}

		fmt.Printf("✅ Descarga completada para: %s\n", bodyRequest["songname"])
		msg.Ack(false)

		for key, value := range bodyRequest {
			fmt.Println(key, value, "\n")
		}
		fmt.Println("Mensaje original:", string(msg.Body))
		return
	}
}
