package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Connect to the database
	var err error
	db, err = sql.Open("postgres", "postgres://root:scribe@localhost/scribe?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Successfully connected to the database")

	for {
		books, err := getAvailableBooks()
		if err != nil {
			log.Printf("Failed to get available books: %v", err)
			return
		}

		fmt.Println("\nBible NASB CLI")
		for i, book := range books {
			fmt.Printf("%d. %s\n", i+1, book)
		}
		fmt.Printf("%d. Exit\n", len(books)+1)
		fmt.Print("Choose a book: ")

		var choice int
		fmt.Scan(&choice)

		if choice == len(books)+1 {
			fmt.Println("Goodbye!")
			return
		}

		if choice < 1 || choice > len(books) {
			fmt.Println("Invalid choice, please try again.")
			continue
		}

		displayBook(books[choice-1])
	}
}

func getAvailableBooks() ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT book FROM verses ORDER BY book")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []string
	for rows.Next() {
		var book string
		if err := rows.Scan(&book); err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, rows.Err()
}

func displayBook(book string) {
	fmt.Printf("\n%s\n", book)

	rows, err := db.Query("SELECT chapter, verse, text FROM verses WHERE book = $1 ORDER BY chapter, verse", book)
	if err != nil {
		log.Printf("Failed to query verses: %v", err)
		return
	}
	defer rows.Close()

	var currentChapter int
	for rows.Next() {
		var chapter int
		var verse int
		var text string
		err := rows.Scan(&chapter, &verse, &text)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		if chapter != currentChapter {
			fmt.Printf("\nChapter %d\n", chapter)
			currentChapter = chapter
		}

		fmt.Printf("%d%s %s\n", verse, superscript(fmt.Sprintf("%d", verse)), text)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}

	fmt.Println()
}

func superscript(n string) string {
	superscriptMap := map[rune]string{
		'1': "¹", '2': "²", '3': "³", '4': "⁴", '5': "⁵",
		'6': "⁶", '7': "⁷", '8': "⁸", '9': "⁹", '0': "⁰",
	}
	var superscripted strings.Builder
	for _, char := range n {
		superscripted.WriteString(superscriptMap[char])
	}
	return superscripted.String()
}
