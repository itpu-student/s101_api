package models

import "time"

type ClaimRequest struct {
	ID         string    `bson:"_id" json:"id"`
	PlaceID    string    `bson:"place_id" json:"place_id"`
	UserID     string    `bson:"user_id" json:"user_id"`
	Phone      string    `bson:"phone" json:"phone"`
	Note       string    `bson:"note,omitempty" json:"note,omitempty"`
	Status     Status    `bson:"status" json:"status"`
	ReviewedBy *string   `bson:"reviewed_by,omitempty" json:"reviewed_by,omitempty"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}
