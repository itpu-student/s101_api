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

func ListReviewsAdmin(ctx context.Context, f ReviewFilter, paging utils.Paging) (*Page[models.Review], error) {
	filter := bson.M{}
	if f.PlaceID != nil {
		filter["place_id"] = *f.PlaceID
	}
	cur, err := db.Reviews().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var items []models.Review
	err = cur.All(ctx, &items)
	if err != nil {
		return nil, err
	}
	total, _ := db.Reviews().CountDocuments(ctx, filter)
	return NewPage(items, paging, total), nil
}

// AdminDeleteReview removes the review and preserves the "one latest per
// (place, user)" invariant by promoting the most recent remaining review to
// latest=true. Recomputes the place's rating afterwards.
func AdminDeleteReview(ctx context.Context, id string) error {
	var r models.Review
	err := db.Reviews().FindOne(ctx, bson.M{"_id": id}).Decode(&r)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return err
	}
	_, err = db.Reviews().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
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

func ListPlaceReviews(ctx context.Context, placeID string, all bool, paging utils.Paging) (*Page[models.Review], error) {
	filter := bson.M{"place_id": placeID}
	if !all {
		filter["latest"] = true
	}
	cur, err := db.Reviews().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var items []models.Review
	err = cur.All(ctx, &items)
	if err != nil {
		return nil, err
	}
	total, _ := db.Reviews().CountDocuments(ctx, filter)
	return NewPage(items, paging, total), nil
}

func CreateReview(ctx context.Context, userID string, placeIDOrSlug string, in CreateReviewInput) (*models.Review, error) {
	p, err := FindPlaceByIDOrSlug(ctx, placeIDOrSlug)
	if err != nil {
		return nil, err
	}
	if p.Status != models.StatusApproved {
		return nil, ErrForbidden
	}

	// Step 1 — demote any existing latest review for (place, user).
	_, err = db.Reviews().UpdateMany(ctx,
		bson.M{"place_id": p.ID, "user_id": userID, "latest": true},
		bson.M{"$set": bson.M{"latest": false}},
	)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	r := models.Review{
		ID:            utils.NewUUIDv7(),
		PlaceID:       p.ID,
		UserID:        &userID,
		StarRating:    in.StarRating,
		PriceRating:   in.PriceRating,
		QualityRating: in.QualityRating,
		Text:          in.Text,
		Images:        CoalesceStrings(in.Images),
		Latest:        true,
		CreatedAt:     now,
	}
	_, err = db.Reviews().InsertOne(ctx, r)
	if err != nil {
		return nil, err
	}

	if err := RecalcPlaceRating(ctx, p.ID); err != nil {
		return nil, err
	}
	return &r, nil
}

func DeleteUserReview(ctx context.Context, userID string, reviewID string) error {
	var r models.Review
	err := db.Reviews().FindOne(ctx, bson.M{"_id": reviewID}).Decode(&r)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return err
	}
	if r.UserID == nil || *r.UserID != userID {
		return ErrForbidden
	}
	_, err = db.Reviews().DeleteOne(ctx, bson.M{"_id": reviewID})
	if err != nil {
		return err
	}
	// If we just deleted a latest review, promote the previous one to latest=true.
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
