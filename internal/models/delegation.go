package models

import (
	"encoding/json"
	"time"
)

type Delegation struct {
	ID        int64     `json:"id"`
	Delegator string    `json:"-"`
	Timestamp time.Time `json:"timestamp"`
	Amount    int64     `json:"amount"`
	Level     int32     `json:"level"`
}

// tzktDelegation is the response structure from TzKT API
type tzktDelegation struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Amount    int64     `json:"amount"`
	Level     int32     `json:"level"`
	Sender    struct {
		Address string `json:"address"`
	} `json:"sender"`
}

// UnmarshalJSON custom unmarshaling to handle nested sender.address
func (d *Delegation) UnmarshalJSON(data []byte) error {
	var t tzktDelegation
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	d.ID = t.ID
	d.Delegator = t.Sender.Address
	d.Timestamp = t.Timestamp
	d.Amount = t.Amount
	d.Level = t.Level

	return nil
}
