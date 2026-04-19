package models

import "time"

const (
	FileUsageAvatar = "avatar"
	FileUsageReview = "review"
	FileUsagePlace  = "place"
)

type File struct {
	ID        string     `bson:"_id"                  json:"id"`
	Ext       string     `bson:"ext"                  json:"ext"`
	Path      string     `bson:"path"                 json:"path"`
	Usage     string     `bson:"usage"                json:"usage"`
	CreatedAt time.Time  `bson:"created_at"           json:"created_at"`
	CreatedBy string     `bson:"created_by"           json:"created_by"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}
