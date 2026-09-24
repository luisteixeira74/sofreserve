package db

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"sof-reserve/internal/config"

	_ "github.com/lib/pq"
)

func NewConnection(cfg config.Config) *sql.DB {
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort),
		Path:   cfg.DBName,
	}

	u.User = url.UserPassword(cfg.DBUser, cfg.DBPassword)
	q := u.Query()
	q.Set("sslmode", cfg.DBSSLMode)
	u.RawQuery = q.Encode()

	connStr := u.String()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}
