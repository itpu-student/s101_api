package services

import "github.com/itpu-student/s101_api/models"

// Place list sort modes (PlaceFilter.Sort).
const (
	SortTop     = "top"     // highest rated, then most reviewed (default)
	SortRecent  = "recent"  // newest first
	SortNearest = "nearest" // closest to near=lat,lon
)

// Filter structs for list endpoints. A nil field means "no filter on this
// dimension" — services build the bson.M from whatever is set.

type PlaceFilter struct {
	Status          *models.Status
	CategoryId      string
	Query           *string
	Sort            *string
	NearLat         *float64
	NearLon         *float64
	NearMaxDistance *int // meters; required with sort=nearest
	OpenNow         *bool
}

type ReviewFilter struct {
	PlaceID *string `form:"place_id"`
	Latest  *bool   `form:"latest"`
	UserID  *string `form:"user_id"`
}

type ClaimFilter struct {
	Status *models.Status
}
