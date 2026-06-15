package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	. "github.com/itpu-student/s101_api/utils/api_err"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListPlacesAdmin(ctx context.Context, f PlaceFilter, paging utils.Paging) (*Page[PlaceView], error) {
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
	var raw []models.Place
	err = cur.All(ctx, &raw)
	if err != nil {
		return nil, err
	}
	total, _ := db.Places().CountDocuments(ctx, filter)
	userIDs := make([]string, 0, len(raw)*2)
	for _, p := range raw {
		if p.CreatedBy != nil {
			userIDs = append(userIDs, *p.CreatedBy)
		}
		if p.ClaimedBy != nil {
			userIDs = append(userIDs, *p.ClaimedBy)
		}
	}
	userMap := fetchUserMiniMap(ctx, userIDs)
	items := make([]PlaceView, 0, len(raw))
	for _, p := range raw {
		v := NewPlaceView(p)
		if p.CreatedBy != nil {
			v.CreatedByUser = userMap[*p.CreatedBy]
		}
		if p.ClaimedBy != nil {
			v.ClaimedByUser = userMap[*p.ClaimedBy]
		}
		items = append(items, *v)
	}
	return NewPage(items, paging, total), nil
}

// earthRadiusMeters is the sphere radius used to convert a metric search radius
// into radians for $centerSphere.
const earthRadiusMeters = 6378100.0

// ListPlaces returns approved places for the public catalog, filtered by
// category / search query and ordered by the requested sort mode.
func ListPlaces(ctx context.Context, f PlaceFilter, paging utils.Paging) (*Page[PlaceView], error) {
	filter := bson.M{"status": models.StatusApproved}
	if f.CategoryId != "" {
		filter["category_id"] = f.CategoryId
	}

	var search bson.M
	if f.Query != nil {
		if q := strings.TrimSpace(*f.Query); q != "" {
			search = buildSearch(q)
		}
	}

	sortMode := SortTop
	if f.Sort != nil && *f.Sort != "" {
		sortMode = *f.Sort
	}

	findOpts := options.Find()
	switch sortMode {
	case SortNearest:
		if f.NearLat == nil || f.NearLon == nil {
			return nil, NewApiErr(AetBadInput, "sort=nearest requires near=lat,lon")
		}
		if f.NearMaxDistance == nil {
			return nil, NewApiErr(AetBadInput, "sort=nearest requires near_max_distance (meters)")
		}
		// $geoWithin/$centerSphere is a plain filter (unlike $nearSphere) so it
		// composes with $text and regex searches in a single query. The radius is
		// in radians. Distance ordering is applied in listPlacesNearest.
		filter["location"] = bson.M{"$geoWithin": bson.M{"$centerSphere": bson.A{
			bson.A{*f.NearLon, *f.NearLat}, float64(*f.NearMaxDistance) / earthRadiusMeters,
		}}}
		if search != nil {
			mergeFilter(filter, search)
		}
		return listPlacesNearest(ctx, filter, *f.NearLat, *f.NearLon, f, paging)
	case SortRecent:
		findOpts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	case SortTop:
		findOpts.SetSort(bson.D{{Key: "avg_rating", Value: -1}, {Key: "review_count", Value: -1}})
	default:
		return nil, NewApiErr(AetBadInput, "invalid sort: %s", sortMode)
	}

	if search != nil {
		mergeFilter(filter, search)
	}

	if f.OpenNow != nil && *f.OpenNow {
		return listPlacesOpenNow(ctx, filter, findOpts, paging)
	}

	findOpts.SetSkip(paging.Skip).SetLimit(int64(paging.Limit))
	cur, err := db.Places().Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	var places []models.Place
	if err = cur.All(ctx, &places); err != nil {
		return nil, err
	}

	total, err := db.Places().CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]PlaceView, 0, len(places))
	for _, p := range places {
		items = append(items, *NewPlaceView(p))
	}
	return NewPage(items, paging, total), nil
}

// listPlacesNearest loads the radius-bounded matches (the $geoWithin filter
// already caps the set), applies open_now, orders by distance from the origin,
// then paginates in memory. $geoWithin doesn't sort by distance and can't be
// combined with a $text-driven server sort, so ordering is done here.
func listPlacesNearest(ctx context.Context, filter bson.M, lat, lon float64, f PlaceFilter, paging utils.Paging) (*Page[PlaceView], error) {
	cur, err := db.Places().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var matched []models.Place
	if err = cur.All(ctx, &matched); err != nil {
		return nil, err
	}

	if f.OpenNow != nil && *f.OpenNow {
		now := time.Now()
		open := make([]models.Place, 0, len(matched))
		for _, p := range matched {
			if v := utils.IsOpen(p.WeeklyHours, now); v != nil && *v {
				open = append(open, p)
			}
		}
		matched = open
	}

	sort.Slice(matched, func(i, j int) bool {
		return geoDistSq(lat, lon, matched[i].Lat, matched[i].Lon) <
			geoDistSq(lat, lon, matched[j].Lat, matched[j].Lon)
	})

	return pagePlaces(matched, paging), nil
}

// geoDistSq is a cheap squared-distance proxy (equirectangular, longitude scaled
// by latitude) — monotonic with true distance over the small radii used here, so
// it's enough for ordering without a sqrt or full haversine.
func geoDistSq(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := lat1 - lat2
	dLon := (lon1 - lon2) * math.Cos(lat1*math.Pi/180)
	return dLat*dLat + dLon*dLon
}

// searchWordMaxLen is the rune-length cutoff below which a single-word query is
// treated as a quick prefix/typeahead lookup (regex contains on name+slug)
// rather than a full $text search.
const searchWordMaxLen = 7

// buildSearch returns the filter fragment for a search query, choosing the
// strategy by query shape:
//
//   - one short word (< searchWordMaxLen runes): case-insensitive "contains"
//     regex on name + slug. Catches partial typing like "Choy" -> "Choyxona",
//     which $text (whole-word) can't, and keeps typeahead snappy.
//   - anything longer or multi-word: $text over the index spanning name, slug,
//     description, and address — word-aware and index-backed for real queries.
//
// Regex input is escaped so user text can't inject metacharacters.
func buildSearch(q string) bson.M {
	if utf8.RuneCountInString(q) < searchWordMaxLen && len(strings.Fields(q)) == 1 {
		rx := primitive.Regex{Pattern: regexp.QuoteMeta(q), Options: "i"}
		return bson.M{"$or": bson.A{
			bson.M{"name": rx},
			bson.M{"slug": rx},
		}}
	}
	return bson.M{"$text": bson.M{"$search": q}}
}

// mergeFilter copies src's keys into dst.
func mergeFilter(dst, src bson.M) {
	for k, v := range src {
		dst[k] = v
	}
}

// listPlacesOpenNow fetches all matching places (no DB-level skip/limit) and
// filters in-memory by IsOpen, then paginates. MongoDB can't evaluate
// time-based weekly_hours natively.
func listPlacesOpenNow(ctx context.Context, filter bson.M, sortOpts *options.FindOptions, paging utils.Paging) (*Page[PlaceView], error) {
	cur, err := db.Places().Find(ctx, filter, sortOpts)
	if err != nil {
		return nil, err
	}
	var all []models.Place
	if err = cur.All(ctx, &all); err != nil {
		return nil, err
	}
	now := time.Now()
	open := make([]models.Place, 0, len(all))
	for _, p := range all {
		if v := utils.IsOpen(p.WeeklyHours, now); v != nil && *v {
			open = append(open, p)
		}
	}
	return pagePlaces(open, paging), nil
}

// pagePlaces applies in-memory pagination to an already-ordered slice and wraps
// the window as PlaceViews.
func pagePlaces(places []models.Place, paging utils.Paging) *Page[PlaceView] {
	total := int64(len(places))
	start := min(int(paging.Skip), len(places))
	end := min(start+paging.Limit, len(places))
	views := make([]PlaceView, 0, end-start)
	for _, p := range places[start:end] {
		views = append(views, *NewPlaceView(p))
	}
	return NewPage(views, paging, total)
}

func GetPlaceView(ctx context.Context, idOrSlug string, viewerID *string, viewerTyp *string) (*PlaceView, error) {
	var p *models.Place
	var err error
	if _, perr := uuid.Parse(idOrSlug); perr == nil {
		p, err = FindPlaceByID(ctx, idOrSlug)
	} else {
		p, err = FindPlaceBySlug(ctx, idOrSlug)
	}
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
			return nil, NewApiErrS(404, AetNotFound, "place not found: %s", idOrSlug)
		}
	}
	v := NewPlaceView(*p)
	v.SavedCount, _ = db.Bookmarks().CountDocuments(ctx, bson.M{"place_id": p.ID})
	return v, nil
}

func CreatePlace(ctx context.Context, creatorID string, in CreatePlaceInput) (*models.Place, error) {
	if _, err := uuid.Parse(in.CategoryID); err != nil {
		return nil, NewApiErr(AetBadInput, "category_id must be a UUID")
	}
	if err := db.Categories().FindOne(ctx, bson.M{"_id": in.CategoryID}).Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewApiErr(AetBadInput, "category not found: %s", in.CategoryID)
		}
		return nil, err
	}
	catID := in.CategoryID

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
		LogoKey:     in.LogoKey,
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

func EditPlace(ctx context.Context, claimantID string, id string, in EditPlaceInput) (*PlaceView, error) {
	var p models.Place
	err := db.Places().FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewApiErrS(404, AetNotFound, "place not found: %s", id)
		}
		return nil, err
	}

	if p.ClaimedBy == nil || *p.ClaimedBy != claimantID {
		return nil, NewApiErrS(403, AetForbidden, "only the claimant can edit this place")
	}

	update := bson.M{"updated_at": time.Now().UTC()}
	if in.Name != nil {
		update["name"] = *in.Name
	}
	if in.Slug != nil {
		taken, err := isSlugTaken(ctx, *in.Slug, id)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, NewApiErr(AetBadInput, "slug already in use: %s", *in.Slug)
		}
		update["slug"] = *in.Slug
	}
	if in.CategoryID != nil {
		if _, err := uuid.Parse(*in.CategoryID); err != nil {
			return nil, NewApiErr(AetBadInput, "category_id must be a UUID")
		}
		if err := db.Categories().FindOne(ctx, bson.M{"_id": *in.CategoryID}).Err(); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, NewApiErr(AetBadInput, "category not found: %s", *in.CategoryID)
			}
			return nil, err
		}
		update["category_id"] = *in.CategoryID
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
	if in.LogoKey != nil {
		update["logo_key"] = *in.LogoKey
	}
	if in.Images != nil {
		update["images"] = *in.Images
	}
	if in.WeeklyHours != nil {
		update["weekly_hours"] = *in.WeeklyHours
	}
	_, err = db.Places().UpdateByID(ctx, id, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}

	if err := db.Places().FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
		return nil, err
	}
	return NewPlaceView(p), nil
}

func SetPlaceStatus(ctx context.Context, in SetPlaceStatusInput) error {
	if !in.Status.IsValid() {
		return NewApiErr(AetBadInput, "invalid status: %s", in.Status)
	}
	res, err := db.Places().UpdateByID(ctx, in.PlaceID, bson.M{"$set": bson.M{
		"status":     in.Status,
		"admin_id":   in.AdminID,
		"updated_at": time.Now().UTC(),
	}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return NewApiErrS(404, AetNotFound, "place not found: %s", in.PlaceID)
	}
	return nil
}

func AdminEditPlace(ctx context.Context, id string, in AdminEditPlaceInput) error {
	update := bson.M{"updated_at": time.Now().UTC()}
	if in.Name != nil {
		update["name"] = *in.Name
	}
	if in.CategoryID != nil {
		if _, err := uuid.Parse(*in.CategoryID); err != nil {
			return NewApiErr(AetBadInput, "category_id must be a UUID")
		}
		if err := db.Categories().FindOne(ctx, bson.M{"_id": *in.CategoryID}).Err(); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return NewApiErr(AetBadInput, "category not found: %s", *in.CategoryID)
			}
			return err
		}
		update["category_id"] = *in.CategoryID
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
	if in.LogoKey != nil {
		update["logo_key"] = *in.LogoKey
	}
	if in.Images != nil {
		update["images"] = *in.Images
	}
	if in.WeeklyHours != nil {
		update["weekly_hours"] = *in.WeeklyHours
	}
	if in.Slug != nil {
		taken, err := isSlugTaken(ctx, *in.Slug, id)
		if err != nil {
			return err
		}
		if taken {
			return NewApiErr(AetBadInput, "slug already in use: %s", *in.Slug)
		}
		update["slug"] = *in.Slug
	}
	res, err := db.Places().UpdateByID(ctx, id, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return NewApiErrS(404, AetNotFound, "place not found: %s", id)
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
		return NewApiErrS(404, AetNotFound, "place not found: %s", id)
	}
	return nil
}
func ListMyPlaces(ctx context.Context, userID string, paging utils.Paging) (*Page[PlaceView], error) {
	filter := bson.M{"$or": bson.A{
		// bson.M{"created_by": userID},
		bson.M{"claimed_by": userID},
	}}
	cur, err := db.Places().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var raw []models.Place
	if err = cur.All(ctx, &raw); err != nil {
		return nil, err
	}
	total, _ := db.Places().CountDocuments(ctx, filter)
	items := make([]PlaceView, 0, len(raw))
	for _, p := range raw {
		items = append(items, *NewPlaceView(p))
	}
	return NewPage(items, paging, total), nil
}

func FindPlaceByID(ctx context.Context, id string) (*models.Place, error) {
	return findPlaceBy(ctx, bson.M{"_id": id}, id)
}

func FindPlaceBySlug(ctx context.Context, slug string) (*models.Place, error) {
	return findPlaceBy(ctx, bson.M{"slug": slug}, slug)
}

func findPlaceBy(ctx context.Context, filter bson.M, key string) (*models.Place, error) {
	var p models.Place
	err := db.Places().FindOne(ctx, filter).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewApiErrS(404, AetNotFound, "place not found: %s", key)
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

func isSlugTaken(ctx context.Context, slug, excludeID string) (bool, error) {
	err := db.Places().FindOne(ctx, bson.M{"slug": slug, "_id": bson.M{"$ne": excludeID}}).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return err == nil, err
}

func CoalesceStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
