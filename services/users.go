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

func ListUsersAdmin(ctx context.Context, paging utils.Paging) (Page[AdminUserView], error) {
	filter := bson.M{}
	cur, err := db.Users().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return Page[AdminUserView]{}, err
	}
	var raw []models.User
	if err := cur.All(ctx, &raw); err != nil {
		return Page[AdminUserView]{}, err
	}
	total, _ := db.Users().CountDocuments(ctx, filter)
	items := make([]AdminUserView, 0, len(raw))
	for _, u := range raw {
		items = append(items, NewAdminUserView(u))
	}
	return NewPage(items, paging, total), nil
}

func SetUserBlocked(ctx context.Context, id string, blocked bool) error {
	res, err := db.Users().UpdateByID(ctx, id, bson.M{"$set": bson.M{
		"blocked":    blocked,
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
