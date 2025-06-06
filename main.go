package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Movie struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Rank      int    `json:"rank"`
	PosterURL string `json:"poster_url"`
}

var movies []Movie

func main() {
	if err := loadCSV(); err != nil {
		log.Fatalf("Failed to load CSV: %v", err)
	}
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/game", withCORS(gameHandler))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadCSV() error {
	file, err := os.Open("movies.csv")
	if err != nil {
		return fmt.Errorf("failed to open movies.csv: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %v", err)
	}
	if len(header) != 4 {
		return fmt.Errorf("invalid CSV header: expected 4 columns, got %d", len(header))
	}

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading CSV record: %v", err)
			continue
		}
		if len(record) != 4 {
			log.Printf("Skipping invalid record: %v", record)
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			log.Printf("Skipping record with invalid ID %q: %v", record[0], err)
			continue
		}
		rank, err := strconv.Atoi(record[2])
		if err != nil {
			log.Printf("Skipping record with invalid rank %q: %v", record[2], err)
			continue
		}

		movies = append(movies, Movie{
			ID:        id,
			Title:     record[1],
			Rank:      rank,
			PosterURL: record[3],
		})
	}
	if len(movies) == 0 {
		return fmt.Errorf("no valid movies loaded from CSV")
	}
	log.Printf("Loaded %d movies from CSV", len(movies))
	return nil
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	rand.Shuffle(len(movies), func(i, j int) { movies[i], movies[j] = movies[j], movies[i] })

	type Question struct {
		MovieA Movie  `json:"movieA"`
		MovieB Movie  `json:"movieB"`
		Type   string `json:"type"`
	}

	var questions []Question
	for i := 0; i < 10 && i*2+1 < len(movies); i++ {
		a := movies[i*2]
		b := movies[i*2+1]
		qType := "more"
		if rand.Intn(2) == 0 {
			qType = "less"
		}

		questions = append(questions, Question{
			MovieA: a,
			MovieB: b,
			Type:   qType,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(questions); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		}
		h.ServeHTTP(w, r)
	}
}
