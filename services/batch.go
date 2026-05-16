package services

import (
	"context"

	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"go.mongodb.org/mongo-driver/bson"
)

func dedupeIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

func fetchUserMiniMap(ctx context.Context, ids []string) map[string]*models.UserMini {
	ids = dedupeIDs(ids)
	m := make(map[string]*models.UserMini, len(ids))
	if len(ids) == 0 {
		return m
	}
	cur, err := db.Users().Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return m
	}
	var users []models.User
	_ = cur.All(ctx, &users)
	for i := range users {
		m[users[i].ID] = users[i].Mini()
	}
	return m
}

func fetchPlaceMiniMap(ctx context.Context, ids []string) map[string]*models.PlaceMini {
	ids = dedupeIDs(ids)
	m := make(map[string]*models.PlaceMini, len(ids))
	if len(ids) == 0 {
		return m
	}
	cur, err := db.Places().Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return m
	}
	var places []models.Place
	_ = cur.All(ctx, &places)
	for i := range places {
		m[places[i].ID] = places[i].Mini()
	}
	return m
}

func fetchAdminMiniMap(ctx context.Context, ids []string) map[string]*models.AdminMini {
	ids = dedupeIDs(ids)
	m := make(map[string]*models.AdminMini, len(ids))
	if len(ids) == 0 {
		return m
	}
	cur, err := db.Admins().Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return m
	}
	var admins []models.Admin
	_ = cur.All(ctx, &admins)
	for i := range admins {
		m[admins[i].ID] = admins[i].Mini()
	}
	return m
}

func fetchPlaceMap(ctx context.Context, ids []string) map[string]*models.Place {
	ids = dedupeIDs(ids)
	m := make(map[string]*models.Place, len(ids))
	if len(ids) == 0 {
		return m
	}
	cur, err := db.Places().Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return m
	}
	var places []models.Place
	_ = cur.All(ctx, &places)
	for i := range places {
		p := places[i]
		m[p.ID] = &p
	}
	return m
}

func fetchReviewMap(ctx context.Context, ids []string) map[string]*models.Review {
	ids = dedupeIDs(ids)
	m := make(map[string]*models.Review, len(ids))
	if len(ids) == 0 {
		return m
	}
	cur, err := db.Reviews().Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return m
	}
	var reviews []models.Review
	_ = cur.All(ctx, &reviews)
	for i := range reviews {
		r := reviews[i]
		m[r.ID] = &r
	}
	return m
}
