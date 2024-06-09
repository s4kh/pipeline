package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/s4kh/backend/api"
	"github.com/s4kh/backend/db"
	"github.com/s4kh/backend/msgbroker"
)

func run() error {
	err := godotenv.Load("./.env")
	if err != nil {
		return fmt.Errorf("could not load .env: %v", err)
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	fmt.Println(connStr)
	db, err := db.NewPostgresConnection(connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	brokers := []string{fmt.Sprintf("%s:%s", os.Getenv("KAFKA_HOST"), os.Getenv("KAFKA_PORT"))}
	consumer := msgbroker.NewReader(brokers, msgbroker.VOTE_RECEIVED, msgbroker.VOTE_GROUP)
	wss := api.NewHub()

	srv := api.NewServer(db, consumer, wss)
	httpServer := http.Server{
		Addr:    ":8082",
		Handler: srv,
	}

	log.Printf("listening on %s\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error listening and serving: %s", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("backend - you are a noob:%v", err)
	}
}
