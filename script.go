package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB
var insertStmt *sql.Stmt

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to the database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database")
}

func main() {
	if err := verifyDBConnection(); err != nil {
		log.Fatal(err)
	}

	if err := checkDatabasePermissions(); err != nil {
		log.Fatal(err)
	}

	if err := prepareInsertStatement(); err != nil {
		log.Fatal(err)
	}

	file, err := os.Open("bible_text.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	verseRegex := regexp.MustCompile(`^(\w+)\s+(\d+):(\d+)\s+(.+)`)

	lineCount := 0
	verseCount := 0

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()

		if line == "" {
			continue
		}

		if !strings.Contains(line, ":") {
			continue
		}

		matches := verseRegex.FindStringSubmatch(line)
		if len(matches) == 5 {
			book := matches[1]
			chapter, _ := strconv.Atoi(matches[2])
			verse, _ := strconv.Atoi(matches[3])
			text := matches[4]

			insertVerse(book, chapter, verse, text)
			verseCount++
		}

		if lineCount%1000 == 0 {
			log.Printf("Processed %d lines, inserted %d verses", lineCount, verseCount)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Finished processing. Total lines: %d, Total verses inserted: %d", lineCount, verseCount)
}

func verifyDBConnection() error {
	err := db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Check if the table exists
	var tableExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'bible_verses')").Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %v", err)
	}

	// If the table doesn't exist, create it
	if !tableExists {
		_, err = db.Exec(`
			CREATE TABLE bible_verses (
				id SERIAL PRIMARY KEY,
				bible_version VARCHAR(50) NOT NULL,
				book VARCHAR(50) NOT NULL,
				chapter INT NOT NULL,
				verse INT NOT NULL,
				text TEXT NOT NULL
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
		fmt.Println("Created bible_verses table")
	} else {
		fmt.Println("bible_verses table already exists")
	}

	fmt.Println("Database connection and table verified successfully")
	return nil
}

func checkDatabasePermissions() error {
	// Try to create a temporary table
	_, err := db.Exec(`
		CREATE TEMPORARY TABLE temp_test (
			id SERIAL PRIMARY KEY,
			name TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create temporary table: %v", err)
	}

	// Try to insert into the temporary table
	_, err = db.Exec(`INSERT INTO temp_test (name) VALUES ('test')`)
	if err != nil {
		return fmt.Errorf("failed to insert into temporary table: %v", err)
	}

	fmt.Println("Database permissions verified successfully")
	return nil
}

func prepareInsertStatement() error {
	var err error
	insertStmt, err = db.Prepare(`
		INSERT INTO bible_verses (bible_version, book, chapter, verse, text)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %v", err)
	}
	return nil
}

func insertVerse(book string, chapter, verse int, content string) {
	_, err := insertStmt.Exec("KJV", book, chapter, verse, content)
	if err != nil {
		log.Printf("Error inserting verse: %v", err)
		log.Printf("Failed to insert: Book: %s, Chapter: %d, Verse: %d", book, chapter, verse)
	} else {
		fmt.Printf("Inserted %s %d:%d\n", book, chapter, verse)
	}
}
