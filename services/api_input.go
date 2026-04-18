package services

import "github.com/itpu-student/s101_api/models"

// Inputs for write-side endpoints. Pure Go structs — handlers bind into these
// with c.ShouldBindJSON, services validate.

type SetPlaceStatusInput struct {
	Status models.Status `json:"status"`
}

type AdminEditPlaceInput struct {
	Name        *string             `json:"name"`
	CategoryID  *string             `json:"category_id"`
	Address     *models.I18nText    `json:"address"`
	Phone       *string             `json:"phone"`
	Description *models.I18nText    `json:"description"`
	Lat         *float64            `json:"lat"`
	Lon         *float64            `json:"lon"`
	Images      *[]string           `json:"images"`
	WeeklyHours *models.WeeklyHours `json:"weekly_hours"`
}

type BlockUserInput struct {
	Blocked bool `json:"blocked"`
}

type ReviewClaimInput struct {
	Status models.Status `json:"status"`
}

type EditCategoryInput struct {
	Name *models.I18nText `json:"name"`
	Desc *models.I18nText `json:"desc"`
}

type AdminLoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VerifyCodeInput struct {
	Code string `json:"code"`
}

type UpdateMeInput struct {
	Name      *string `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}
