package models

import "time"

type Report struct {
	ID             string           `bson:"_id" json:"id"`
	UserID         string           `bson:"user_id" json:"user_id"`
	TargetType     ReportTargetType `bson:"target_type" json:"target_type"`
	TargetID       string           `bson:"target_id" json:"target_id"`
	ReportedUserID *string          `bson:"reported_user_id,omitempty" json:"reported_user_id,omitempty"`
	Type           *ReportType      `bson:"type,omitempty" json:"type,omitempty"`
	Text           string           `bson:"text" json:"text"`
	Status         ReportStatus     `bson:"status" json:"status"`
	AdminResponse  *string          `bson:"admin_response,omitempty" json:"admin_response,omitempty"`
	AdminID        *string          `bson:"admin_id,omitempty" json:"admin_id,omitempty"`
	ReviewedAt     *time.Time       `bson:"reviewed_at,omitempty" json:"reviewed_at,omitempty"`
	CreatedAt      time.Time        `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time        `bson:"updated_at" json:"updated_at"`
}
