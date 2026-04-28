package models

import "time"

type Admin struct {
	ID           string    `bson:"_id" json:"id"`
	Username     string    `bson:"username" json:"username"`
	PasswordHash string    `bson:"password" json:"-"`
	Name         string    `bson:"name" json:"name"`
	AvatarKey    *string   `bson:"avatar_key,omitempty" json:"avatar_key,omitempty"`
	Power        float32   `bson:"power" json:"power"`
	CreatedBy    *string   `bson:"created_by,omitempty" json:"created_by,omitempty"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}
