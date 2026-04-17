package models

// I18nText is a bilingual string field stored as { en, uz }.
type I18nText struct {
	EN string `bson:"en" json:"en"`
	UZ string `bson:"uz" json:"uz"`
}

// Status values shared across places and claim_requests.
const (
	StatusPending  = 0
	StatusApproved = 10
	StatusRejected = -10
)
