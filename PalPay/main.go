package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Username string  `json:"username"`
	Balance  float64 `json:"balance"`
}

func main() {
	db, err := sql.Open("sqlite3", "/app/palpay.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	setupRoutes(db)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
