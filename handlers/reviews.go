package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
	. "github.com/itpu-student/s101_api/utils/api_err"
)

// GET /api/places/:id/reviews?all=true
// By default only latest=true reviews. ?all=true returns full history.
func ListPlaceReviews(c *gin.Context) {
	paging := utils.ParsePaging(c)
	all := c.Query("all") == "true"

	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		hasErr(c, NewApiErr(AetBadInput, "id must be a UUID"))
		return
	}
	p, err := services.FindPlaceByID(c.Request.Context(), id)

	if hasErr(c, err) {
		return
	}

	page, err := services.ListPlaceReviews(c.Request.Context(), p.ID, all, paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// POST /api/places/:id/reviews   — :id must be a UUID.
func CreateReview(c *gin.Context) {
	u := middleware.CurrentUser(c)
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		hasErr(c, NewApiErr(AetBadInput, "id must be a UUID"))
		return
	}
	var in services.CreateReviewInput
	if bindHasErr(c, &in) {
		return
	}

	r, err := services.CreateReview(c.Request.Context(), u.ID, id, in)
	if hasErr(c, err) {
		return
	}
	utils.Created(c, r)
}

// DELETE /api/reviews/:id   — author only
func DeleteReview(c *gin.Context) {
	u := middleware.CurrentUser(c)
	if hasErr(c, services.DeleteUserReview(c.Request.Context(), u.ID, c.Param("id"))) {
		return
	}
	utils.NoContent(c)
}
