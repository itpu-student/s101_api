package models

import "time"

// GeoPoint is a GeoJSON Point: { type: "Point", coordinates: [lon, lat] }.
type GeoPoint struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}

func NewGeoPoint(lat, lon float64) GeoPoint {
	return GeoPoint{Type: "Point", Coordinates: []float64{lon, lat}}
}

// WeeklyHours stores open/close ranges per weekday.
// Each day is either nil (closed) or a list of [open, close] HH:MM strings.
type WeeklyHours struct {
	Mon []HourRange `bson:"mon,omitempty" json:"mon,omitempty"`
	Tue []HourRange `bson:"tue,omitempty" json:"tue,omitempty"`
	Wed []HourRange `bson:"wed,omitempty" json:"wed,omitempty"`
	Thu []HourRange `bson:"thu,omitempty" json:"thu,omitempty"`
	Fri []HourRange `bson:"fri,omitempty" json:"fri,omitempty"`
	Sat []HourRange `bson:"sat,omitempty" json:"sat,omitempty"`
	Sun []HourRange `bson:"sun,omitempty" json:"sun,omitempty"`
}

type HourRange struct {
	Open  string `bson:"open" json:"open"`   // "09:00"
	Close string `bson:"close" json:"close"` // "22:00"
}

type Place struct {
	ID           string      `bson:"_id" json:"id"`
	Slug         string      `bson:"slug" json:"slug"`
	ATCID        string      `bson:"atc_id" json:"atc_id"`
	Name         string      `bson:"name" json:"name"`
	CategoryID   string      `bson:"category_id" json:"category_id"`
	Address      I18nText    `bson:"address" json:"address"`
	Phone        string      `bson:"phone" json:"phone"`
	Description  I18nText    `bson:"description" json:"description"`
	Lat          float64     `bson:"lat" json:"lat"`
	Lon          float64     `bson:"lon" json:"lon"`
	Location     GeoPoint    `bson:"location" json:"location"`
	Images       []string    `bson:"images" json:"images"`
	WeeklyHours  WeeklyHours `bson:"weekly_hours" json:"weekly_hours"`
	Status       int         `bson:"status" json:"status"`
	AvgRating    float64     `bson:"avg_rating" json:"avg_rating"`
	ReviewCount  int         `bson:"review_count" json:"review_count"`
	CreatedBy    *string     `bson:"created_by" json:"created_by"`
	ClaimedBy    *string     `bson:"claimed_by,omitempty" json:"claimed_by,omitempty"`
	CreatedAt    time.Time   `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time   `bson:"updated_at" json:"updated_at"`
}
