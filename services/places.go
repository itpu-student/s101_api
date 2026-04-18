package services

import (
	"context"
	"time"

	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListPlacesAdmin(ctx context.Context, f PlaceFilter, paging utils.Paging) (Page[models.Place], error) {
	filter := bson.M{}
	if f.Status != nil {
		filter["status"] = *f.Status
	}
	cur, err := db.Places().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return Page[models.Place]{}, err
	}
	var items []models.Place
	if err := cur.All(ctx, &items); err != nil {
		return Page[models.Place]{}, err
	}
	total, _ := db.Places().CountDocuments(ctx, filter)
	return NewPage(items, paging, total), nil
}

func SetPlaceStatus(ctx context.Context, id string, status models.Status) error {
	if status != models.StatusPending && status != models.StatusApproved && status != models.StatusRejected {
		return ErrBadInput
	}
	res, err := db.Places().UpdateByID(ctx, id, bson.M{"$set": bson.M{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func AdminEditPlace(ctx context.Context, id string, in AdminEditPlaceInput) error {
	update := bson.M{"updated_at": time.Now().UTC()}
	if in.Name != nil {
		update["name"] = *in.Name
	}
	if in.CategoryID != nil {
		catID, ok := ResolveCategoryID(ctx, *in.CategoryID)
		if !ok {
			return ErrBadInput
		}
		update["category_id"] = catID
	}
	if in.Address != nil {
		update["address"] = *in.Address
	}
	if in.Phone != nil {
		update["phone"] = *in.Phone
	}
	if in.Description != nil {
		update["description"] = *in.Description
	}
	if in.Lat != nil {
		update["lat"] = *in.Lat
	}
	if in.Lon != nil {
		update["lon"] = *in.Lon
	}
	if in.Lat != nil && in.Lon != nil {
		update["location"] = models.NewGeoPoint(*in.Lat, *in.Lon)
	}
	if in.Images != nil {
		update["images"] = *in.Images
	}
	if in.WeeklyHours != nil {
		update["weekly_hours"] = *in.WeeklyHours
	}
	res, err := db.Places().UpdateByID(ctx, id, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// DeletePlaceCascade removes the place and all related rows (reviews,
// bookmarks, claim requests) in one call. Child deletes are best-effort —
// partial failures don't abort the place delete, since orphaned child rows
// can be cleaned up later but a half-deleted place is worse.
func DeletePlaceCascade(ctx context.Context, id string) error {
	_, _ = db.Reviews().DeleteMany(ctx, bson.M{"place_id": id})
	_, _ = db.Bookmarks().DeleteMany(ctx, bson.M{"place_id": id})
	_, _ = db.ClaimRequests().DeleteMany(ctx, bson.M{"place_id": id})
	res, err := db.Places().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
