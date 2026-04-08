package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func NewConnection() *sql.DB {
	connStr := "postgres://postgres:postgres@localhost:5432/sofreserve?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}