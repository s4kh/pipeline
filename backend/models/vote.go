package models

import "time"

type Vote struct {
	CandidateId string    `json:"candidateId,omitempty"`
	PartyId     string    `json:"partyId"`
	Count       int       `json:"count"`
	Type        string    `json:"type,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

type VoteUpdateBroadCaster interface {
	BroadcastVoteUpdate(v *Vote)
}

type AllVotes struct {
	CandidateVotes []CandidateVote `json:"candidateVotes"`
	PartyVotes     []Vote          `json:"partyVotes"`
}

type CandidateVote struct {
	Vote
	CandidateName string `json:"candidateName"`
}
