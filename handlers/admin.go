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

// ----- places -----

// GET /api/admin/places?status=0|10|-10
func AdminListPlaces(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	paging := utils.ParsePaging(c)
	filter := bson.M{}
	if s := c.Query("status"); s != "" {
		// allow numeric values only; ignore otherwise
		switch s {
		case "0":
			filter["status"] = 0
		case "10":
			filter["status"] = 10
		case "-10":
			filter["status"] = -10
		}
	}
	cur, err := db.Places().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		utils.Internal(c, "place list failed")
		return
	}
	var out []models.Place
	if err := cur.All(ctx, &out); err != nil {
		utils.Internal(c, "place decode failed")
		return
	}
	total, _ := db.Places().CountDocuments(ctx, filter)
	utils.OK(c, gin.H{"items": out, "page": paging.Page, "limit": paging.Limit, "total": total})
}

// PUT /api/admin/places/:id/status   { status: 0 | 10 | -10 }
func AdminSetPlaceStatus(c *gin.Context) {
	var in struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if in.Status != models.StatusPending && in.Status != models.StatusApproved && in.Status != models.StatusRejected {
		utils.BadRequest(c, "status must be 0, 10, or -10")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	res, err := db.Places().UpdateByID(ctx, c.Param("id"), bson.M{"$set": bson.M{
		"status":     in.Status,
		"updated_at": time.Now().UTC(),
	}})
	if err != nil {
		utils.Internal(c, "status update failed")
		return
	}
	if res.MatchedCount == 0 {
		utils.NotFound(c, "place not found")
		return
	}
	utils.OK(c, gin.H{"ok": true})
}

// PUT /api/admin/places/:id   — admin can edit arbitrary fields.
func AdminEditPlace(c *gin.Context) {
	var in struct {
		Name        *string             `json:"name"`
		CategoryID  *string             `json:"category_id"`
		Address     *models.I18nText    `json:"address"`
		Phone       *string             `json:"phone"`
		Description *models.I18nText    `json:"description"`
		Lat         *float64            `json:"lat"`
		Lon         *float64            `json:"lon"`
		Images      *[]string           `json:"images"`
		WeeklyHours *models.WeeklyHours `json:"weekly_hours"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	update := bson.M{"updated_at": time.Now().UTC()}
	if in.Name != nil {
		update["name"] = *in.Name
	}
	if in.CategoryID != nil {
		id, ok := resolveCategoryID(ctx, *in.CategoryID)
		if !ok {
			utils.BadRequest(c, "invalid category")
			return
		}
		update["category_id"] = id
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
	res, err := db.Places().UpdateByID(ctx, c.Param("id"), bson.M{"$set": update})
	if err != nil {
		utils.Internal(c, "place update failed")
		return
	}
	if res.MatchedCount == 0 {
		utils.NotFound(c, "place not found")
		return
	}
	utils.OK(c, gin.H{"ok": true})
}

// DELETE /api/admin/places/:id
func AdminDeletePlace(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	id := c.Param("id")
	_, _ = db.Reviews().DeleteMany(ctx, bson.M{"place_id": id})
	_, _ = db.Bookmarks().DeleteMany(ctx, bson.M{"place_id": id})
	_, _ = db.ClaimRequests().DeleteMany(ctx, bson.M{"place_id": id})
	res, err := db.Places().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		utils.Internal(c, "place delete failed")
		return
	}
	if res.DeletedCount == 0 {
		utils.NotFound(c, "place not found")
		return
	}
	utils.NoContent(c)
}

// ----- reviews -----

func AdminListReviews(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	paging := utils.ParsePaging(c)
	filter := bson.M{}
	if pid := c.Query("place_id"); pid != "" {
		filter["place_id"] = pid
	}
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

func AdminDeleteReview(c *gin.Context) {
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
	if _, err := db.Reviews().DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		utils.Internal(c, "review delete failed")
		return
	}
	// Preserve "latest" invariant if we deleted the latest for (place, user).
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
	_ = RecalcPlaceRating(ctx, r.PlaceID)
	utils.NoContent(c)
}

// ----- users -----

func AdminListUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	paging := utils.ParsePaging(c)
	cur, err := db.Users().Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		utils.Internal(c, "user list failed")
		return
	}
	var out []models.User
	if err := cur.All(ctx, &out); err != nil {
		utils.Internal(c, "user decode failed")
		return
	}
	total, _ := db.Users().CountDocuments(ctx, bson.M{})
	// Admins may see the phone; keep the full struct. JSON tag `-` on User
	// hides phone, so serialize manually.
	items := make([]gin.H, 0, len(out))
	for _, u := range out {
		items = append(items, gin.H{
			"id":          u.ID,
			"name":        u.Name,
			"username":    u.Username,
			"telegram_id": u.TelegramID,
			"phone":       u.Phone,
			"avatar_url":  u.AvatarURL,
			"blocked":     u.Blocked,
			"created_at":  u.CreatedAt,
		})
	}
	utils.OK(c, gin.H{"items": items, "page": paging.Page, "limit": paging.Limit, "total": total})
}

// PUT /api/admin/users/:id/block  { blocked: true|false }
func AdminBlockUser(c *gin.Context) {
	var in struct {
		Blocked bool `json:"blocked"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	res, err := db.Users().UpdateByID(ctx, c.Param("id"), bson.M{"$set": bson.M{
		"blocked":    in.Blocked,
		"updated_at": time.Now().UTC(),
	}})
	if err != nil {
		utils.Internal(c, "user update failed")
		return
	}
	if res.MatchedCount == 0 {
		utils.NotFound(c, "user not found")
		return
	}
	utils.OK(c, gin.H{"ok": true, "blocked": in.Blocked})
}

// ----- claims -----

func AdminListClaims(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	paging := utils.ParsePaging(c)
	filter := bson.M{}
	if s := c.Query("status"); s != "" {
		switch s {
		case "0":
			filter["status"] = 0
		case "10":
			filter["status"] = 10
		case "-10":
			filter["status"] = -10
		}
	}
	cur, err := db.ClaimRequests().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		utils.Internal(c, "claim list failed")
		return
	}
	var out []models.ClaimRequest
	if err := cur.All(ctx, &out); err != nil {
		utils.Internal(c, "claim decode failed")
		return
	}
	total, _ := db.ClaimRequests().CountDocuments(ctx, filter)
	utils.OK(c, gin.H{"items": out, "page": paging.Page, "limit": paging.Limit, "total": total})
}

// PUT /api/admin/claims/:id   { status: 10 | -10 }
// On approval, set place.claimed_by to claim.user_id.
func AdminReviewClaim(c *gin.Context) {
	a := middleware.CurrentAdmin(c)
	var in struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if in.Status != models.StatusApproved && in.Status != models.StatusRejected {
		utils.BadRequest(c, "status must be 10 or -10")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var cr models.ClaimRequest
	if err := db.ClaimRequests().FindOne(ctx, bson.M{"_id": c.Param("id")}).Decode(&cr); err != nil {
		utils.NotFound(c, "claim not found")
		return
	}

	if in.Status == models.StatusApproved {
		// Refuse if the place is already claimed.
		var p models.Place
		if err := db.Places().FindOne(ctx, bson.M{"_id": cr.PlaceID}).Decode(&p); err != nil {
			utils.NotFound(c, "place not found")
			return
		}
		if p.ClaimedBy != nil && *p.ClaimedBy != cr.UserID {
			utils.Conflict(c, "place already claimed by another user")
			return
		}
		if _, err := db.Places().UpdateByID(ctx, cr.PlaceID, bson.M{"$set": bson.M{
			"claimed_by": cr.UserID,
			"updated_at": time.Now().UTC(),
		}}); err != nil {
			utils.Internal(c, "place update failed")
			return
		}
	}

	if _, err := db.ClaimRequests().UpdateByID(ctx, cr.ID, bson.M{"$set": bson.M{
		"status":      in.Status,
		"reviewed_by": a.ID,
		"updated_at":  time.Now().UTC(),
	}}); err != nil {
		utils.Internal(c, "claim update failed")
		return
	}
	utils.OK(c, gin.H{"ok": true, "status": in.Status})
}

// ----- categories -----

func AdminListCategories(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	cur, err := db.Categories().Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "slug", Value: 1}}))
	if err != nil {
		utils.Internal(c, "category list failed")
		return
	}
	var out []models.Category
	if err := cur.All(ctx, &out); err != nil {
		utils.Internal(c, "category decode failed")
		return
	}
	utils.OK(c, out)
}

// PUT /api/admin/categories/:id  { name: {en,uz}, desc: {en,uz} }
// Slug is immutable.
func AdminEditCategory(c *gin.Context) {
	var in struct {
		Name *models.I18nText `json:"name"`
		Desc *models.I18nText `json:"desc"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	update := bson.M{"updated_at": time.Now().UTC()}
	if in.Name != nil {
		update["name"] = *in.Name
	}
	if in.Desc != nil {
		update["desc"] = *in.Desc
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	res, err := db.Categories().UpdateByID(ctx, c.Param("id"), bson.M{"$set": update})
	if err != nil {
		utils.Internal(c, "category update failed")
		return
	}
	if res.MatchedCount == 0 {
		utils.NotFound(c, "category not found")
		return
	}
	utils.OK(c, gin.H{"ok": true})
}
