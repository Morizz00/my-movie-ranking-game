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
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Rank     int    `json:"rank"`
	ImageUrl string `json:"image_url"`
}

var movies []Movie

func main() {
	if err := loadCSV(); err != nil {
		log.Fatalf("Failed to load CSV:%v", err)
	}
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/game", withCORS(gameHandler))

	log.Println("server started at port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func loadCSV() error {
	file, err := os.Open("movies.csv")
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
		movies = append(movies, Movie{
			ID:       id,
			Title:    record[1],
			Rank:     rank,
			ImageUrl: record[3],
		})
	}
	if len(movies) == 0 {
		return fmt.Errorf("nil movies loaded fron csv")
	}
	log.Printf("loaded %d movies from csv", len(movies))
	return nil
}
func gameHandler(w http.ResponseWriter, r *http.Request) {
	rand.Shuffle(len(movies), func(i, j int) {
		movies[i], movies[j] = movies[j], movies[i]
	})
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
		http.Error(w, "Failed to encode resp", http.StatusInternalServerError)
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

