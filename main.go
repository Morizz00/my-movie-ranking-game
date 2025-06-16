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
	"strings"
	"time"
)

type Item struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Rank     int    `json:"rank"`
	ImageUrl string `json:"image_url"`
	Type     string `json:"type"`
}

type Question struct {
	ItemA Item   `json:"itemA"`
	ItemB Item   `json:"itemB"`
	Type  string `json:"type"`
}

var movies []Item
var tvShows []Item

func main() {
	if err := loadCSV("movies.csv", &movies, "movie"); err != nil {
		log.Fatalf("failed to load movies CSV: %v", err)
	}

	if err := loadCSV("tvshows.csv", &tvShows, "tv"); err != nil {
		log.Printf("failed to load TV shows CSV: %v", err)
		log.Println("TV shows mode not working")
	}

	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/game", withCORS(gameHandler))
	http.HandleFunc("/game/movies", withCORS(movieGameHandler))
	http.HandleFunc("/game/tv", withCORS(tvGameHandler))

	log.Println("Server started at port :8080")
	log.Printf("Loaded %d movies", len(movies))
	log.Printf("Loaded %d TV shows", len(tvShows))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadCSV(filename string, items *[]Item, itemType string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open cs file:%v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("cannot read csv header:%v", err)
	}
	if len(header) != 4 {
		return fmt.Errorf("invalid number of arguements:%d", len(header))
	}
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading CSV:%v", err)
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			log.Printf("skip record with invalid id %q:%v", record[0], err)
			continue
		}
		rank, err := strconv.Atoi(record[2])
		if err != nil {
			log.Printf("skip record with invalid rank %q:%v", record[2], err)
			continue
		}
		*items = append(*items, Item{
			ID:       id,
			Title:    record[1],
			Rank:     rank,
			ImageUrl: record[3],
			Type:     itemType,
		})
	}
	if len(*items) == 0 {
		return fmt.Errorf("nil movies loaded fron csv")
	}
	log.Printf("loaded %d movies from csv", len(*items))
	return nil
}

func generateQuestions(items []Item, count int) []Question {
	if len(items) < 2 {
		return []Question{}
	}

	itemsCopy := make([]Item, len(items))
	copy(itemsCopy, items)
	rand.Shuffle(len(itemsCopy), func(i, j int) {
		itemsCopy[i], itemsCopy[j] = itemsCopy[j], itemsCopy[i]
	})

	var questions []Question
	maxQuestions := min(count, len(itemsCopy)/2)

	for i := 0; i < maxQuestions; i++ {
		a := itemsCopy[i*2]
		b := itemsCopy[i*2+1]

		qType := "more"
		if rand.Intn(2) == 0 {
			qType = "less"
		}

		questions = append(questions, Question{
			ItemA: a,
			ItemB: b,
			Type:  qType,
		})
	}

	return questions
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")

	var questions []Question

	switch strings.ToLower(mode) {
	case "tv", "tvshows":
		if len(tvShows) == 0 {
			http.Error(w, "TV shows not available", http.StatusNotFound)
			return
		}
		questions = generateQuestions(tvShows, 10)
	case "mixed", "both":
		if len(movies) == 0 && len(tvShows) == 0 {
			http.Error(w, "No content available", http.StatusNotFound)
			return
		}
		var combined []Item
		combined = append(combined, movies...)
		combined = append(combined, tvShows...)
		questions = generateQuestions(combined, 10)
	case "movies", "":
		if len(movies) == 0 {
			http.Error(w, "Movies not available", http.StatusNotFound)
			return
		}
		questions = generateQuestions(movies, 10)
	default:
		http.Error(w, "Invalid mode. Use 'movies', 'tv', or 'mixed'", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(questions); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func movieGameHandler(w http.ResponseWriter, r *http.Request) {
	if len(movies) == 0 {
		http.Error(w, "Movies not available", http.StatusNotFound)
		return
	}

	questions := generateQuestions(movies, 10)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(questions); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func tvGameHandler(w http.ResponseWriter, r *http.Request) {
	if len(tvShows) == 0 {
		http.Error(w, "TV shows not available", http.StatusNotFound)
		return
	}

	questions := generateQuestions(tvShows, 10)

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
