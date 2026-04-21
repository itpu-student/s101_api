package services

import "github.com/itpu-student/s101_api/models"

// Filter structs for list endpoints. A nil field means "no filter on this
// dimension" — services build the bson.M from whatever is set.

type PlaceFilter struct {
	Status   *models.Status
	CategoryId string
	Query    *string
	Sort     *string
	NearLat  *float64
	NearLon  *float64
}

type ReviewFilter struct {
	PlaceID *string
}

type ClaimFilter struct {
	Status *models.Status
}
