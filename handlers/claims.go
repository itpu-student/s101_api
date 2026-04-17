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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// POST /api/claims
// Body: { place_id, phone, note? }
func SubmitClaim(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in struct {
		PlaceID string `json:"place_id" binding:"required"`
		Phone   string `json:"phone" binding:"required"`
		Note    string `json:"note"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var p models.Place
	if err := db.Places().FindOne(ctx, bson.M{"_id": in.PlaceID}).Decode(&p); err != nil {
		utils.NotFound(c, "place not found")
		return
	}
	if p.ClaimedBy != nil {
		utils.Conflict(c, "this place is already claimed")
		return
	}
	// Reject duplicate pending claims by the same user.
	existing := db.ClaimRequests().FindOne(ctx, bson.M{
		"place_id": p.ID,
		"user_id":  u.ID,
		"status":   models.StatusPending,
	})
	if existing.Err() == nil {
		utils.Conflict(c, "you already have a pending claim for this place")
		return
	}

	now := time.Now().UTC()
	cr := models.ClaimRequest{
		ID:        utils.NewUUIDv7(),
		PlaceID:   p.ID,
		UserID:    u.ID,
		Phone:     in.Phone,
		Note:      in.Note,
		Status:    models.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := db.ClaimRequests().InsertOne(ctx, cr); err != nil {
		utils.Internal(c, "claim insert failed")
		return
	}
	utils.Created(c, cr)
}

// GET /api/claims/mine
func MyClaims(c *gin.Context) {
	u := middleware.CurrentUser(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	cur, err := db.ClaimRequests().Find(ctx, bson.M{"user_id": u.ID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		utils.Internal(c, "claim list failed")
		return
	}
	var out []models.ClaimRequest
	if err := cur.All(ctx, &out); err != nil {
		utils.Internal(c, "claim decode failed")
		return
	}
	utils.OK(c, out)
}
