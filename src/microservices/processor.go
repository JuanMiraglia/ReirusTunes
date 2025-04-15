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

func failOnError(msg string, err error) {
	if err != nil {
		log.Panicf("Error: %s;%s", msg, err)
	}
}

func main() {
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

	failOnError("No se pudo crear la cola", err)
	fmt.Println("Consumiendo...")
	wg := *&sync.WaitGroup{}
	wg.Add(1)
	go ConsumingMessages(&wg, ch, &queue, "")
	wg.Wait()
}

func ConsumingMessages(wg *sync.WaitGroup, ch *amqp.Channel, queue *amqp.Queue, tag string) {
	defer wg.Done()
	msgs, err := ch.Consume(
		queue.Name,
		tag,
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError("No se pudo consumir los mensajes", err)

	for {
		select {
		case msg := <-msgs:
			body_request := make(map[string]string)
			err := json.Unmarshal(msg.Body, &body_request)
			failOnError("No se pudo deserializar", err)
			// testo
			currentTime := time.Now()

			searchQuery := "ytsearch:" + body_request["songname"]

			cmd := exec.Command("yt-dlp",
				"-x",
				"--geo-bypass",
				"--age-limit", "99",
				"--audio-format", "mp3",
				searchQuery,
			)

			go func() {
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Println("Error: en la funcion 'ConsumingMessages', no se pudo leer la salida; ", err)
					fmt.Println(string(output))
					if strings.Contains(string(output), "Sign in to confirm you're not a bot") {
						fmt.Println("Se a detectado un bloqueo por Youtube")
						file_log, err := os.OpenFile("./logs/blocked.json", os.O_CREATE|os.O_APPEND|os.O_EXCL, os.ModeAppend)
						if err != nil {
							fmt.Println("Error: No se pudo abrir 'logs.json'; ", err)
							os.Mkdir("logs", os.ModeAppend)
						}
						file_log.WriteString(fmt.Sprint("Fecha: ", currentTime, "; Blocked by Youtube."))
						// Aqui la funcion de cambio de IP
					}
				}
			}()
			fmt.Printf("Descargando audio para: %s\n", body_request["songname"])
			// fin del testeo
			for key, value := range body_request {
				fmt.Println(key, value, "\n")
			}
			fmt.Println("Original Messsage: ", string(msg.Body))
		}
	}
}
