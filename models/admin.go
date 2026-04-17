package models

import "time"

type Admin struct {
	ID           string    `bson:"_id" json:"id"`
	Username     string    `bson:"username" json:"username"`
	PasswordHash string    `bson:"password" json:"-"`
	Name         string    `bson:"name" json:"name"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}
