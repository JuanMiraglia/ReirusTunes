package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors" // Importa el paquete CORS
	"github.com/gin-gonic/gin"
)

// Estructura para recibir los datos desde el frontend
type RequestData struct {
	Songs []string `json:"songs"`
}

// downloadAudio ejecuta yt-dlp para descargar el audio en formato mp3
func downloadAudio(query string) error {
	currentTime := time.Now()

	searchQuery := "ytsearch:" + query

	cmd := exec.Command("yt-dlp",
		"-x",
		"--geo-bypass",
		"--age-limit", "99",
		"--audio-format", "mp3",
		searchQuery,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Error: en la funcion 'downloadAudio', no se pudo leer la salida; ", err)
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
	}(wg)
	wg.Wait()
	fmt.Printf("Descargando audio para: %s\n", query)
	return cmd.Run()
}

// Handler para recibir los nombres de canciones y procesarlas
func downloadHandler(c *gin.Context) {
	var requestData RequestData

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato JSON inválido"})
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 4) // Máximo 4 descargas concurrentes

	for _, song := range requestData.Songs {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(q string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := downloadAudio(q); err != nil {
				fmt.Printf("Error al descargar \"%s\": %v\n", q, err)
			} else {
				fmt.Printf("Descarga completada para: %s\n", q)
			}
		}(song)
	}

	wg.Wait()
	c.JSON(http.StatusOK, gin.H{"message": "Todas las descargas han finalizado"})
}

func main() {
	router := gin.Default()
	router.Static("/static", "./templates/static")
	router.LoadHTMLFiles("./templates/index.html")

	// Configurar CORS para permitir solicitudes desde cualquier origen
	router.Use(cors.Default()) // Habilita CORS

	// Definir la ruta para manejar la descarga de canciones
	router.POST("/songs", downloadHandler)

	router.GET("/home", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// Ejecutar el servidor
	router.Run(":8080")
}
