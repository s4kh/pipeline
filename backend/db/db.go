package db

import (
	"context"

	"github.com/s4kh/backend/models"
)

type DB interface {
	FetchCandidateVotes(ctx context.Context, page, pageSize int) ([]models.CandidateVote, error)
	FetchPartyVotes(ctx context.Context, page, pageSize int) ([]models.Vote, error)
	UpsertVoteEvent(ctx context.Context, v models.Vote) error
	Close() error
}
