package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type RequestData struct {
	Songs []string `json:"songs"`
}

// Sanitiza el nombre del archivo quitando caracteres inválidos
func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(" ", "_", "/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return replacer.Replace(name)
}

// Descarga el audio y lo guarda en ./downloads
func downloadAudio(query string) (string, error) {
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%06x", rand.Intn(0xffffff)) // ID hexadecimal corto

	fileName := fmt.Sprintf("%s_%s.mp3", sanitizeFileName(query), id)
	outputPath := filepath.Join("downloads", fileName)

	cmd := exec.Command("yt-dlp",
		"-x",
		"--geo-bypass",
		"--audio-format", "mp3",
		"-o", outputPath,
		"ytsearch1:"+query,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Descargando: %s => %s\n", query, fileName)
	err := cmd.Run()

	return fileName, err
}

func downloadHandler(c *gin.Context) {
	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato JSON inválido"})
		return
	}

	os.MkdirAll("downloads", os.ModePerm)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 50) // Hasta 3 descargas a la vez

	var mu sync.Mutex
	var downloadedFiles []string

	for _, song := range requestData.Songs {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(q string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			file, err := downloadAudio(q)
			if err != nil {
				fmt.Printf("Error con %s: %v\n", q, err)
				return
			}

			mu.Lock()
			downloadedFiles = append(downloadedFiles, file)
			mu.Unlock()
		}(song)
	}

	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"message": "Todas las descargas han finalizado",
		"files":   downloadedFiles,
	})
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	router.Static("/downloads", "./downloads") // Exponer archivos

	router.POST("/songs", downloadHandler)

	router.Run(":8080")
}
