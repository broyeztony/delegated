package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDelegation_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Delegation
		wantErr bool
	}{
		{
			name: "valid delegation with nested sender",
			json: `{
				"id": 123,
				"level": 456,
				"timestamp": "2022-05-05T06:29:14Z",
				"amount": 98765,
				"sender": {
					"address": "tz1a1SAaXRt9yoGMx29rh9FsBF4UzmvojdTL"
				}
			}`,
			want: Delegation{
				ID:        123,
				Delegator: "tz1a1SAaXRt9yoGMx29rh9FsBF4UzmvojdTL",
				Timestamp: time.Date(2022, 5, 5, 6, 29, 14, 0, time.UTC),
				Amount:    98765,
				Level:     456,
			},
			wantErr: false,
		},
		{
			name: "missing sender address field",
			json: `{
				"id": 123,
				"level": 456,
				"timestamp": "2022-05-05T06:29:14Z",
				"amount": 98765,
				"sender": {}
			}`,
			want: Delegation{
				ID:        123,
				Delegator: "", // Empty when sender.address is missing
				Timestamp: time.Date(2022, 5, 5, 6, 29, 14, 0, time.UTC),
				Amount:    98765,
				Level:     456,
			},
			wantErr: false,
		},
		{
			name: "invalid timestamp",
			json: `{
				"id": 123,
				"level": 456,
				"timestamp": "invalid",
				"amount": 98765,
				"sender": {
					"address": "tz1a1SAaXRt9yoGMx29rh9FsBF4UzmvojdTL"
				}
			}`,
			want:    Delegation{},
			wantErr: true,
		},
		{
			name: "invalid id (too large)",
			json: `{
				"id": 9223372036854775808,
				"level": 456,
				"timestamp": "2022-05-05T06:29:14Z",
				"amount": 98765,
				"sender": {
					"address": "tz1a1SAaXRt9yoGMx29rh9FsBF4UzmvojdTL"
				}
			}`,
			want:    Delegation{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Delegation
			err := json.Unmarshal([]byte(tt.json), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.ID != tt.want.ID {
					t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
				}
				if got.Delegator != tt.want.Delegator {
					t.Errorf("Delegator = %v, want %v", got.Delegator, tt.want.Delegator)
				}
				if got.Amount != tt.want.Amount {
					t.Errorf("Amount = %v, want %v", got.Amount, tt.want.Amount)
				}
				if got.Level != tt.want.Level {
					t.Errorf("Level = %v, want %v", got.Level, tt.want.Level)
				}
				if !got.Timestamp.Equal(tt.want.Timestamp) {
					t.Errorf("Timestamp = %v, want %v", got.Timestamp, tt.want.Timestamp)
				}
			}
		})
	}
}
