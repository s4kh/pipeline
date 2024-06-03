package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/s4kh/trader-app/producer/api"
	"github.com/s4kh/trader-app/producer/msgbroker"
)

func run() error {
	err := godotenv.Load("./.env")

	if err != nil {
		return fmt.Errorf("could not load .env: %v", err)
	}

	mb := msgbroker.NewMsgBrokerClient(os.Getenv("KAFKA_HOST"), os.Getenv("KAFKA_PORT"))
	defer mb.Writer.Close()

	// t := os.Getenv("TICKERS")
	// topics := strings.Split(t, ",")

	// err = trades.SubscribeAndListen(topics, mb)
	// if err != nil {
	// 	return err
	// }

	// defer trades.CloseConnections()

	// server section
	srv := api.NewServer(mb)
	httpServer := http.Server{
		Addr:    ":8081",
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
		fmt.Fprintf(os.Stderr, "you are a noob: %v\n", err)
		os.Exit(1)
	}
}
