package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

var db *sql.DB

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	var err error
	db, err = sql.Open("postgres", connStr) // Initialize the variable here
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully connected to the database")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/verse", handleVerseRequest)
	mux.HandleFunc("/api/books", handleBooksRequest)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		Debug:          true, // Enable CORS logging
	})

	handler := c.Handler(mux)
	apiPort := os.Getenv("API_PORT")
	fmt.Printf("API server is running on http://0.0.0.0:%s and http://[::]:%s\n", apiPort, apiPort)

	log.Fatal(http.ListenAndServe("0.0.0.0:"+apiPort, handler))
}

func handleVerseRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received verse request from %s: %s", r.RemoteAddr, r.URL.String())
	version := r.URL.Query().Get("version")
	book := r.URL.Query().Get("book")
	chapter := r.URL.Query().Get("chapter")
	verse := r.URL.Query().Get("verse")
	log.Printf("Params: version=%s, book=%s, chapter=%s, verse=%s", version, book, chapter, verse)

	if version == "" || book == "" || chapter == "" || verse == "" {
		log.Println("Missing required parameters")
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	var text string
	err := db.QueryRow("SELECT text FROM bible_verses WHERE bible_version = $1 AND book = $2 AND chapter = $3 AND verse = $4",
		version, book, chapter, verse).Scan(&text)
	if err == sql.ErrNoRows {
		log.Printf("Verse not found: %s %s:%s", book, chapter, verse)
		http.Error(w, "Verse not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Version string `json:"version"`
		Book    string `json:"book"`
		Chapter string `json:"chapter"`
		Verse   string `json:"verse"`
		Text    string `json:"text"`
	}{
		Version: version,
		Book:    book,
		Chapter: chapter,
		Verse:   verse,
		Text:    text,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	} else {
		log.Printf("Successfully sent response for %s %s:%s", book, chapter, verse)
	}
}

func handleBooksRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received books request from %s", r.RemoteAddr)
	rows, err := db.Query("SELECT DISTINCT book FROM bible_verses ORDER BY book")
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var books []string
	for rows.Next() {
		var book string
		if err := rows.Scan(&book); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		books = append(books, book)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(books); err != nil {
		log.Printf("Error encoding books response: %v", err)
	} else {
		log.Printf("Successfully sent response with %d books", len(books))
	}
}
