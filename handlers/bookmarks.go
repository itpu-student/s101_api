package handlers

import (
	"errors"

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
	if err != nil {
		utils.Internal(c, "bookmarks list failed")
		return
	}
	utils.OK(c, view)
}

// POST /api/bookmarks/:placeId
func AddBookmark(c *gin.Context) {
	u := middleware.CurrentUser(c)
	placeID := c.Param("placeId")

	b, already, err := services.AddBookmark(c.Request.Context(), u.ID, placeID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			utils.NotFound(c, "place not found")
		default:
			utils.Internal(c, "bookmark insert failed")
		}
		return
	}

	if already {
		utils.OK(c, services.BookmarkAlreadyAck{Ok: true, Already: true})
		return
	}

	utils.Created(c, b)
}

// DELETE /api/bookmarks/:placeId
func RemoveBookmark(c *gin.Context) {
	u := middleware.CurrentUser(c)
	placeID := c.Param("placeId")

	if err := services.RemoveBookmark(c.Request.Context(), u.ID, placeID); err != nil {
		utils.Internal(c, "bookmark delete failed")
		return
	}
	utils.NoContent(c)
}
