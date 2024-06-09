package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/s4kh/backend/db"
	"github.com/s4kh/backend/models"
	"github.com/s4kh/backend/msgbroker"
)

func handleGetVotes(db db.DB) http.Handler {
	const (
		maxPageSize     = 100
		defaultPageSize = 50
		defaultPage     = 1
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()

		page, err := strconv.Atoi(queryParams.Get("page"))
		if err != nil || page < 1 {
			page = defaultPage
		}

		pageSize, err := strconv.Atoi(queryParams.Get("pageSize"))
		if err != nil || pageSize < 1 || pageSize > maxPageSize {
			pageSize = defaultPageSize
		}

		cVotes, err := db.FetchCandidateVotes(r.Context(), page, pageSize)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to fetch votes"))
			return
		}

		pVotes, err := db.FetchPartyVotes(r.Context(), page, pageSize)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to fetch votes"))
			return
		}
		log.Println("=========", pVotes)

		allVotes := &models.AllVotes{CandidateVotes: cVotes, PartyVotes: pVotes}

		err = encode(w, 200, allVotes)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to fetch votes"))
		}

	})
}

// func handleWebSocket() http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		conn, err:=
// 	})
// }

func consume(br msgbroker.BrokerReader, db db.DB, b models.VoteUpdateBroadCaster) {
	for {
		msg, err := br.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error during msg consumption: %v", err)
			continue
		}

		go func(msg []byte) {
			err := handleMessage(msg, db, b)
			if err != nil {
				log.Printf("failed to handle message: %v, err: %v\n", string(msg), err)
			}
		}(msg)
	}
}

func handleMessage(msg []byte, db db.DB, b models.VoteUpdateBroadCaster) error {
	var v models.Vote
	if err := json.Unmarshal(msg, &v); err != nil {
		return fmt.Errorf("could not unmarshal msg: %v", err)
	}

	// write to db

	if err := db.UpsertVoteEvent(context.Background(), v); err != nil {
		return fmt.Errorf("failed to update the vote in db: %v", err)
	}

	b.BroadcastVoteUpdate(&v)

	log.Println("processed", v)

	return nil
}

func NewServer(db db.DB, br msgbroker.BrokerReader, wss *Hub) http.Handler {
	mux := http.NewServeMux()
	// if you have multiple routes you would extract a routes.go
	mux.Handle("GET /votes", handleGetVotes(db))
	mux.Handle("/ws", wss.HandleWebSocket())
	mux.HandleFunc("/", http.NotFoundHandler().ServeHTTP)

	go consume(br, db, wss)
	go wss.StartBroadCast()

	return mux
}
