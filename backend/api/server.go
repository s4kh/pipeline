package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/s4kh/app/db"
	"github.com/s4kh/app/lib"
)

type Vote struct {
	CandidateId string `json:"candidateId"`
	PartyId     string `json:"partyId"`
	Count       int    `json:"count"`
}

func handleGetVotes(db db.Conn) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read from db
	})
}

// func handleWebSocket() http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		conn, err:=
// 	})
// }

func consume(c lib.Broker, db db.Conn) {
	for {
		msg, err := c.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error during msg consumption: %v", err)
			continue
		}

		go func(msg []byte) {
			err := handleMessage(msg, db)
			if err != nil {
				log.Printf("failed to handle message: %v, err: %v\n", string(msg), err)
			}
		}(msg)
	}
}

func handleMessage(msg []byte, db db.Conn) error {
	var v Vote
	if err := json.Unmarshal(msg, &v); err != nil {
		return fmt.Errorf("could not unmarshal msg: %v", err)
	}

	// write to db
	_, err := db.Get().Exec("INSERT INTO votes (candidate_id, party_id, total_vote) VALUES ($1, $2, $3) ON CONFLICT (candidate_id) DO UPDATE SET total_vote = votes.\"total_vote\" + $3, updated_at = $4",
		v.CandidateId, v.PartyId, v.Count, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update the vote in db: %v", err)
	}

	log.Println("processed", v)

	return nil
}

func NewServer(db db.Conn, consumer lib.Broker) http.Handler {
	mux := http.NewServeMux()
	// if you have multiple routes you would extract a routes.go
	mux.Handle("GET /votes", handleGetVotes(db))
	mux.HandleFunc("/", http.NotFoundHandler().ServeHTTP)

	go consume(consumer, db)

	return mux
}
