package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

func main() {
	// Flags para parametrizar concurrencia y URL fija
	numUsers := flag.Int("users", 5, "Número de goroutines concurrentes (usuarios)")
	downloads := flag.Int("downloads", 10, "Descargas por usuario")
	memInterval := flag.Duration("mem-interval", time.Second, "Intervalo para mostrar stats de memoria")
	flag.Parse()

	// URL fija de destino
	const url = "http://localhost:8080/songs"

	// Lista de 25 canciones
	songs := []string{
		"Bohemian Rhapsody", "Stairway to Heaven", "Smells Like Teen Spirit",
		"Hotel California", "Billie Jean", "Sweet Child O' Mine",
		"Imagine", "Hey Jude", "Wonderwall", "Thriller",
		"Like a Rolling Stone", "Lose Yourself", "Rolling in the Deep",
		"Yesterday", "Shape of You", "Blinding Lights",
		"Despacito", "Uptown Funk", "Thinking Out Loud",
		"Take On Me", "Bad Guy", "Shallow",
		"Radioactive", "Viva La Vida", "Hips Don't Lie",
	}

	// Preparamos canal de tareas
	totalRequests := (*numUsers) * (*downloads)
	tasks := make(chan string, totalRequests)
	for i := 0; i < totalRequests; i++ {
		tasks <- songs[i%len(songs)]
	}
	close(tasks)

	// Canal para detener monitor de memoria al finalizar
	done := make(chan struct{})

	// Monitor de memoria hasta que cierren todas las descargas
	go func() {
		ticker := time.NewTicker(*memInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				fmt.Printf("🖥️ Memoria: Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB, NumGC = %v\n",
					bToMiB(m.Alloc), bToMiB(m.TotalAlloc), bToMiB(m.Sys), m.NumGC)
			case <-done:
				return
			}
		}
	}()

	// Workers concurrentes
	var wg sync.WaitGroup
	wg.Add(*numUsers)
	for i := 0; i < *numUsers; i++ {
		go func(id int) {
			defer wg.Done()
			client := &http.Client{}
			for song := range tasks {
				payload := map[string]string{"songname": song}
				jsonData, err := json.Marshal(payload)
				if err != nil {
					fmt.Printf("Worker %d ❌ Error JSON %s: %v\n", id, song, err)
					continue
				}

				resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					fmt.Printf("Worker %d ❌ POST %s: %v\n", id, song, err)
					continue
				}
				fmt.Printf("Worker %d ✅ POST %s => %s\n", id, song, resp.Status)
				resp.Body.Close()
			}
		}(i + 1)
	}

	// Esperamos a que terminen y detenemos monitor
	wg.Wait()
	close(done)
}

// bToMiB convierte bytes a MiB
func bToMiB(b uint64) uint64 {
	return b / 1024 / 1024
}
