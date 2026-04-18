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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListBookmarks(ctx context.Context, userID string, paging utils.Paging) (BookmarksView, error) {
	cur, err := db.Bookmarks().Find(ctx, bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return BookmarksView{}, err
	}
	var bms []models.Bookmark
	if err := cur.All(ctx, &bms); err != nil {
		return BookmarksView{}, err
	}
	ids := make([]string, 0, len(bms))
	for _, b := range bms {
		ids = append(ids, b.PlaceID)
	}
	var places []models.Place
	if len(ids) > 0 {
		pcur, err := db.Places().Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
		if err != nil {
			return BookmarksView{}, err
		}
		if err := pcur.All(ctx, &places); err != nil {
			return BookmarksView{}, err
		}
	}
	return BookmarksView{Bookmarks: bms, Places: places}, nil
}

// AddBookmark is idempotent. The bool return is true when the row already
// existed (duplicate key); handler decides between 201 and 200 on that.
func AddBookmark(ctx context.Context, userID, placeID string) (models.Bookmark, bool, error) {
	if err := db.Places().FindOne(ctx, bson.M{"_id": placeID}).Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Bookmark{}, false, ErrNotFound
		}
		return models.Bookmark{}, false, err
	}
	b := models.Bookmark{
		ID:        utils.NewUUIDv7(),
		UserID:    userID,
		PlaceID:   placeID,
		CreatedAt: time.Now().UTC(),
	}
	if _, err := db.Bookmarks().InsertOne(ctx, b); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return models.Bookmark{}, true, nil
		}
		return models.Bookmark{}, false, err
	}
	return b, false, nil
}

func RemoveBookmark(ctx context.Context, userID, placeID string) error {
	_, err := db.Bookmarks().DeleteOne(ctx, bson.M{
		"user_id":  userID,
		"place_id": placeID,
	})
	return err
}
