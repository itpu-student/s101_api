package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
	. "github.com/itpu-student/s101_api/utils/api_err"
)

// @Summary      List reviews for a place
// @Tags         reviews
// @Produce      json
// @Param        id    path  string false "Place UUID"
// @Param        all   query bool   false "Return full history (default: latest only)"
// @Param        page  query int    false "Page number"
// @Param        limit query int    false "Page size"
// @Success      200 {object} services.Page[services.ReviewView]
// @Failure      400 {object} api_err.ApiErr
// @Router       /places/{id}/reviews [get]
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

// @Summary      Create a review for a place
// @Tags         reviews
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Place UUID"
// @Param        body body services.CreateReviewInput true "Review data"
// @Success      201 {object} services.ReviewView
// @Failure      400 {object} api_err.ApiErr
// @Router       /places/{id}/reviews [post]
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

// @Summary      Get a review by ID
// @Tags         reviews
// @Produce      json
// @Param        id path string true "Review ID"
// @Success      200 {object} services.ReviewView
// @Failure      404 {object} api_err.ApiErr
// @Router       /reviews/{id} [get]
func GetReview(c *gin.Context) {
	v, err := services.GetReview(c.Request.Context(), c.Param("id"))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, v)
}

// @Summary      List previous reviews (latest=false) for the same place+user
// @Tags         reviews
// @Produce      json
// @Param        id    path  string true  "Review ID (must be latest=true)"
// @Param        page  query int    false "Page number"
// @Param        limit query int    false "Page size"
// @Success      200 {object} services.Page[services.ReviewView]
// @Failure      404 {object} api_err.ApiErr
// @Router       /reviews/prevs/{id} [get]
func GetPrevReviews(c *gin.Context) {
	paging := utils.ParsePaging(c)
	page, err := services.ListPrevReviews(c.Request.Context(), c.Param("id"), paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// @Summary      Delete own review
// @Tags         reviews
// @Security     BearerAuth
// @Param        id path string true "Review ID"
// @Success      204
// @Failure      403 {object} api_err.ApiErr
// @Router       /reviews/{id} [delete]
func DeleteReview(c *gin.Context) {
	u := middleware.CurrentUser(c)
	if hasErr(c, services.DeleteUserReview(c.Request.Context(), u.ID, c.Param("id"))) {
		return
	}
	utils.NoContent(c)
}
