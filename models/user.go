package models

import "time"

type User struct {
	ID         string    `bson:"_id" json:"id"`
	Name       string    `bson:"name" json:"name"`
	Username   *string   `bson:"username,omitempty" json:"username,omitempty"`
	TelegramID string    `bson:"telegram_id" json:"-"`
	Phone      string    `bson:"phone" json:"-"` // private, only returned via /auth/me
	AvatarURL  string    `bson:"avatar_url" json:"avatar_url"`
	Blocked    bool      `bson:"blocked" json:"blocked"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}

// PublicUser strips private fields (phone, telegram_id) for public profile endpoints.
type PublicUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  *string   `json:"username,omitempty"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) Public() PublicUser {
	return PublicUser{
		ID:        u.ID,
		Name:      u.Name,
		Username:  u.Username,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt,
	}
}
