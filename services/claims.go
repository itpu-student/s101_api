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

func ListClaimsAdmin(ctx context.Context, f ClaimFilter, paging utils.Paging) (Page[models.ClaimRequest], error) {
	filter := bson.M{}
	if f.Status != nil {
		filter["status"] = *f.Status
	}
	cur, err := db.ClaimRequests().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return Page[models.ClaimRequest]{}, err
	}
	var items []models.ClaimRequest
	if err := cur.All(ctx, &items); err != nil {
		return Page[models.ClaimRequest]{}, err
	}
	total, _ := db.ClaimRequests().CountDocuments(ctx, filter)
	return NewPage(items, paging, total), nil
}

// ReviewClaim approves or rejects a claim. On approval, sets place.claimed_by
// to the claim's user. Rejects the approval if the place is already claimed
// by someone else.
func ReviewClaim(ctx context.Context, claimID string, status models.Status, reviewerID string) error {
	if status != models.StatusApproved && status != models.StatusRejected {
		return ErrBadInput
	}

	var cr models.ClaimRequest
	if err := db.ClaimRequests().FindOne(ctx, bson.M{"_id": claimID}).Decode(&cr); err != nil {
		return ErrNotFound
	}

	if status == models.StatusApproved {
		var p models.Place
		if err := db.Places().FindOne(ctx, bson.M{"_id": cr.PlaceID}).Decode(&p); err != nil {
			return ErrNotFound
		}
		if p.ClaimedBy != nil && *p.ClaimedBy != cr.UserID {
			return ErrConflict
		}
		if _, err := db.Places().UpdateByID(ctx, cr.PlaceID, bson.M{"$set": bson.M{
			"claimed_by": cr.UserID,
			"updated_at": time.Now().UTC(),
		}}); err != nil {
			return err
		}
	}

	if _, err := db.ClaimRequests().UpdateByID(ctx, cr.ID, bson.M{"$set": bson.M{
		"status":      status,
		"reviewed_by": reviewerID,
		"updated_at":  time.Now().UTC(),
	}}); err != nil {
		return err
	}
	return nil
}
