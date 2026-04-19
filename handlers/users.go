package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// GET /api/users/:id — public profile (no phone).
func GetPublicUser(c *gin.Context) {
	view, err := services.GetPublicUserView(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			utils.NotFound(c, "user not found")
			return
		}
		utils.Internal(c, "user lookup failed")
		return
	}
	utils.OK(c, view)
}

// PUT /api/users/me
func UpdateMe(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in services.UpdateMeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	res, err := services.UpdateMe(c.Request.Context(), u.ID, in)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			utils.NotFound(c, "user not found")
			return
		}
		utils.Internal(c, "user update failed")
		return
	}
	utils.OK(c, res)
}

// DELETE /api/users/me — hard delete.
// Reviews and places stay but are orphaned (user_id / created_by -> null).
// Bookmarks and claim requests are deleted.
func DeleteMe(c *gin.Context) {
	u := middleware.CurrentUser(c)
	if err := services.DeleteUserCascade(c.Request.Context(), u.ID); err != nil {
		utils.Internal(c, "user delete failed")
		return
	}
	utils.NoContent(c)
}

// GET /api/users/:id/reviews
func UserReviews(c *gin.Context) {
	paging := utils.ParsePaging(c)
	page, err := services.ListUserReviews(c.Request.Context(), c.Param("id"), paging)
	if err != nil {
		utils.Internal(c, "user reviews failed")
		return
	}
	utils.OK(c, page)
}
