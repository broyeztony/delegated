package models

import "time"

type Delegation struct {
	ID        int64     `json:"id"`
	Delegator string    `json:"delegator"`
	Timestamp time.Time `json:"timestamp"`
	Amount    int64     `json:"amount"`
	Level     int32     `json:"level"`
}
