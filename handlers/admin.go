package handlers

import (
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
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// PUT /api/admin/places/:id/status   { status: "pending"|"approved"|"rejected" }
func AdminSetPlaceStatus(c *gin.Context) {
	var in services.SetPlaceStatusInput
	if err := c.ShouldBindJSON(&in); err != nil {
		hasErr(c, services.NewApiErr("bad_input", "%s", err.Error()))
		return
	}
	if hasErr(c, services.SetPlaceStatus(c.Request.Context(), c.Param("id"), in.Status)) {
		return
	}
	utils.OK(c, services.Ok{Ok: true})
}

// PUT /api/admin/places/:id   — admin can edit arbitrary fields.
func AdminEditPlace(c *gin.Context) {
	var in services.AdminEditPlaceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		hasErr(c, services.NewApiErr("bad_input", "%s", err.Error()))
		return
	}
	if hasErr(c, services.AdminEditPlace(c.Request.Context(), c.Param("id"), in)) {
		return
	}
	utils.OK(c, services.Ok{Ok: true})
}

// DELETE /api/admin/places/:id
func AdminDeletePlace(c *gin.Context) {
	if hasErr(c, services.DeletePlaceCascade(c.Request.Context(), c.Param("id"))) {
		return
	}
	utils.NoContent(c)
}

// ----- reviews -----

func AdminListReviews(c *gin.Context) {
	var pid *string
	if v := c.Query("place_id"); v != "" {
		pid = &v
	}
	page, err := services.ListReviewsAdmin(c.Request.Context(),
		services.ReviewFilter{PlaceID: pid}, utils.ParsePaging(c))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

func AdminDeleteReview(c *gin.Context) {
	if hasErr(c, services.AdminDeleteReview(c.Request.Context(), c.Param("id"))) {
		return
	}
	utils.NoContent(c)
}

// ----- users -----

func AdminListUsers(c *gin.Context) {
	page, err := services.ListUsersAdmin(c.Request.Context(), utils.ParsePaging(c))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// PUT /api/admin/users/:id/block  { blocked: true|false }
func AdminBlockUser(c *gin.Context) {
	var in services.BlockUserInput
	if err := c.ShouldBindJSON(&in); err != nil {
		hasErr(c, services.NewApiErr("bad_input", "%s", err.Error()))
		return
	}
	if hasErr(c, services.SetUserBlocked(c.Request.Context(), c.Param("id"), in.Blocked)) {
		return
	}
	utils.OK(c, services.Ok{Ok: true, Blocked: &in.Blocked})
}

// ----- claims -----

func AdminListClaims(c *gin.Context) {
	page, err := services.ListClaimsAdmin(c.Request.Context(), services.ClaimFilter{
		Status: parseStatusQuery(c.Query("status")),
	}, utils.ParsePaging(c))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// PUT /api/admin/claims/:id   { status: "approved" | "rejected" }
func AdminReviewClaim(c *gin.Context) {
	a := middleware.CurrentAdmin(c)
	var in services.ReviewClaimInput
	if err := c.ShouldBindJSON(&in); err != nil {
		hasErr(c, services.NewApiErr("bad_input", "%s", err.Error()))
		return
	}
	if hasErr(c, services.ReviewClaim(c.Request.Context(), c.Param("id"), in.Status, a.ID)) {
		return
	}
	utils.OK(c, services.Ok{Ok: true, Status: &in.Status})
}

// ----- categories -----

func AdminListCategories(c *gin.Context) {
	cats, err := services.ListCategories(c.Request.Context())
	if hasErr(c, err) {
		return
	}
	utils.OK(c, cats)
}

// PUT /api/admin/categories/:id  { name: {en,uz}, desc: {en,uz} }  (slug is immutable)
func AdminEditCategory(c *gin.Context) {
	var in services.EditCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		hasErr(c, services.NewApiErr("bad_input", "%s", err.Error()))
		return
	}
	if hasErr(c, services.EditCategory(c.Request.Context(), c.Param("id"), in)) {
		return
	}
	utils.OK(c, services.Ok{Ok: true})
}

// parseStatusQuery turns the ?status= query param into a *models.Status
// suitable for a filter struct. Unknown values mean "no filter".
func parseStatusQuery(s string) *models.Status {
	if v, ok := models.ParseStatus(s); ok {
		return &v
	}
	return nil
}
