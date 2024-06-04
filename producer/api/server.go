package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/s4kh/trader-app/producer/msgbroker"
)

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
	mux.HandleFunc("/", http.NotFoundHandler().ServeHTTP)

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
	log.Println("Listening for result channel")
	for val := range rc {
		if val.Code == 1 {
			log.Printf("listenChan: %v\n", val.Err)
		}
	}
}
