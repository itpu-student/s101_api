package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// GET /api/bookmarks — returns the user's bookmarked places.
func ListBookmarks(c *gin.Context) {
	u := middleware.CurrentUser(c)
	paging := utils.ParsePaging(c)

	view, err := services.ListBookmarks(c.Request.Context(), u.ID, paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, view)
}

// POST /api/bookmarks/:placeId
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

// DELETE /api/bookmarks/:placeId
func RemoveBookmark(c *gin.Context) {
	u := middleware.CurrentUser(c)
	placeID := c.Param("placeId")

	if hasErr(c, services.RemoveBookmark(c.Request.Context(), u.ID, placeID)) {
		return
	}
	utils.NoContent(c)
}
