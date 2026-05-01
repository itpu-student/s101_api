package services

import (
	"time"

	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
)

// Page is the canonical shape for any paginated list response. Used directly
// as the HTTP response body — handlers never hand-build a gin.H for list
// endpoints.
type Page[T any] struct {
	Items []T   `json:"items"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

func NewPage[T any](items []T, paging utils.Paging, total int64) *Page[T] {
	return &Page[T]{Items: items, Page: paging.Page, Limit: paging.Limit, Total: total}
}

// Ok is the shape for write-side acknowledgements. Optional fields echo back
// whatever the endpoint changed (e.g. new status, new blocked flag). They're
// pointers so their zero values (StatusPending, false) still serialize
// correctly — a nil pointer means "this endpoint didn't set it".
type Ok struct {
	Ok      bool           `json:"ok"`
	Status  *models.Status `json:"status,omitempty"`
	Blocked *bool          `json:"blocked,omitempty"`
}

// AdminUserView is the admin-facing user projection — includes phone and
// telegram_id (hidden in the default User JSON tags).
type AdminUserView struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Username   *string   `json:"username,omitempty"`
	TelegramID string    `json:"telegram_id"`
	Phone      string    `json:"phone"`
	AvatarKey  *string   `json:"avatar_key,omitempty"`
	Blocked    bool      `json:"blocked"`
	CreatedAt  time.Time `json:"created_at"`
}

// AdminUserDetailView extends AdminUserView with activity counts for the
// GET /api/admin/users/:id endpoint.
type AdminUserDetailView struct {
	AdminUserView
	ReviewCount       int64     `json:"review_count"`
	PlaceCount        int64     `json:"place_count"`
	ClaimedPlaceCount int64     `json:"claimed_place_count"`
	ClaimCount        int64     `json:"claim_count"`
	ReportCount       int64     `json:"report_count"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AdminLoginOutput struct {
	Token string       `json:"token"`
	Admin models.Admin `json:"admin"`
}

type VerifyCodeOutput struct {
	Token string            `json:"token"`
	User  models.PublicUser `json:"user"`
}

// MeView is the /auth/me response — full profile plus owns_place.
type MeView struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  *string   `json:"username,omitempty"`
	Phone     string    `json:"phone"`
	AvatarKey *string   `json:"avatar_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	OwnsPlace bool      `json:"owns_place"`
	Blocked   bool      `json:"blocked"`
}

// PublicUserView is the public user profile paired with the user's review count.
type PublicUserView struct {
	User        models.PublicUser `json:"user"`
	ReviewCount int64             `json:"review_count"`
}

type BookmarkView struct {
	models.Bookmark
	Place *PlaceView `json:"place,omitempty"`
}

func NewAdminUserView(u models.User) *AdminUserView {
	return &AdminUserView{
		ID:         u.ID,
		Name:       u.Name,
		Username:   u.Username,
		TelegramID: u.TelegramID,
		Phone:      u.Phone,
		AvatarKey:  u.AvatarKey,
		Blocked:    u.Blocked,
		CreatedAt:  u.CreatedAt,
	}
}

type PlaceView struct {
	models.Place
	IsOpen        bool             `json:"is_open"`
	CategoryName  *models.I18nText `json:"category_name,omitempty"`
	CreatedByUser *models.UserMini `json:"created_by_user,omitempty"`
	ClaimedByUser *models.UserMini `json:"claimed_by_user,omitempty"`
}

func NewPlaceView(p models.Place) *PlaceView {
	return &PlaceView{Place: p, IsOpen: utils.IsOpen(p.WeeklyHours, time.Now())}
}

// ReviewView wraps a Review with the author's UserMini for list endpoints.
type ReviewView struct {
	models.Review
	User *models.UserMini `json:"user,omitempty"`
}

// ClaimView wraps a ClaimRequest with the claimant's UserMini for admin endpoints.
type ClaimView struct {
	models.ClaimRequest
	User *models.UserMini `json:"user,omitempty"`
}

// ReportTarget is a uniform preview card for whatever a report points at — a
// review or a place. Frontend renders it the same way regardless of type.
type ReportTarget struct {
	ID        string                  `json:"id"`
	Type      models.ReportTargetType `json:"type"`
	Name      string                  `json:"name"`
	AvatarKey *string                 `json:"avatar_key,omitempty"`
	Content   string                  `json:"content"`
}

// ReportView is the API shape for a Report — embeds the report and adds the
// target card plus reporter/reported_user/admin details when the caller is admin.
type ReportView struct {
	models.Report
	Target       *ReportTarget     `json:"target,omitempty"`
	ReporterUser *models.UserMini  `json:"reporter_user,omitempty"`
	ReportedUser *models.UserMini  `json:"reported_user,omitempty"`
	Admin        *models.AdminMini `json:"admin,omitempty"`
}

// ReportTypeMeta is one entry in the /reports/meta response.
type ReportTypeMeta struct {
	Value models.ReportType `json:"value"`
	Label models.I18nText   `json:"label"`
}

type ReportMeta struct {
	Types          []ReportTypeMeta `json:"types"`
	TextInputLimit int              `json:"text_input_limit"`
}
