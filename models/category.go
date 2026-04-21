package models

import "time"

type Category struct {
	ID        string    `bson:"_id" json:"id"`
	Slug      string    `bson:"slug" json:"slug"`
	Name      I18nText  `bson:"name" json:"name"`
	Desc      I18nText  `bson:"desc" json:"desc"`
	Emoji     string    `bson:"emoji" json:"emoji"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
