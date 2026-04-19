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

func ListClaimsAdmin(ctx context.Context, f ClaimFilter, paging utils.Paging) (*Page[models.ClaimRequest], error) {
	filter := bson.M{}
	if f.Status != nil {
		filter["status"] = *f.Status
	}
	cur, err := db.ClaimRequests().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var items []models.ClaimRequest
	err = cur.All(ctx, &items)
	if err != nil {
		return nil, err
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
	err := db.ClaimRequests().FindOne(ctx, bson.M{"_id": claimID}).Decode(&cr)
	if err != nil {
		return ErrNotFound
	}

	if status == models.StatusApproved {
		var p models.Place
		err = db.Places().FindOne(ctx, bson.M{"_id": cr.PlaceID}).Decode(&p)
		if err != nil {
			return ErrNotFound
		}
		if p.ClaimedBy != nil && *p.ClaimedBy != cr.UserID {
			return ErrConflict
		}
		_, err = db.Places().UpdateByID(ctx, cr.PlaceID, bson.M{"$set": bson.M{
			"claimed_by": cr.UserID,
			"updated_at": time.Now().UTC(),
		}})
		if err != nil {
			return err
		}
	}

	_, err = db.ClaimRequests().UpdateByID(ctx, cr.ID, bson.M{"$set": bson.M{
		"status":      status,
		"reviewed_by": reviewerID,
		"updated_at":  time.Now().UTC(),
	}})
	if err != nil {
		return err
	}
	return nil
}

func SubmitClaim(ctx context.Context, userID string, in SubmitClaimInput) (*models.ClaimRequest, error) {
	if in.PlaceID == "" || in.Phone == "" {
		return nil, ErrBadInput
	}

	var p models.Place
	err := db.Places().FindOne(ctx, bson.M{"_id": in.PlaceID}).Decode(&p)
	if err != nil {
		return nil, ErrNotFound
	}
	if p.ClaimedBy != nil {
		return nil, ErrAlreadyClaimed
	}

	// Reject duplicate pending claims by the same user.
	existing := db.ClaimRequests().FindOne(ctx, bson.M{
		"place_id": p.ID,
		"user_id":  userID,
		"status":   models.StatusPending,
	})
	if existing.Err() == nil {
		return nil, ErrPendingClaimExists
	}

	now := time.Now().UTC()
	cr := models.ClaimRequest{
		ID:        utils.NewUUIDv7(),
		PlaceID:   p.ID,
		UserID:    userID,
		Phone:     in.Phone,
		Note:      in.Note,
		Status:    models.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = db.ClaimRequests().InsertOne(ctx, cr)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

func ListClaimsForUser(ctx context.Context, userID string, paging utils.Paging) (*Page[models.ClaimRequest], error) {
	filter := bson.M{"user_id": userID}
	cur, err := db.ClaimRequests().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var items []models.ClaimRequest
	err = cur.All(ctx, &items)
	if err != nil {
		return nil, err
	}
	total, _ := db.ClaimRequests().CountDocuments(ctx, filter)
	return NewPage(items, paging, total), nil
}
