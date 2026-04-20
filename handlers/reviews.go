package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// GET /api/places/:id/reviews?all=true
// By default only latest=true reviews. ?all=true returns full history.
func ListPlaceReviews(c *gin.Context) {
	paging := utils.ParsePaging(c)
	all := c.Query("all") == "true"

	// We still need the place ID if a slug was provided for the filter
	p, err := services.FindPlaceByIDOrSlug(c.Request.Context(), c.Param("id"))
	if hasErr(c, err) {
		return
	}

	page, err := services.ListPlaceReviews(c.Request.Context(), p.ID, all, paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// POST /api/places/:id/reviews
func CreateReview(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in services.CreateReviewInput
	if err := c.ShouldBindJSON(&in); err != nil {
		hasErr(c, services.NewApiErr("bad_input", "%s", err.Error()))
		return
	}

	r, err := services.CreateReview(c.Request.Context(), u.ID, c.Param("id"), in)
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
