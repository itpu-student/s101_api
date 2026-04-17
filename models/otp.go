package models

import "time"

type OTPCode struct {
	ID         string    `bson:"_id" json:"id"`
	TelegramID string    `bson:"telegram_id" json:"telegram_id"`
	Phone      string    `bson:"phone" json:"phone"`
	Username   *string   `bson:"username,omitempty" json:"username,omitempty"`
	FirstName  string    `bson:"first_name,omitempty" json:"first_name,omitempty"`
	Code       string    `bson:"code" json:"code"`
	ExpiresAt  time.Time `bson:"expires_at" json:"expires_at"`
	Used       bool      `bson:"used" json:"used"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
}
