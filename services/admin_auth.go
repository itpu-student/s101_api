package services

import (
	"context"
	"errors"

	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AdminLogin verifies admin credentials and issues a JWT. Returns a generic
// "unauthorized: invalid credentials" for both "unknown username" and "wrong
// password" to avoid leaking which of the two was incorrect.
func AdminLogin(ctx context.Context, in AdminLoginInput) (*AdminLoginOutput, error) {
	if in.Username == "" || in.Password == "" {
		return nil, NewApiErr("bad_input", "username and password are required")
	}
	var a models.Admin
	err := db.Admins().FindOne(ctx, bson.M{"username": in.Username}).Decode(&a)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewApiErrS(401, "unauthorized", "invalid credentials")
		}
		return nil, err
	}
	if err := utils.ComparePassword(a.PasswordHash, in.Password); err != nil {
		return nil, NewApiErrS(401, "unauthorized", "invalid credentials")
	}
	token, err := utils.IssueJWT(a.ID, utils.TypAdmin)
	if err != nil {
		return nil, err
	}
	return &AdminLoginOutput{Token: token, Admin: a}, nil
}
