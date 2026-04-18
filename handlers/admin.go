package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// ----- places -----

// GET /api/admin/places?status=pending|approved|rejected
func AdminListPlaces(c *gin.Context) {
	paging := utils.ParsePaging(c)
	page, err := services.ListPlacesAdmin(c.Request.Context(), services.PlaceFilter{
		Status: parseStatusQuery(c.Query("status")),
	}, paging)
	if err != nil {
		utils.Internal(c, "place list failed")
		return
	}
	utils.OK(c, page)
}

// PUT /api/admin/places/:id/status   { status: "pending"|"approved"|"rejected" }
func AdminSetPlaceStatus(c *gin.Context) {
	var in services.SetPlaceStatusInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	err := services.SetPlaceStatus(c.Request.Context(), c.Param("id"), in.Status)
	switch {
	case errors.Is(err, services.ErrBadInput):
		utils.BadRequest(c, "status must be pending, approved, or rejected")
	case errors.Is(err, services.ErrNotFound):
		utils.NotFound(c, "place not found")
	case err != nil:
		utils.Internal(c, "status update failed")
	default:
		utils.OK(c, services.Ok{Ok: true})
	}
}

// PUT /api/admin/places/:id   — admin can edit arbitrary fields.
func AdminEditPlace(c *gin.Context) {
	var in services.AdminEditPlaceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	err := services.AdminEditPlace(c.Request.Context(), c.Param("id"), in)
	switch {
	case errors.Is(err, services.ErrBadInput):
		utils.BadRequest(c, "invalid category")
	case errors.Is(err, services.ErrNotFound):
		utils.NotFound(c, "place not found")
	case err != nil:
		utils.Internal(c, "place update failed")
	default:
		utils.OK(c, services.Ok{Ok: true})
	}
}

// DELETE /api/admin/places/:id
func AdminDeletePlace(c *gin.Context) {
	err := services.DeletePlaceCascade(c.Request.Context(), c.Param("id"))
	switch {
	case errors.Is(err, services.ErrNotFound):
		utils.NotFound(c, "place not found")
	case err != nil:
		utils.Internal(c, "place delete failed")
	default:
		utils.NoContent(c)
	}
}

// ----- reviews -----

func AdminListReviews(c *gin.Context) {
	var pid *string
	if v := c.Query("place_id"); v != "" {
		pid = &v
	}
	page, err := services.ListReviewsAdmin(c.Request.Context(),
		services.ReviewFilter{PlaceID: pid}, utils.ParsePaging(c))
	if err != nil {
		utils.Internal(c, "review list failed")
		return
	}
	utils.OK(c, page)
}

func AdminDeleteReview(c *gin.Context) {
	err := services.AdminDeleteReview(c.Request.Context(), c.Param("id"))
	switch {
	case errors.Is(err, services.ErrNotFound):
		utils.NotFound(c, "review not found")
	case err != nil:
		utils.Internal(c, "review delete failed")
	default:
		utils.NoContent(c)
	}
}

// ----- users -----

func AdminListUsers(c *gin.Context) {
	page, err := services.ListUsersAdmin(c.Request.Context(), utils.ParsePaging(c))
	if err != nil {
		utils.Internal(c, "user list failed")
		return
	}
	utils.OK(c, page)
}

// PUT /api/admin/users/:id/block  { blocked: true|false }
func AdminBlockUser(c *gin.Context) {
	var in services.BlockUserInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	err := services.SetUserBlocked(c.Request.Context(), c.Param("id"), in.Blocked)
	switch {
	case errors.Is(err, services.ErrNotFound):
		utils.NotFound(c, "user not found")
	case err != nil:
		utils.Internal(c, "user update failed")
	default:
		utils.OK(c, services.Ok{Ok: true, Blocked: &in.Blocked})
	}
}

// ----- claims -----

func AdminListClaims(c *gin.Context) {
	page, err := services.ListClaimsAdmin(c.Request.Context(), services.ClaimFilter{
		Status: parseStatusQuery(c.Query("status")),
	}, utils.ParsePaging(c))
	if err != nil {
		utils.Internal(c, "claim list failed")
		return
	}
	utils.OK(c, page)
}

// PUT /api/admin/claims/:id   { status: "approved" | "rejected" }
func AdminReviewClaim(c *gin.Context) {
	a := middleware.CurrentAdmin(c)
	var in services.ReviewClaimInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	err := services.ReviewClaim(c.Request.Context(), c.Param("id"), in.Status, a.ID)
	switch {
	case errors.Is(err, services.ErrBadInput):
		utils.BadRequest(c, "status must be approved or rejected")
	case errors.Is(err, services.ErrNotFound):
		utils.NotFound(c, "claim or place not found")
	case errors.Is(err, services.ErrConflict):
		utils.Conflict(c, "place already claimed by another user")
	case err != nil:
		utils.Internal(c, "claim update failed")
	default:
		utils.OK(c, services.Ok{Ok: true, Status: &in.Status})
	}
}

// ----- categories -----

func AdminListCategories(c *gin.Context) {
	cats, err := services.ListCategories(c.Request.Context())
	if err != nil {
		utils.Internal(c, "category list failed")
		return
	}
	utils.OK(c, cats)
}

// PUT /api/admin/categories/:id  { name: {en,uz}, desc: {en,uz} }  (slug is immutable)
func AdminEditCategory(c *gin.Context) {
	var in services.EditCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	err := services.EditCategory(c.Request.Context(), c.Param("id"), in)
	switch {
	case errors.Is(err, services.ErrNotFound):
		utils.NotFound(c, "category not found")
	case err != nil:
		utils.Internal(c, "category update failed")
	default:
		utils.OK(c, services.Ok{Ok: true})
	}
}

// parseStatusQuery turns the ?status= query param into a *models.Status
// suitable for a filter struct. Unknown values mean "no filter".
func parseStatusQuery(s string) *models.Status {
	if v, ok := models.ParseStatus(s); ok {
		return &v
	}
	return nil
}
