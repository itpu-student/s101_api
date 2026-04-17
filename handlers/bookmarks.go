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

// GET /api/bookmarks — returns the user's bookmarked places.
func ListBookmarks(c *gin.Context) {
	u := middleware.CurrentUser(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	paging := utils.ParsePaging(c)
	cur, err := db.Bookmarks().Find(ctx, bson.M{"user_id": u.ID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		utils.Internal(c, "bookmarks list failed")
		return
	}
	var bms []models.Bookmark
	if err := cur.All(ctx, &bms); err != nil {
		utils.Internal(c, "bookmarks decode failed")
		return
	}
	ids := make([]string, 0, len(bms))
	for _, b := range bms {
		ids = append(ids, b.PlaceID)
	}
	var places []models.Place
	if len(ids) > 0 {
		pcur, _ := db.Places().Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
		_ = pcur.All(ctx, &places)
	}
	utils.OK(c, gin.H{"bookmarks": bms, "places": places})
}

// POST /api/bookmarks/:placeId
func AddBookmark(c *gin.Context) {
	u := middleware.CurrentUser(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	placeID := c.Param("placeId")

	if err := db.Places().FindOne(ctx, bson.M{"_id": placeID}).Err(); err != nil {
		utils.NotFound(c, "place not found")
		return
	}

	b := models.Bookmark{
		ID:        utils.NewUUIDv7(),
		UserID:    u.ID,
		PlaceID:   placeID,
		CreatedAt: time.Now().UTC(),
	}
	if _, err := db.Bookmarks().InsertOne(ctx, b); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			utils.OK(c, gin.H{"ok": true, "already": true})
			return
		}
		utils.Internal(c, "bookmark insert failed")
		return
	}
	utils.Created(c, b)
}

// DELETE /api/bookmarks/:placeId
func RemoveBookmark(c *gin.Context) {
	u := middleware.CurrentUser(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err := db.Bookmarks().DeleteOne(ctx, bson.M{
		"user_id":  u.ID,
		"place_id": c.Param("placeId"),
	})
	if err != nil {
		utils.Internal(c, "bookmark delete failed")
		return
	}
	utils.NoContent(c)
}
