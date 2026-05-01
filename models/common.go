package models

import (
	"encoding/json"
	"fmt"
)

// I18nText is a bilingual string field stored as { en, uz }.
type I18nText struct {
	EN string `bson:"en" json:"en"`
	UZ string `bson:"uz" json:"uz"`
}

// Status is the shared approval state for places and claim requests.
// Internally (Go + Mongo) it is an int — ordered so rejected < pending < approved.
// Externally (REST JSON) it is a string enum: "pending" | "approved" | "rejected".
type Status int

const (
	StatusRejected Status = -10
	StatusPending  Status = 0
	StatusApproved Status = 10
)

func (s Status) IsValid() bool {
	switch s {
	case StatusPending:
	case StatusApproved:
	case StatusRejected:
	default:
		return false
	}
	return true
}

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusApproved:
		return "approved"
	case StatusRejected:
		return "rejected"
	default:
		return "unknown"
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("status must be a string (pending|approved|rejected)")
	}
	v, ok := ParseStatus(str)
	if !ok {
		return fmt.Errorf("invalid status %q: must be pending, approved, or rejected", str)
	}
	*s = v
	return nil
}

// ParseStatus converts a string enum value to its internal int representation.
// Returns ok=false for unknown strings.
func ParseStatus(s string) (Status, bool) {
	switch s {
	case "pending":
		return StatusPending, true
	case "approved":
		return StatusApproved, true
	case "rejected":
		return StatusRejected, true
	}
	return 0, false
}
