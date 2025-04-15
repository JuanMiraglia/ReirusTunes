package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	// Lista de 25 canciones reales
	songs := []string{
		"Bohemian Rhapsody",
		"Stairway to Heaven",
		"Smells Like Teen Spirit",
		"Hotel California",
		"Billie Jean",
		"Sweet Child O' Mine",
		"Imagine",
		"Hey Jude",
		"Wonderwall",
		"Thriller",
		"Like a Rolling Stone",
		"Lose Yourself",
		"Rolling in the Deep",
		"Yesterday",
		"Shape of You",
		"Blinding Lights",
		"Despacito",
		"Uptown Funk",
		"Thinking Out Loud",
		"Take On Me",
		"Bad Guy",
		"Shallow",
		"Radioactive",
		"Viva La Vida",
		"Hips Don't Lie",
	}

	url := "http://localhost:8080/songs"

	for _, song := range songs {
		data := map[string]string{"songname": song}
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("❌ Error codificando JSON para %s: %v\n", song, err)
			continue
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("❌ Error haciendo POST para %s: %v\n", song, err)
			continue
		}

		fmt.Printf("✅ POST %s => Status: %s\n", song, resp.Status)
		resp.Body.Close()
	}
}
