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
	LogoKey     *string             `json:"logo_key"`
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
	Name  *models.I18nText `json:"name"`
	Desc  *models.I18nText `json:"desc"`
	Emoji *string          `json:"emoji"`
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
	Username  *string `json:"username"`
	AvatarKey *string `json:"avatar_key"`
}

type SubmitClaimInput struct {
	PlaceID string `json:"place_id"`
	Phone   string `json:"phone"`
	Note    string `json:"note"`
}

type CreateReviewInput struct {
	StarRating    int      `json:"star_rating"`
	PriceRating   *int     `json:"price_rating"`
	QualityRating *int     `json:"quality_rating"`
	Text          string   `json:"text"`
	Images        []string `json:"images"`
}

type CreatePlaceInput struct {
	Name        string             `json:"name"`
	CategoryID  string             `json:"category_id"`
	Address     models.I18nText    `json:"address"`
	Phone       string             `json:"phone"`
	Description models.I18nText    `json:"description"`
	Lat         float64            `json:"lat"`
	Lon         float64            `json:"lon"`
	LogoKey     string             `json:"logo_key"`
	Images      []string           `json:"images"`
	WeeklyHours models.WeeklyHours `json:"weekly_hours"`
}

type EditPlaceInput struct {
	Phone       *string             `json:"phone"`
	Description *models.I18nText    `json:"description"`
	WeeklyHours *models.WeeklyHours `json:"weekly_hours"`
	LogoKey     *string             `json:"logo_key"`
	Images      *[]string           `json:"images"`
}

type SubmitReportInput struct {
	TargetType models.ReportTargetType `json:"target_type"`
	TargetID   string                  `json:"target_id"`
	Type       *models.ReportType      `json:"type"`
	Text       string                  `json:"text"`
}

type EditReportInput struct {
	Type *models.ReportType `json:"type"`
	Text *string            `json:"text"`
}

type ReviewReportInput struct {
	Status             models.ReportStatus `json:"status"`
	AdminResponse      *string             `json:"admin_response"`
	DeleteTargetReview bool                `json:"delete_target_review"`
	BlockReportedUser  bool                `json:"block_reported_user"`
}

type ReportFilter struct {
	Status         *models.ReportStatus
	Type           *models.ReportType
	TargetType     *models.ReportTargetType
	TargetID       *string
	ReportedUserID *string
	AdminID        *string
}
