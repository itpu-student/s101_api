package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// @Summary      Get public user profile
// @Tags         users
// @Produce      json
// @Param        alias path string true "User ID (UUID) or username"
// @Success      200 {object} services.PublicUserView
// @Failure      404 {object} api_err.ApiErr
// @Router       /users/{alias} [get]
func GetPublicUser(c *gin.Context) {
	view, err := services.GetPublicUserView(c.Request.Context(), c.Param("alias"))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, view)
}

// @Summary      Update own profile
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body services.UpdateMeInput true "Fields to update"
// @Success      200 {object} services.MeView
// @Failure      401 {object} api_err.ApiErr
// @Router       /users/me [put]
func UpdateMe(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in services.UpdateMeInput
	if bindHasErr(c, &in) {
		return
	}

	res, err := services.UpdateMe(c.Request.Context(), u.ID, in)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, res)
}

// @Summary      Delete own account
// @Tags         users
// @Security     BearerAuth
// @Success      204
// @Failure      401 {object} api_err.ApiErr
// @Router       /users/me [delete]
func DeleteMe(c *gin.Context) {
	u := middleware.CurrentUser(c)
	if hasErr(c, services.DeleteUserCascade(c.Request.Context(), u.ID)) {
		return
	}
	utils.NoContent(c)
}

// @Summary      List reviews by a user
// @Tags         users
// @Produce      json
// @Param        id    path  string true  "User ID"
// @Param        page  query int    false "Page number"
// @Param        limit query int    false "Page size"
// @Success      200 {object} services.Page[services.ReviewView]
// @Failure      404 {object} api_err.ApiErr
// @Router       /users/{id}/reviews [get]
func UserReviews(c *gin.Context) {
	paging := utils.ParsePaging(c)
	page, err := services.ListUserReviews(c.Request.Context(), c.Param("id"), paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}
