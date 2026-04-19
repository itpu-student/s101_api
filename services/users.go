package services

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var reUsername = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func ListUsersAdmin(ctx context.Context, paging utils.Paging) (*Page[AdminUserView], error) {
	filter := bson.M{}
	cur, err := db.Users().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var raw []models.User
	err = cur.All(ctx, &raw)
	if err != nil {
		return nil, err
	}
	total, _ := db.Users().CountDocuments(ctx, filter)
	items := make([]AdminUserView, 0, len(raw))
	for _, u := range raw {
		items = append(items, *NewAdminUserView(u))
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

func GetPublicUserView(ctx context.Context, id string) (*PublicUserView, error) {
	var u models.User
	err := db.Users().FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	reviewCount, _ := db.Reviews().CountDocuments(ctx, bson.M{"user_id": u.ID, "latest": true})
	return &PublicUserView{
		User:        *u.Public(),
		ReviewCount: reviewCount,
	}, nil
}

func UpdateMe(ctx context.Context, id string, in UpdateMeInput) (*models.PublicUser, error) {
	update := bson.M{"updated_at": time.Now().UTC()}
	if in.Name != nil {
		update["name"] = *in.Name
	}
	if in.Username != nil {
		un := strings.ToLower(strings.TrimSpace(*in.Username))
		if !reUsername.MatchString(un) {
			return nil, ErrBadInput
		}
		// check uniqueness
		var existing models.User
		err := db.Users().FindOne(ctx, bson.M{
			"username": un,
			"_id":      bson.M{"$ne": id},
		}).Decode(&existing)
		if err == nil {
			return nil, ErrConflict
		}
		update["username"] = un
	}
	if in.AvatarKey != nil {
		update["avatar_key"] = *in.AvatarKey
	}
	res, err := db.Users().UpdateByID(ctx, id, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}
	if res.MatchedCount == 0 {
		return nil, ErrNotFound
	}
	var u models.User
	err = db.Users().FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		return nil, err
	}

	return u.Public(), nil
}

func DeleteUserCascade(ctx context.Context, id string) error {
	_, _ = db.Reviews().UpdateMany(ctx, bson.M{"user_id": id}, bson.M{"$set": bson.M{"user_id": nil}})
	_, _ = db.Places().UpdateMany(ctx, bson.M{"created_by": id}, bson.M{"$set": bson.M{"created_by": nil}})
	_, _ = db.Places().UpdateMany(ctx, bson.M{"claimed_by": id}, bson.M{"$unset": bson.M{"claimed_by": ""}})
	_, _ = db.Bookmarks().DeleteMany(ctx, bson.M{"user_id": id})
	_, _ = db.ClaimRequests().DeleteMany(ctx, bson.M{"user_id": id})
	res, err := db.Users().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func ListUserReviews(ctx context.Context, userID string, paging utils.Paging) (*Page[models.Review], error) {
	filter := bson.M{"user_id": userID, "latest": true}
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
