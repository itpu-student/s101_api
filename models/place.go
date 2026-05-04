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
	LogoKey      string      `bson:"logo_key" json:"logo_key"`
	Images       []string    `bson:"images" json:"images"`
	WeeklyHours  WeeklyHours `bson:"weekly_hours" json:"weekly_hours"`
	Status       Status      `bson:"status" json:"status" swaggertype:"string" enums:"pending,approved,rejected,suspended"`
	AvgRating    float64     `bson:"avg_rating" json:"avg_rating"`
	ReviewCount  int         `bson:"review_count" json:"review_count"`
	CreatedBy    *string     `bson:"created_by" json:"created_by"`
	ClaimedBy    *string     `bson:"claimed_by,omitempty" json:"claimed_by,omitempty"`
	CreatedAt    time.Time   `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time   `bson:"updated_at" json:"updated_at"`
}

type PlaceMini struct {
	ID          string      `json:"id"`
	Slug        string      `json:"slug"`
	Name        string      `json:"name"`
	CategoryID  string      `json:"category_id"`
	Address     I18nText    `json:"address"`
	Phone       string      `json:"phone"`
	Description I18nText    `json:"description"`
	Lat         float64     `json:"lat"`
	Lon         float64     `json:"lon"`
	Location    GeoPoint    `json:"location"`
	LogoKey     string      `json:"logo_key"`
	WeeklyHours WeeklyHours `json:"weekly_hours"`
	AvgRating   float64     `json:"avg_rating"`
	ReviewCount int         `json:"review_count"`
}

func (p *Place) Mini() *PlaceMini {
	return &PlaceMini{
		ID:          p.ID,
		Slug:        p.Slug,
		Name:        p.Name,
		CategoryID:  p.CategoryID,
		Address:     p.Address,
		Phone:       p.Phone,
		Description: p.Description,
		Lat:         p.Lat,
		Lon:         p.Lon,
		Location:    p.Location,
		LogoKey:     p.LogoKey,
		WeeklyHours: p.WeeklyHours,
		AvgRating:   p.AvgRating,
		ReviewCount: p.ReviewCount,
	}
}

func (wh WeeklyHours) IsBlank() bool {
	for _, day := range [][]HourRange{wh.Mon, wh.Tue, wh.Wed, wh.Thu, wh.Fri, wh.Sat, wh.Sun} {
		if len(day) > 0 {
			return false
		}
	}
	return true
}

