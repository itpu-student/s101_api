package models

import "time"

type Bookmark struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	PlaceID   string    `bson:"place_id" json:"place_id"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
