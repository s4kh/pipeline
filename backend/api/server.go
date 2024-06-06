package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/s4kh/app/db"
	"github.com/s4kh/app/msgbroker"
)

type Vote struct {
	CandidateId string `json:"candidateId"`
	PartyId     string `json:"partyId"`
	Count       int    `json:"count"`
}

type VoteUpdateBroadCaster interface {
	BroadcastVoteUpdate(v *Vote)
}

type VoteRes struct {
	Vote
	Timestamp     time.Time `json:"timestamp"`
	CandidateName string    `json:"candidateName"`
	PartyName     string    `json:"partyName"`
}

func fetchVotes(db db.Conn, page, pageSize int) ([]VoteRes, error) {
	offset := (page - 1) * pageSize
	voteRows, err := db.Get().Query("SELECT * FROM votes ORDER BY total_vote DESC LIMIT $1 OFFSET $2", pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve votes: %v", err)
	}

	var votes []VoteRes

	for voteRows.Next() {
		var v VoteRes
		var candName, partyName sql.NullString
		if err := voteRows.Scan(&v.CandidateId, &candName, &v.PartyId, &partyName, &v.Count, &v.Timestamp); err != nil {
			return nil, err
		}
		v.PartyName = partyName.String
		v.CandidateName = candName.String

		votes = append(votes, v)
	}

	return votes, err
}

func handleGetVotes(db db.Conn) http.Handler {
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

		votes, err := fetchVotes(db, page, pageSize)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to fetch votes"))
			return
		}

		err = encode(w, 200, votes)
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

func consume(br msgbroker.BrokerReader, db db.Conn, b VoteUpdateBroadCaster) {
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

func handleMessage(msg []byte, db db.Conn, b VoteUpdateBroadCaster) error {
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

	b.BroadcastVoteUpdate(&v)

	log.Println("processed", v)

	return nil
}

func NewServer(db db.Conn, br msgbroker.BrokerReader, wss *Hub) http.Handler {
	mux := http.NewServeMux()
	// if you have multiple routes you would extract a routes.go
	mux.Handle("GET /votes", handleGetVotes(db))
	mux.Handle("/ws", wss.HandleWebSocket())
	mux.HandleFunc("/", http.NotFoundHandler().ServeHTTP)

	go consume(br, db, wss)
	go wss.StartBroadCast()

	return mux
}
