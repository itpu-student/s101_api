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

func ListBookmarks(ctx context.Context, userID string, paging utils.Paging) (*Page[BookmarkView], error) {
	filter := bson.M{"user_id": userID}
	cur, err := db.Bookmarks().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var bms []models.Bookmark
	err = cur.All(ctx, &bms)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(bms))
	for _, b := range bms {
		ids = append(ids, b.PlaceID)
	}
	placeMap := make(map[string]PlaceView)
	if len(ids) > 0 {
		pcur, err := db.Places().Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
		if err != nil {
			return nil, err
		}
		var rawPlaces []models.Place
		err = pcur.All(ctx, &rawPlaces)
		if err != nil {
			return nil, err
		}
		for _, p := range rawPlaces {
			placeMap[p.ID] = *NewPlaceView(p)
		}
	}

	items := make([]BookmarkView, 0, len(bms))
	for _, b := range bms {
		bv := BookmarkView{Bookmark: b}
		if p, ok := placeMap[b.PlaceID]; ok {
			bv.Place = &p
		}
		items = append(items, bv)
	}

	total, _ := db.Bookmarks().CountDocuments(ctx, filter)
	return NewPage(items, paging, total), nil
}



// AddBookmark is idempotent. The bool return is true when the row already
// existed (duplicate key); handler decides between 201 and 200 on that.
func AddBookmark(ctx context.Context, userID, placeID string) (*models.Bookmark, bool, error) {
	err := db.Places().FindOne(ctx, bson.M{"_id": placeID}).Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, false, NewApiErrS(404, "not_found", "place not found: %s", placeID)
		}
		return nil, false, err
	}
	b := models.Bookmark{
		ID:        utils.NewUUIDv7(),
		UserID:    userID,
		PlaceID:   placeID,
		CreatedAt: time.Now().UTC(),
	}
	_, err = db.Bookmarks().InsertOne(ctx, b)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, true, nil
		}
		return nil, false, err
	}
	return &b, false, nil
}

func RemoveBookmark(ctx context.Context, userID, placeID string) error {
	_, err := db.Bookmarks().DeleteOne(ctx, bson.M{
		"user_id":  userID,
		"place_id": placeID,
	})
	return err
}
