package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/s4kh/backend/models"
)

type PostgresDB struct {
	Conn *sql.DB
}

func NewPostgresConnection(url string) (DB, error) {
	conn, err := sql.Open("postgres", url)

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("error connecting to DB: %s, %v", url, err)
	}

	log.Println(conn.Stats().InUse)

	return &PostgresDB{Conn: conn}, err
}

func (db *PostgresDB) Close() error {
	if db.Conn != nil {
		return db.Conn.Close()
	}

	return fmt.Errorf("cannot close nil db")
}

func (db *PostgresDB) UpsertVoteEvent(ctx context.Context, v models.Vote) error {
	if v.Type == "candidate" {
		res, err := db.Conn.ExecContext(ctx, "UPDATE candidate_votes SET total_vote = candidate_votes.\"total_vote\" + $2, updated_at = $3  WHERE candidate_id = $1",
			v.CandidateId, v.Count, time.Now(),
		)

		if err != nil {
			return fmt.Errorf("upsert error:%v", err)
		}

		// insert
		if count, _ := res.RowsAffected(); count == 0 {
			_, err := db.Conn.ExecContext(ctx, "INSERT INTO candidate_votes (candidate_id, party_id, total_vote) VALUES ($1, $2, $3)",
				v.CandidateId, v.PartyId, v.Count,
			)

			if err != nil {
				return fmt.Errorf("failed to update the candidate vote in db: %v", err)
			}
		}

		return nil
	}
	res, err := db.Conn.ExecContext(ctx, "UPDATE party_votes SET total_vote = party_votes.\"total_vote\" + $2, updated_at = $3  WHERE party_id = $1",
		v.PartyId, v.Count, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("upsert error:%v", err)
	}

	// insert
	if count, _ := res.RowsAffected(); count == 0 {
		_, err := db.Conn.ExecContext(ctx, "INSERT INTO party_votes (party_id, total_vote) VALUES ($1, $2)",
			v.PartyId, v.Count,
		)

		if err != nil {
			return fmt.Errorf("failed to update the party vote in db: %v", err)
		}
	}

	return nil
}

func (db *PostgresDB) FetchCandidateVotes(ctx context.Context, page, pageSize int) ([]models.CandidateVote, error) {
	offset := (page - 1) * pageSize
	voteRows, err := db.Conn.QueryContext(ctx, "SELECT * FROM candidate_votes ORDER BY total_vote DESC LIMIT $1 OFFSET $2", pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve votes: %v", err)
	}

	var votes []models.CandidateVote

	for voteRows.Next() {
		var v models.CandidateVote
		var candName sql.NullString
		if err := voteRows.Scan(&v.CandidateId, &candName, &v.PartyId, &v.Count, &v.Timestamp); err != nil {
			return nil, err
		}
		v.CandidateName = candName.String

		votes = append(votes, v)
	}

	return votes, err
}

func (db *PostgresDB) FetchPartyVotes(ctx context.Context, page, pageSize int) ([]models.Vote, error) {
	offset := (page - 1) * pageSize
	voteRows, err := db.Conn.QueryContext(ctx, "SELECT * FROM party_votes ORDER BY total_vote DESC LIMIT $1 OFFSET $2", pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve party votes: %v", err)
	}

	votes := make([]models.Vote, 0)
	log.Println("232332", votes)

	for voteRows.Next() {
		var v models.Vote
		if err := voteRows.Scan(&v.PartyId, &v.Count, &v.Timestamp); err != nil {
			return nil, err
		}

		votes = append(votes, v)
	}

	return votes, err
}
