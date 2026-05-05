package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
	. "github.com/itpu-student/s101_api/utils/api_err"
)

// requireUUIDParam enforces that the named path param is a UUID. Writes target
// a stable resource — slugs are for reads only.
func requireUUIDParam(c *gin.Context, name string) (string, bool) {
	v := c.Param(name)
	if _, err := uuid.Parse(v); err != nil {
		hasErr(c, NewApiErr(AetBadInput, "%s must be a UUID", name))
		return "", false
	}
	return v, true
}

// ----- places -----

// @Summary      Admin: list places
// @Tags         admin-places
// @Security     BearerAuth
// @Produce      json
// @Param        status query string false "pending|approved|rejected"
// @Param        page   query int    false "Page number"
// @Param        limit  query int    false "Page size"
// @Success      200 {object} services.Page[services.PlaceView]
// @Router       /admin/places [get]
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

// @Summary      Admin: set place status
// @Tags         admin-places
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Place UUID"
// @Param        body body services.SetPlaceStatusInput true "New status"
// @Success      200 {object} services.Ok
// @Failure      400 {object} api_err.ApiErr
// @Router       /admin/places/{id}/status [put]
func AdminSetPlaceStatus(c *gin.Context) {
	a := middleware.CurrentAdmin(c)

	id, ok := requireUUIDParam(c, "id")
	if !ok {
		return
	}
	var in services.SetPlaceStatusInput
	if bindHasErr(c, &in) {
		return
	}

	in.PlaceID = id
	in.AdminID = a.ID
	if hasErr(c, services.SetPlaceStatus(c.Request.Context(), in)) {
		return
	}
	utils.OK(c, services.Ok{Ok: true})
}

// @Summary      Admin: edit place
// @Tags         admin-places
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Place UUID"
// @Param        body body services.AdminEditPlaceInput true "Fields to update"
// @Success      200 {object} services.Ok
// @Failure      400 {object} api_err.ApiErr
// @Router       /admin/places/{id} [put]
func AdminEditPlace(c *gin.Context) {
	id, ok := requireUUIDParam(c, "id")
	if !ok {
		return
	}
	var in services.AdminEditPlaceInput
	if bindHasErr(c, &in) {
		return
	}
	if hasErr(c, services.AdminEditPlace(c.Request.Context(), id, in)) {
		return
	}
	utils.OK(c, services.Ok{Ok: true})
}

// @Summary      Admin: delete place
// @Tags         admin-places
// @Security     BearerAuth
// @Param        id path string true "Place UUID"
// @Success      204
// @Failure      400 {object} api_err.ApiErr
// @Router       /admin/places/{id} [delete]
func AdminDeletePlace(c *gin.Context) {
	id, ok := requireUUIDParam(c, "id")
	if !ok {
		return
	}
	if hasErr(c, services.DeletePlaceCascade(c.Request.Context(), id)) {
		return
	}
	utils.NoContent(c)
}

// ----- reviews -----

// @Summary      Admin: list reviews
// @Tags         admin-reviews
// @Security     BearerAuth
// @Produce      json
// @Param        place_id query string false "Filter by place ID"
// @Param        page     query int    false "Page number"
// @Param        limit    query int    false "Page size"
// @Success      200 {object} services.Page[services.ReviewView]
// @Router       /admin/reviews [get]
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

// @Summary      Admin: delete review
// @Tags         admin-reviews
// @Security     BearerAuth
// @Param        id path string true "Review ID"
// @Success      204
// @Failure      404 {object} api_err.ApiErr
// @Router       /admin/reviews/{id} [delete]
func AdminDeleteReview(c *gin.Context) {
	if hasErr(c, services.AdminDeleteReview(c.Request.Context(), c.Param("id"))) {
		return
	}
	utils.NoContent(c)
}

// ----- users -----

// @Summary      Admin: list users
// @Tags         admin-users
// @Security     BearerAuth
// @Produce      json
// @Param        page  query int false "Page number"
// @Param        limit query int false "Page size"
// @Success      200 {object} services.Page[services.AdminUserView]
// @Router       /admin/users [get]
func AdminListUsers(c *gin.Context) {
	page, err := services.ListUsersAdmin(c.Request.Context(), utils.ParsePaging(c))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// @Summary      Admin: get user by ID
// @Tags         admin-users
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} services.AdminUserDetailView
// @Failure      404 {object} api_err.ApiErr
// @Router       /admin/users/{id} [get]
func AdminGetUser(c *gin.Context) {
	view, err := services.GetUserAdmin(c.Request.Context(), c.Param("id"))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, view)
}

// @Summary      Admin: block/unblock user
// @Tags         admin-users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "User ID"
// @Param        body body services.BlockUserInput true "Block flag"
// @Success      200 {object} services.Ok
// @Failure      400 {object} api_err.ApiErr
// @Router       /admin/users/{id}/block [put]
func AdminBlockUser(c *gin.Context) {
	var in services.BlockUserInput
	if bindHasErr(c, &in) {
		return
	}
	if hasErr(c, services.SetUserBlocked(c.Request.Context(), c.Param("id"), in.Blocked)) {
		return
	}
	utils.OK(c, services.Ok{Ok: true, Blocked: &in.Blocked})
}

// ----- claims -----

// @Summary      Admin: list claims
// @Tags         admin-claims
// @Security     BearerAuth
// @Produce      json
// @Param        status query string false "pending|approved|rejected"
// @Param        page   query int    false "Page number"
// @Param        limit  query int    false "Page size"
// @Success      200 {object} services.Page[services.ClaimView]
// @Router       /admin/claims [get]
func AdminListClaims(c *gin.Context) {
	page, err := services.ListClaimsAdmin(c.Request.Context(), services.ClaimFilter{
		Status: parseStatusQuery(c.Query("status")),
	}, utils.ParsePaging(c))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// @Summary      Admin: approve or reject a claim
// @Tags         admin-claims
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Claim ID"
// @Param        body body services.ReviewClaimInput true "Decision"
// @Success      200 {object} services.Ok
// @Failure      400 {object} api_err.ApiErr
// @Router       /admin/claims/{id} [put]
func AdminReviewClaim(c *gin.Context) {
	a := middleware.CurrentAdmin(c)
	var in services.ReviewClaimInput
	if bindHasErr(c, &in) {
		return
	}
	if hasErr(c, services.ReviewClaim(c.Request.Context(), c.Param("id"), in.Status, a.ID)) {
		return
	}
	utils.OK(c, services.Ok{Ok: true, Status: &in.Status})
}

// ----- categories -----

// @Summary      Admin: list categories
// @Tags         admin-categories
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} models.Category
// @Router       /admin/categories [get]
func AdminListCategories(c *gin.Context) {
	cats, err := services.ListCategories(c.Request.Context())
	if hasErr(c, err) {
		return
	}
	utils.OK(c, cats)
}

// @Summary      Admin: edit category
// @Tags         admin-categories
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Category UUID"
// @Param        body body services.EditCategoryInput true "Fields to update"
// @Success      200 {object} services.Ok
// @Failure      400 {object} api_err.ApiErr
// @Router       /admin/categories/{id} [put]
func AdminEditCategory(c *gin.Context) {
	id, ok := requireUUIDParam(c, "id")
	if !ok {
		return
	}
	var in services.EditCategoryInput
	if bindHasErr(c, &in) {
		return
	}
	if hasErr(c, services.EditCategory(c.Request.Context(), id, in)) {
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
