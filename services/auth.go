package services

import (
	"context"
	"errors"
	"time"

	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	. "github.com/itpu-student/s101_api/utils/api_err"
)

// VerifyCode consumes a 6-digit OTP, upserts the user (first-time -> register,
// returning user -> refresh username from TG), and issues a JWT.
func VerifyCode(ctx context.Context, in VerifyCodeInput) (*VerifyCodeOutput, error) {
		// return nil, NewApiErr(AetBadInput, "code must be 6 digits")

	var otp models.OTPCode
	err := db.OTPCodes().FindOne(ctx, bson.M{
		"code":       in.Code,
		"used":       false,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	}).Decode(&otp)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewApiErrS(401, AetUnauthorized, "invalid or expired code: %s", in.Code)
		}
		return nil, err
	}
	_, err = db.OTPCodes().UpdateByID(ctx, otp.ID, bson.M{"$set": bson.M{"used": true}})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var user models.User
	err = db.Users().FindOne(ctx, bson.M{"telegram_id": otp.TelegramID}).Decode(&user)
	switch {
	case err == nil:
		update := bson.M{"updated_at": now}
		if otp.Username != nil && (user.Username == nil || *user.Username != *otp.Username) {
			update["username"] = *otp.Username
			user.Username = otp.Username
		}
		_, _ = db.Users().UpdateByID(ctx, user.ID, bson.M{"$set": update})
	case errors.Is(err, mongo.ErrNoDocuments):
		user = models.User{
			ID:         utils.NewUUID(),
			Name:       firstNonEmpty(otp.FirstName, "User"),
			Username:   otp.Username,
			TelegramID: otp.TelegramID,
			Phone:      otp.Phone,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		_, err = db.Users().InsertOne(ctx, user)
		if err != nil {
			return nil, err
		}
	default:
		return nil, err
	}

	if user.Blocked {
		return nil, NewApiErrS(403, AetForbidden, "user is blocked")
	}

	token, err := utils.IssueJWT(user.ID, utils.TypUser)
	if err != nil {
		return nil, err
	}
	return &VerifyCodeOutput{Token: token, User: *user.Public()}, nil
}

// GetMe builds the /auth/me payload for the already-loaded user.
func GetMe(ctx context.Context, u *models.User) (*MeView, error) {
	count, _ := db.Places().CountDocuments(ctx, bson.M{"claimed_by": u.ID})
	return &MeView{
		ID:        u.ID,
		Name:      u.Name,
		Username:  u.Username,
		Phone:     u.Phone,
		AvatarKey: u.AvatarKey,
		CreatedAt: u.CreatedAt,
		OwnsPlace: count > 0,
		Blocked:   u.Blocked,
	}, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
