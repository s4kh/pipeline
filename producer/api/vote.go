package api

import "context"

// Validator is an object that can be validated.
type Validator interface {
	// Valid checks the object and returns any
	// problems. If len(problems) == 0 then
	// the object is valid.
	Valid(ctx context.Context) (problems map[string]string)
}

type Vote struct {
	CandidateId string `json:"candidateId,omitempty"`
	PartyId     string `json:"partyId"`
	Count       int    `json:"count"`
	Type        string `json:"type"`
}

const (
	TYPE_CANDIDATE = "candidate"
	TYPE_PARTY     = "party"
)

func (v Vote) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if len(v.Type) == 0 || (v.Type != TYPE_CANDIDATE && v.Type != TYPE_PARTY) {
		problems["Type"] = "Invalid vote type"
	}

	if len(v.CandidateId) == 0 && v.Type == TYPE_CANDIDATE {
		problems["CandidateId"] = "Candidate ID cannot be empty or null for candidate vote"
	}

	if len(v.PartyId) == 0 {
		problems["PartyId"] = "Party ID cannot be empty or null"
	}

	if v.Count == 0 || v.Count > 10000 {
		problems["Count"] = "Count must be in range of 1 and 9999"
	}

	return problems
}
