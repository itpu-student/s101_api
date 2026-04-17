package models

import "time"

type Review struct {
	ID            string    `bson:"_id" json:"id"`
	PlaceID       string    `bson:"place_id" json:"place_id"`
	UserID        *string   `bson:"user_id" json:"user_id"`
	StarRating    int       `bson:"star_rating" json:"star_rating"`
	PriceRating   *int      `bson:"price_rating,omitempty" json:"price_rating,omitempty"`
	QualityRating *int      `bson:"quality_rating,omitempty" json:"quality_rating,omitempty"`
	Text          string    `bson:"text" json:"text"`
	Images        []string  `bson:"images" json:"images"`
	Latest        bool      `bson:"latest" json:"latest"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
}
