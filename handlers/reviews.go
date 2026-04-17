package handlers

import (
	"context"
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

// GET /api/places/:id/reviews?all=true
// By default only latest=true reviews. ?all=true returns full history.
func ListPlaceReviews(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	p, ok := findPlaceByIDOrSlug(ctx, c.Param("id"))
	if !ok {
		utils.NotFound(c, "place not found")
		return
	}
	filter := bson.M{"place_id": p.ID}
	if c.Query("all") != "true" {
		filter["latest"] = true
	}
	paging := utils.ParsePaging(c)
	cur, err := db.Reviews().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		utils.Internal(c, "review list failed")
		return
	}
	var out []models.Review
	if err := cur.All(ctx, &out); err != nil {
		utils.Internal(c, "review decode failed")
		return
	}
	total, _ := db.Reviews().CountDocuments(ctx, filter)
	utils.OK(c, gin.H{"items": out, "page": paging.Page, "limit": paging.Limit, "total": total})
}

// POST /api/places/:id/reviews
func CreateReview(c *gin.Context) {
	u := middleware.CurrentUser(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	p, ok := findPlaceByIDOrSlug(ctx, c.Param("id"))
	if !ok {
		utils.NotFound(c, "place not found")
		return
	}
	if p.Status != models.StatusApproved {
		utils.Forbidden(c, "cannot review a non-approved place")
		return
	}

	var in struct {
		StarRating    int      `json:"star_rating" binding:"required,min=1,max=5"`
		PriceRating   *int     `json:"price_rating" binding:"omitempty,min=1,max=5"`
		QualityRating *int     `json:"quality_rating" binding:"omitempty,min=1,max=5"`
		Text          string   `json:"text"`
		Images        []string `json:"images"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// Step 1 — demote any existing latest review for (place, user).
	// Must happen BEFORE inserting the new latest=true row so the partial
	// unique index on {place_id, user_id | latest:true} holds.
	if _, err := db.Reviews().UpdateMany(ctx,
		bson.M{"place_id": p.ID, "user_id": u.ID, "latest": true},
		bson.M{"$set": bson.M{"latest": false}},
	); err != nil {
		utils.Internal(c, "review demotion failed")
		return
	}

	userID := u.ID
	now := time.Now().UTC()
	r := models.Review{
		ID:            utils.NewUUIDv7(),
		PlaceID:       p.ID,
		UserID:        &userID,
		StarRating:    in.StarRating,
		PriceRating:   in.PriceRating,
		QualityRating: in.QualityRating,
		Text:          in.Text,
		Images:        coalesceStrings(in.Images),
		Latest:        true,
		CreatedAt:     now,
	}
	if _, err := db.Reviews().InsertOne(ctx, r); err != nil {
		utils.Internal(c, "review insert failed")
		return
	}

	if err := RecalcPlaceRating(ctx, p.ID); err != nil {
		utils.Internal(c, "rating recalc failed")
		return
	}
	utils.Created(c, r)
}

// DELETE /api/reviews/:id   — author only
func DeleteReview(c *gin.Context) {
	u := middleware.CurrentUser(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	var r models.Review
	if err := db.Reviews().FindOne(ctx, bson.M{"_id": id}).Decode(&r); err != nil {
		if err == mongo.ErrNoDocuments {
			utils.NotFound(c, "review not found")
			return
		}
		utils.Internal(c, "review lookup failed")
		return
	}
	if r.UserID == nil || *r.UserID != u.ID {
		utils.Forbidden(c, "not your review")
		return
	}
	if _, err := db.Reviews().DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		utils.Internal(c, "review delete failed")
		return
	}
	// If we just deleted a latest review, promote the previous one to latest=true.
	if r.Latest && r.UserID != nil {
		var prev models.Review
		err := db.Reviews().FindOne(ctx,
			bson.M{"place_id": r.PlaceID, "user_id": *r.UserID},
			options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}}),
		).Decode(&prev)
		if err == nil {
			_, _ = db.Reviews().UpdateByID(ctx, prev.ID, bson.M{"$set": bson.M{"latest": true}})
		}
	}
	if err := RecalcPlaceRating(ctx, r.PlaceID); err != nil {
		utils.Internal(c, "rating recalc failed")
		return
	}
	utils.NoContent(c)
}

// RecalcPlaceRating recomputes avg_rating & review_count from latest=true reviews.
func RecalcPlaceRating(ctx context.Context, placeID string) error {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"place_id": placeID, "latest": true}}},
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"avg": bson.M{"$avg": "$star_rating"},
			"cnt": bson.M{"$sum": 1},
		}}},
	}
	cur, err := db.Reviews().Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	var out struct {
		Avg float64 `bson:"avg"`
		Cnt int     `bson:"cnt"`
	}
	if cur.Next(ctx) {
		_ = cur.Decode(&out)
	}
	_, err = db.Places().UpdateByID(ctx, placeID, bson.M{"$set": bson.M{
		"avg_rating":   round1(out.Avg),
		"review_count": out.Cnt,
		"updated_at":   time.Now().UTC(),
	}})
	return err
}

func round1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}
