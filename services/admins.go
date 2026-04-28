package services

import (
	"context"
	"errors"
	"time"

	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	. "github.com/itpu-student/s101_api/utils/api_err"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListAdmins(ctx context.Context, paging utils.Paging) (*Page[models.Admin], error) {
	filter := bson.M{}
	cur, err := db.Admins().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var items []models.Admin
	if err = cur.All(ctx, &items); err != nil {
		return nil, err
	}
	total, _ := db.Admins().CountDocuments(ctx, filter)
	return NewPage(items, paging, total), nil
}

func GetAdmin(ctx context.Context, id string) (*models.Admin, error) {
	var a models.Admin
	if err := db.Admins().FindOne(ctx, bson.M{"_id": id}).Decode(&a); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewApiErrS(404, AetNotFound, "admin not found")
		}
		return nil, err
	}
	return &a, nil
}

func CreateAdmin(ctx context.Context, in CreateAdminInput, creator *models.Admin) (*models.Admin, error) {
	if creator.Power < 50 {
		return nil, NewApiErrS(403, AetForbidden, "requires power >= 50 to create admins")
	}

	if in.Power < 0 || in.Power >= creator.Power {
		return nil, NewApiErr(AetBadInput, "power must be >= 0 and < your power (%.0f)", creator.Power)
	}
	hash, err := utils.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	a := models.Admin{
		ID:           utils.NewUUID(),
		Username:     in.Username,
		PasswordHash: hash,
		Name:         in.Name,
		Power:        in.Power,
		CreatedBy:    &creator.ID,
		CreatedAt:    time.Now().UTC(),
	}
	if _, err = db.Admins().InsertOne(ctx, a); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, NewApiErrS(409, AetBadInput, "username already taken")
		}
		return nil, err
	}
	return &a, nil
}

func EditAdmin(ctx context.Context, id string, in EditAdminInput, editor *models.Admin) (*models.Admin, error) {
	target, err := GetAdmin(ctx, id)
	if err != nil {
		return nil, err
	}
	if editor.Power <= target.Power {
		return nil, NewApiErrS(403, AetForbidden, "cannot edit admin with equal or higher power")
	}
	update := bson.M{}
	if in.Name != nil {
		update["name"] = *in.Name
	}
	if in.Password != nil {
		hash, err := utils.HashPassword(*in.Password)
		if err != nil {
			return nil, err
		}
		update["password"] = hash
	}
	if in.Power != nil {
		if *in.Power < 0 || *in.Power >= editor.Power {
			return nil, NewApiErr(AetBadInput, "power must be >= 0 and < your power (%.0f)", editor.Power)
		}
		update["power"] = *in.Power
	}
	if len(update) == 0 {
		return target, nil
	}
	if _, err = db.Admins().UpdateByID(ctx, id, bson.M{"$set": update}); err != nil {
		return nil, err
	}
	return GetAdmin(ctx, id)
}
