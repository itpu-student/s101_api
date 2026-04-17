package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// POST /api/auth/verify-code
// Body: { "code": "123456" }
// Looks up an unused, unexpired OTP, marks it used, upserts the user, returns JWT.
func VerifyCode(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid body")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var otp models.OTPCode
	filter := bson.M{
		"code":       req.Code,
		"used":       false,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	}
	if err := db.OTPCodes().FindOne(ctx, filter).Decode(&otp); err != nil {
		if err == mongo.ErrNoDocuments {
			utils.Unauthorized(c, "invalid or expired code")
			return
		}
		utils.Internal(c, "otp lookup failed")
		return
	}

	// mark OTP used
	_, _ = db.OTPCodes().UpdateByID(ctx, otp.ID, bson.M{"$set": bson.M{"used": true}})

	now := time.Now().UTC()
	var user models.User
	err := db.Users().FindOne(ctx, bson.M{"telegram_id": otp.TelegramID}).Decode(&user)
	switch err {
	case nil:
		// login: refresh username if TG sent a new one
		update := bson.M{"updated_at": now}
		if otp.Username != nil && (user.Username == nil || *user.Username != *otp.Username) {
			update["username"] = *otp.Username
		}
		_, _ = db.Users().UpdateByID(ctx, user.ID, bson.M{"$set": update})
	case mongo.ErrNoDocuments:
		// register
		user = models.User{
			ID:         utils.NewUUID(),
			Name:       firstNonEmpty(otp.FirstName, "User"),
			Username:   otp.Username,
			TelegramID: otp.TelegramID,
			Phone:      otp.Phone,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if _, err := db.Users().InsertOne(ctx, user); err != nil {
			utils.Internal(c, "user create failed")
			return
		}
	default:
		utils.Internal(c, "user lookup failed")
		return
	}

	if user.Blocked {
		utils.Forbidden(c, "account is blocked")
		return
	}

	token, err := utils.IssueJWT(user.ID, utils.TypUser)
	if err != nil {
		utils.Internal(c, "token issue failed")
		return
	}
	utils.OK(c, gin.H{"token": token, "user": user.Public()})
}

// GET /api/auth/me
// Returns the current user's full profile including phone and an
// owns_place bool indicating whether they have an approved claim.
func Me(c *gin.Context) {
	u := middleware.CurrentUser(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	count, _ := db.Places().CountDocuments(ctx, bson.M{"claimed_by": u.ID})
	utils.OK(c, gin.H{
		"id":          u.ID,
		"name":        u.Name,
		"username":    u.Username,
		"phone":       u.Phone,
		"avatar_url":  u.AvatarURL,
		"created_at":  u.CreatedAt,
		"owns_place":  count > 0,
		"blocked":     u.Blocked,
	})
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
