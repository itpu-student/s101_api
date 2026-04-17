package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type placeInput struct {
	Name        string             `json:"name" binding:"required"`
	CategoryID  string             `json:"category_id" binding:"required"`
	Address     models.I18nText    `json:"address" binding:"required"`
	Phone       string             `json:"phone"`
	Description models.I18nText    `json:"description"`
	Lat         float64            `json:"lat" binding:"required"`
	Lon         float64            `json:"lon" binding:"required"`
	Images      []string           `json:"images"`
	WeeklyHours models.WeeklyHours `json:"weekly_hours"`
}

type placeResponse struct {
	models.Place
	IsOpen bool `json:"is_open"`
}

func withIsOpen(p models.Place) placeResponse {
	return placeResponse{Place: p, IsOpen: utils.IsOpen(p.WeeklyHours, time.Now())}
}

// GET /api/places?query=&sort=top|recent|nearest&category=<slug-or-id>&near=lat,lon&page=
func ListPlaces(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	paging := utils.ParsePaging(c)

	filter := bson.M{"status": models.StatusApproved}

	// category: accept either category_id (UUID) or slug
	if cat := c.Query("category"); cat != "" {
		if id, ok := resolveCategoryID(ctx, cat); ok {
			filter["category_id"] = id
		} else {
			utils.OK(c, gin.H{"items": []any{}, "page": paging.Page, "limit": paging.Limit, "total": 0})
			return
		}
	}

	if q := strings.TrimSpace(c.Query("query")); q != "" {
		filter["$text"] = bson.M{"$search": q}
	}

	sort := c.DefaultQuery("sort", "top")
	findOpts := options.Find().SetSkip(paging.Skip).SetLimit(int64(paging.Limit))

	// "nearest" mode supports near=lat,lon using 2dsphere. It replaces the
	// default sort with $nearSphere (implicit distance-asc ordering).
	if sort == "nearest" || c.Query("near") != "" {
		if lat, lon, ok := parseNear(c.Query("near")); ok {
			filter["location"] = bson.M{
				"$nearSphere": bson.M{
					"$geometry": bson.M{"type": "Point", "coordinates": []float64{lon, lat}},
				},
			}
			// $nearSphere already sorts; just set limit/skip.
		} else if sort == "nearest" {
			utils.BadRequest(c, "sort=nearest requires near=lat,lon")
			return
		}
	} else {
		switch sort {
		case "recent":
			findOpts.SetSort(bson.D{{Key: "created_at", Value: -1}})
		default: // "top"
			findOpts.SetSort(bson.D{{Key: "avg_rating", Value: -1}, {Key: "review_count", Value: -1}})
		}
	}

	cur, err := db.Places().Find(ctx, filter, findOpts)
	if err != nil {
		utils.Internal(c, "place list failed")
		return
	}
	var places []models.Place
	if err := cur.All(ctx, &places); err != nil {
		utils.Internal(c, "place decode failed")
		return
	}
	total, _ := db.Places().CountDocuments(ctx, filter)

	items := make([]placeResponse, 0, len(places))
	for _, p := range places {
		items = append(items, withIsOpen(p))
	}
	utils.OK(c, gin.H{"items": items, "page": paging.Page, "limit": paging.Limit, "total": total})
}

// GET /api/places/:id   (accepts UUID or slug)
func GetPlace(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	idOrSlug := c.Param("id")

	p, ok := findPlaceByIDOrSlug(ctx, idOrSlug)
	if !ok {
		utils.NotFound(c, "place not found")
		return
	}

	// Hide pending/rejected from the public unless the viewer is admin/creator/claimant.
	if p.Status != models.StatusApproved {
		claims := middleware.CurrentClaims(c)
		if claims == nil {
			utils.NotFound(c, "place not found")
			return
		}
		allowed := claims.Typ == utils.TypAdmin ||
			(p.CreatedBy != nil && *p.CreatedBy == claims.Subject) ||
			(p.ClaimedBy != nil && *p.ClaimedBy == claims.Subject)
		if !allowed {
			utils.NotFound(c, "place not found")
			return
		}
	}
	utils.OK(c, withIsOpen(p))
}

// POST /api/places/create
func CreatePlace(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in placeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// validate category exists
	catID, ok := resolveCategoryID(ctx, in.CategoryID)
	if !ok {
		utils.BadRequest(c, "invalid category")
		return
	}

	atc, err := utils.ResolveSOATOID(ctx, in.Lat, in.Lon)
	if err != nil {
		atc = "uz_unknown"
	}

	slug, err := generateUniqueSlug(ctx, in.Name)
	if err != nil {
		utils.Internal(c, "slug generation failed")
		return
	}

	now := time.Now().UTC()
	createdBy := u.ID
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
		Images:      coalesceStrings(in.Images),
		WeeklyHours: in.WeeklyHours,
		Status:      models.StatusPending,
		CreatedBy:   &createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := db.Places().InsertOne(ctx, p); err != nil {
		utils.Internal(c, "place insert failed")
		return
	}
	utils.Created(c, p)
}

// PUT /api/places/:id   — only the claimant may edit
func EditPlace(c *gin.Context) {
	u := middleware.CurrentUser(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	p, ok := findPlaceByIDOrSlug(ctx, c.Param("id"))
	if !ok {
		utils.NotFound(c, "place not found")
		return
	}
	if p.ClaimedBy == nil || *p.ClaimedBy != u.ID {
		utils.Forbidden(c, "only the claimant can edit this place")
		return
	}

	var in struct {
		Phone       *string             `json:"phone"`
		Description *models.I18nText    `json:"description"`
		WeeklyHours *models.WeeklyHours `json:"weekly_hours"`
		Images      *[]string           `json:"images"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
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
	if _, err := db.Places().UpdateByID(ctx, p.ID, bson.M{"$set": update}); err != nil {
		utils.Internal(c, "place update failed")
		return
	}
	// reload
	_ = db.Places().FindOne(ctx, bson.M{"_id": p.ID}).Decode(&p)
	utils.OK(c, withIsOpen(p))
}

// ---- helpers ----

func findPlaceByIDOrSlug(ctx context.Context, idOrSlug string) (models.Place, bool) {
	var p models.Place
	err := db.Places().FindOne(ctx, bson.M{"$or": bson.A{
		bson.M{"_id": idOrSlug},
		bson.M{"slug": idOrSlug},
	}}).Decode(&p)
	if err != nil {
		return models.Place{}, false
	}
	return p, true
}

func resolveCategoryID(ctx context.Context, val string) (string, bool) {
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

// generateUniqueSlug appends -2, -3, ... until the slug is unique.
func generateUniqueSlug(ctx context.Context, name string) (string, error) {
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

func parseNear(s string) (lat, lon float64, ok bool) {
	if s == "" {
		return 0, 0, false
	}
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, 0, false
	}
	var err error
	if lat, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err != nil {
		return 0, 0, false
	}
	if lon, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err != nil {
		return 0, 0, false
	}
	return lat, lon, true
}

func coalesceStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
