package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func goDotEnvVariable(key string) string {
	return os.Getenv(key)
}

func dsn() string {
	username := "root"
	password := ""
	hostname := "localhost"
	dbName := "eventocerto"
	err := godotenv.Load()

	if err == nil {
		username = os.Getenv("DATABASE_USERNAME")
		password = os.Getenv("DATABASE_PASSWORD")
		hostname = os.Getenv("DATABASE_HOST")
	}

	return fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbName)
}

func dbConnection() (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn())
	if err != nil {
		log.Printf("Error %s when opening DB", err)
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.PingContext(ctx)
	if err != nil {
		log.Printf("Errors %s pinging DB", err)
		return nil, err
	}
	log.Printf("Connected to DB successfully\n")
	return db, nil
}

func resetWebsiteEvents(db *sql.DB, website string) error {
	query := "DELETE FROM `events` WHERE website = (?)"
	_, err := db.ExecContext(context.Background(), query, website)
	if err != nil {
		log.Fatalf("impossible to reset events: %s", err)
	}
	return nil
}

func insertEvent(db *sql.DB, event Event) error {
	uuid := uuid.New()

	idStr := uuid.String()
	query := "INSERT INTO `events` (`id`, `title`, `date`, `thumbnail`, `location`, `genre`, `website`) VALUES (?, ?, ?, ?, ?, ?, ?)"
	insertResult, err := db.ExecContext(context.Background(), query, idStr, event.Title, event.Date, event.Thumbnail, event.Location, event.Genre, event.Website)
	if err != nil {
		log.Fatalf("impossible insert event: %s", err)
	}
	_, err = insertResult.LastInsertId()
	if err != nil {
		log.Fatalf("impossible to retrieve last inserted id: %s", err)
	}
	return nil
}
