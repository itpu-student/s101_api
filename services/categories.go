package services

import (
	"context"
	"time"

	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListCategories(ctx context.Context) ([]models.Category, error) {
	cur, err := db.Categories().Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "slug", Value: 1}}))
	if err != nil {
		return nil, err
	}
	var out []models.Category
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func EditCategory(ctx context.Context, id string, in EditCategoryInput) error {
	update := bson.M{"updated_at": time.Now().UTC()}
	if in.Name != nil {
		update["name"] = *in.Name
	}
	if in.Desc != nil {
		update["desc"] = *in.Desc
	}
	res, err := db.Categories().UpdateByID(ctx, id, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// ResolveCategoryID accepts a category UUID or slug and returns the canonical ID.
func ResolveCategoryID(ctx context.Context, val string) (string, bool) {
	var cat models.Category
	err := db.Categories().FindOne(ctx, bson.M{"$or": bson.A{
		bson.M{"_id": val},
		bson.M{"slug": val},
	}}).Decode(&cat)
	if err != nil {
		return "", false
	}
	return cat.ID, true
}
