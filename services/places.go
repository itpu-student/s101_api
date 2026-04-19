package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListPlacesAdmin(ctx context.Context, f PlaceFilter, paging utils.Paging) (*Page[models.Place], error) {
	filter := bson.M{}
	if f.Status != nil {
		filter["status"] = *f.Status
	}
	cur, err := db.Places().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var items []models.Place
	err = cur.All(ctx, &items)
	if err != nil {
		return nil, err
	}
	total, _ := db.Places().CountDocuments(ctx, filter)
	return NewPage(items, paging, total), nil
}

func ListPlaces(ctx context.Context, f PlaceFilter, paging utils.Paging) (*Page[PlaceView], error) {
	filter := bson.M{"status": models.StatusApproved}
	if f.Status != nil {
		filter["status"] = *f.Status
	}

	if f.Category != nil {
		if id, ok := ResolveCategoryID(ctx, *f.Category); ok {
			filter["category_id"] = id
		} else {
			return NewPage([]PlaceView{}, paging, 0), nil
		}
	}

	if f.Query != nil {
		q := strings.TrimSpace(*f.Query)
		if q != "" {
			filter["$text"] = bson.M{"$search": q}
		}
	}

	findOpts := options.Find().SetSkip(paging.Skip).SetLimit(int64(paging.Limit))

	if f.NearLat != nil && f.NearLon != nil {
		filter["location"] = bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{"type": "Point", "coordinates": []float64{*f.NearLon, *f.NearLat}},
			},
		}
	} else if f.Sort != nil {
		switch *f.Sort {
		case "recent":
			findOpts.SetSort(bson.D{{Key: "created_at", Value: -1}})
		default: // "top"
			findOpts.SetSort(bson.D{{Key: "avg_rating", Value: -1}, {Key: "review_count", Value: -1}})
		}
	} else {
		findOpts.SetSort(bson.D{{Key: "avg_rating", Value: -1}, {Key: "review_count", Value: -1}})
	}

	cur, err := db.Places().Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	var places []models.Place
	err = cur.All(ctx, &places)
	if err != nil {
		return nil, err
	}
	total, _ := db.Places().CountDocuments(ctx, filter)

	items := make([]PlaceView, 0, len(places))
	for _, p := range places {
		items = append(items, *NewPlaceView(p))
	}
	return NewPage(items, paging, total), nil
}

func GetPlaceView(ctx context.Context, idOrSlug string, viewerID *string, viewerTyp *string) (*PlaceView, error) {
	p, err := FindPlaceByIDOrSlug(ctx, idOrSlug)
	if err != nil {
		return nil, err
	}

	if p.Status != models.StatusApproved {
		allowed := false
		if viewerTyp != nil && *viewerTyp == utils.TypAdmin {
			allowed = true
		} else if viewerID != nil {
			if (p.CreatedBy != nil && *p.CreatedBy == *viewerID) ||
				(p.ClaimedBy != nil && *p.ClaimedBy == *viewerID) {
				allowed = true
			}
		}
		if !allowed {
			return nil, ErrNotFound
		}
	}
	return NewPlaceView(*p), nil
}

func CreatePlace(ctx context.Context, creatorID string, in CreatePlaceInput) (*models.Place, error) {
	catID, ok := ResolveCategoryID(ctx, in.CategoryID)
	if !ok {
		return nil, ErrBadInput
	}

	atc, err := utils.ResolveSOATOID(ctx, in.Lat, in.Lon)
	if err != nil {
		atc = "uz_unknown"
	}

	slug, err := GenerateUniqueSlug(ctx, in.Name)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	p := models.Place{
		ID:          utils.NewUUID(),
		Slug:        slug,
		ATCID:       atc,
		Name:        in.Name,
		CategoryID:  catID,
		Address:     in.Address,
		Phone:       in.Phone,
		Description: in.Description,
		Lat:         in.Lat,
		Lon:         in.Lon,
		Location:    models.NewGeoPoint(in.Lat, in.Lon),
		Images:      CoalesceStrings(in.Images),
		WeeklyHours: in.WeeklyHours,
		Status:      models.StatusPending,
		CreatedBy:   &creatorID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err = db.Places().InsertOne(ctx, p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func EditPlace(ctx context.Context, claimantID string, idOrSlug string, in EditPlaceInput) (*PlaceView, error) {
	p, err := FindPlaceByIDOrSlug(ctx, idOrSlug)
	if err != nil {
		return nil, err
	}

	if p.ClaimedBy == nil || *p.ClaimedBy != claimantID {
		return nil, ErrForbidden
	}

	update := bson.M{"updated_at": time.Now().UTC()}
	if in.Phone != nil {
		update["phone"] = *in.Phone
	}
	if in.Description != nil {
		update["description"] = *in.Description
	}
	if in.WeeklyHours != nil {
		update["weekly_hours"] = *in.WeeklyHours
	}
	if in.Images != nil {
		update["images"] = *in.Images
	}
	_, err = db.Places().UpdateByID(ctx, p.ID, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}

	// reload
	p, err = FindPlaceByIDOrSlug(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	return NewPlaceView(*p), nil
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
func FindPlaceByIDOrSlug(ctx context.Context, idOrSlug string) (*models.Place, error) {
	var p models.Place
	err := db.Places().FindOne(ctx, bson.M{"$or": bson.A{
		bson.M{"_id": idOrSlug},
		bson.M{"slug": idOrSlug},
	}}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// GenerateUniqueSlug appends -2, -3, ... until the slug is unique.
func GenerateUniqueSlug(ctx context.Context, name string) (string, error) {
	base := utils.Slugify(name)
	candidate := base
	for i := 2; i < 1000; i++ {
		err := db.Places().FindOne(ctx, bson.M{"slug": candidate}).Err()
		if errors.Is(err, mongo.ErrNoDocuments) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return "", errors.New("could not generate unique slug")
}

func CoalesceStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
