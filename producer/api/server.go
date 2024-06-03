package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/s4kh/trader-app/producer/msgbroker"
)

// Validator is an object that can be validated.
type Validator interface {
	// Valid checks the object and returns any
	// problems. If len(problems) == 0 then
	// the object is valid.
	Valid(ctx context.Context) (problems map[string]string)
}

type Vote struct {
	CandidateId string `json:"candidateId"`
	PartyId     string `json:"partyId"`
	Count       int    `json:"count"`
}

func (v Vote) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if len(v.CandidateId) == 0 {
		problems["CandidateId"] = "Candidate ID cannot be empty or null"
	}

	if len(v.PartyId) == 0 {
		problems["PartyId"] = "Party ID cannot be empty or null"
	}

	if v.Count == 0 || v.Count > 10000 {
		problems["Count"] = "Count must be in range of 1 and 9999"
	}

	return problems
}

type responseLogger struct {
	http.ResponseWriter
	status int
}

func (rl *responseLogger) WriteHeader(code int) {
	rl.status = code
	rl.ResponseWriter.WriteHeader(code)
}

func Logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &responseLogger{ResponseWriter: w}

		h.ServeHTTP(recorder, r)
		log.Println(recorder.status, r.Method, r.URL.Path, time.Since(start))
	})
}

func NewServer(mb msgbroker.MsgBroker) http.Handler {
	mux := http.NewServeMux()
	// if you have multiple routes you would extract a routes.go
	mux.Handle("POST /vote", Logging(handleVote(mb)))

	return mux
}

func handleVote(mb msgbroker.MsgBroker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v Vote
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			w.WriteHeader(400)
			w.Write([]byte("malformed request body"))
			return
		}

		if problems := v.Valid(r.Context()); len(problems) > 0 {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprint(problems)))
			return
		}

		jsonVote, err := json.Marshal(v)
		if err != nil {
			log.Printf("error converting vote to json: %v", err)
			return
		}

		rc := make(chan msgbroker.PublishRes)
		go listenChan(rc)

		go mb.Publish(string(jsonVote), v.CandidateId, msgbroker.VOTE_RECEIVED, rc)
		w.WriteHeader(http.StatusOK)
	})
}

func listenChan(rc <-chan msgbroker.PublishRes) {
	for val := range rc {
		fmt.Println(val)
	}
}
