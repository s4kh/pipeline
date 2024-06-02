package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/s4kh/trader-app/producer/trades"
)

func run() error {
	err := godotenv.Load("./.env")

	if err != nil {
		return fmt.Errorf("could not load .env: %v", err)
	}

	mb := trades.NewMsgBrokerClient(os.Getenv("KAFKA_HOST"), os.Getenv("KAFKA_PORT"))

	t := os.Getenv("TICKERS")
	topics := strings.Split(t, ",")

	err = trades.SubscribeAndListen(topics, mb)
	if err != nil {
		return err
	}

	trades.CloseConnections()
	mb.Writer.Close()

	return nil
}

func main() {
	fmt.Println("Hello from producer")
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "you are a noob: %v\n", err)
		os.Exit(1)
	}
}
