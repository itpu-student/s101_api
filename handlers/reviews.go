package handlers

import (
	"errors"

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
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			utils.NotFound(c, "place not found")
			return
		}
		utils.Internal(c, "place lookup failed")
		return
	}

	page, err := services.ListPlaceReviews(c.Request.Context(), p.ID, all, paging)
	if err != nil {
		utils.Internal(c, "review list failed")
		return
	}
	utils.OK(c, page)
}

// POST /api/places/:id/reviews
func CreateReview(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in services.CreateReviewInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	r, err := services.CreateReview(c.Request.Context(), u.ID, c.Param("id"), in)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			utils.NotFound(c, "place not found")
		case errors.Is(err, services.ErrForbidden):
			utils.Forbidden(c, "cannot review a non-approved place")
		default:
			utils.Internal(c, "review creation failed")
		}
		return
	}
	utils.Created(c, r)
}

// DELETE /api/reviews/:id   — author only
func DeleteReview(c *gin.Context) {
	u := middleware.CurrentUser(c)
	if err := services.DeleteUserReview(c.Request.Context(), u.ID, c.Param("id")); err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			utils.NotFound(c, "review not found")
		case errors.Is(err, services.ErrForbidden):
			utils.Forbidden(c, "not your review")
		default:
			utils.Internal(c, "review delete failed")
		}
		return
	}
	utils.NoContent(c)
}
