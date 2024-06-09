package models

type Candidate struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Region    int    `json:"region"`
	PartyId   string `json:"partyId"`
	Gender    string `json:"gender"`
	ListOrder int    `json:"listOrder"`
}
