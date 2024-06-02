package trades

import "fmt"

type Ticker struct {
	Symbol   string `json:"s"`
	Price    string `json:"p"`
	Quantity string `json:"q"`
	Time     int64  `json:"T"` //timestamp
}

func (t Ticker) String() string {
	return fmt.Sprintf("%s:%s - %s", t.Symbol, t.Price, t.Quantity)
}
