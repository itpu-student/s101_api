package services

import (
	"context"
	"time"

	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListReviewsAdmin(ctx context.Context, f ReviewFilter, paging utils.Paging) (Page[models.Review], error) {
	filter := bson.M{}
	if f.PlaceID != nil {
		filter["place_id"] = *f.PlaceID
	}
	cur, err := db.Reviews().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return Page[models.Review]{}, err
	}
	var items []models.Review
	if err := cur.All(ctx, &items); err != nil {
		return Page[models.Review]{}, err
	}
	total, _ := db.Reviews().CountDocuments(ctx, filter)
	return NewPage(items, paging, total), nil
}

// AdminDeleteReview removes the review and preserves the "one latest per
// (place, user)" invariant by promoting the most recent remaining review to
// latest=true. Recomputes the place's rating afterwards.
func AdminDeleteReview(ctx context.Context, id string) error {
	var r models.Review
	if err := db.Reviews().FindOne(ctx, bson.M{"_id": id}).Decode(&r); err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		return err
	}
	if _, err := db.Reviews().DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return err
	}
	if r.Latest && r.UserID != nil {
		var prev models.Review
		err := db.Reviews().FindOne(ctx,
			bson.M{"place_id": r.PlaceID, "user_id": *r.UserID},
			options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}}),
		).Decode(&prev)
		if err == nil {
			_, _ = db.Reviews().UpdateByID(ctx, prev.ID, bson.M{"$set": bson.M{"latest": true}})
		}
	}
	return RecalcPlaceRating(ctx, r.PlaceID)
}

// RecalcPlaceRating recomputes avg_rating & review_count on a place from its
// latest=true reviews. Safe to call after any review mutation.
func RecalcPlaceRating(ctx context.Context, placeID string) error {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"place_id": placeID, "latest": true}}},
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"avg": bson.M{"$avg": "$star_rating"},
			"cnt": bson.M{"$sum": 1},
		}}},
	}
	cur, err := db.Reviews().Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	var out struct {
		Avg float64 `bson:"avg"`
		Cnt int     `bson:"cnt"`
	}
	if cur.Next(ctx) {
		_ = cur.Decode(&out)
	}
	_, err = db.Places().UpdateByID(ctx, placeID, bson.M{"$set": bson.M{
		"avg_rating":   round1(out.Avg),
		"review_count": out.Cnt,
		"updated_at":   time.Now().UTC(),
	}})
	return err
}

func round1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}
