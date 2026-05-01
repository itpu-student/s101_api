package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// @Summary      List own bookmarks
// @Tags         bookmarks
// @Security     BearerAuth
// @Produce      json
// @Param        page  query int false "Page number"
// @Param        limit query int false "Page size"
// @Success      200 {object} object
// @Failure      401 {object} api_err.ApiErr
// @Router       /bookmarks [get]
func ListBookmarks(c *gin.Context) {
	u := middleware.CurrentUser(c)
	paging := utils.ParsePaging(c)

	view, err := services.ListBookmarks(c.Request.Context(), u.ID, paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, view)
}

// @Summary      Add bookmark
// @Tags         bookmarks
// @Security     BearerAuth
// @Param        placeId path string true "Place ID"
// @Success      201 {object} services.BookmarkView
// @Success      208
// @Failure      401 {object} api_err.ApiErr
// @Router       /bookmarks/{placeId} [post]
func AddBookmark(c *gin.Context) {
	u := middleware.CurrentUser(c)
	placeID := c.Param("placeId")

	b, already, err := services.AddBookmark(c.Request.Context(), u.ID, placeID)
	if hasErr(c, err) {
		return
	}

	if already {
		c.Status(208) // http.StatusAlreadyReported
		return
	}

	utils.Created(c, b)
}

// @Summary      Remove bookmark
// @Tags         bookmarks
// @Security     BearerAuth
// @Param        placeId path string true "Place ID"
// @Success      204
// @Failure      401 {object} api_err.ApiErr
// @Router       /bookmarks/{placeId} [delete]
func RemoveBookmark(c *gin.Context) {
	u := middleware.CurrentUser(c)
	placeID := c.Param("placeId")

	if hasErr(c, services.RemoveBookmark(c.Request.Context(), u.ID, placeID)) {
		return
	}
	utils.NoContent(c)
}
