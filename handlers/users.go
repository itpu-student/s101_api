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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GET /api/users/:id — public profile (no phone).
func GetPublicUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	var u models.User
	if err := db.Users().FindOne(ctx, bson.M{"_id": c.Param("id")}).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			utils.NotFound(c, "user not found")
			return
		}
		utils.Internal(c, "user lookup failed")
		return
	}
	reviewCount, _ := db.Reviews().CountDocuments(ctx, bson.M{"user_id": u.ID, "latest": true})
	utils.OK(c, gin.H{
		"user":         u.Public(),
		"review_count": reviewCount,
	})
}

// PUT /api/users/me
func UpdateMe(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in struct {
		Name      *string `json:"name"`
		AvatarURL *string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	update := bson.M{"updated_at": time.Now().UTC()}
	if in.Name != nil {
		update["name"] = *in.Name
	}
	if in.AvatarURL != nil {
		update["avatar_url"] = *in.AvatarURL
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if _, err := db.Users().UpdateByID(ctx, u.ID, bson.M{"$set": update}); err != nil {
		utils.Internal(c, "user update failed")
		return
	}
	_ = db.Users().FindOne(ctx, bson.M{"_id": u.ID}).Decode(u)
	utils.OK(c, u.Public())
}

// DELETE /api/users/me — hard delete.
// Reviews and places stay but are orphaned (user_id / created_by -> null).
// Bookmarks and claim requests are deleted.
func DeleteMe(c *gin.Context) {
	u := middleware.CurrentUser(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	_, _ = db.Reviews().UpdateMany(ctx, bson.M{"user_id": u.ID}, bson.M{"$set": bson.M{"user_id": nil}})
	_, _ = db.Places().UpdateMany(ctx, bson.M{"created_by": u.ID}, bson.M{"$set": bson.M{"created_by": nil}})
	// If they owned a place, revoke ownership too.
	_, _ = db.Places().UpdateMany(ctx, bson.M{"claimed_by": u.ID}, bson.M{"$unset": bson.M{"claimed_by": ""}})
	_, _ = db.Bookmarks().DeleteMany(ctx, bson.M{"user_id": u.ID})
	_, _ = db.ClaimRequests().DeleteMany(ctx, bson.M{"user_id": u.ID})
	if _, err := db.Users().DeleteOne(ctx, bson.M{"_id": u.ID}); err != nil {
		utils.Internal(c, "user delete failed")
		return
	}
	utils.NoContent(c)
}

// GET /api/users/:id/reviews
func UserReviews(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	paging := utils.ParsePaging(c)
	filter := bson.M{"user_id": c.Param("id"), "latest": true}
	cur, err := db.Reviews().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		utils.Internal(c, "user reviews failed")
		return
	}
	var out []models.Review
	if err := cur.All(ctx, &out); err != nil {
		utils.Internal(c, "review decode failed")
		return
	}
	total, _ := db.Reviews().CountDocuments(ctx, filter)
	utils.OK(c, gin.H{"items": out, "page": paging.Page, "limit": paging.Limit, "total": total})
}
