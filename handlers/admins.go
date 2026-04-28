package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// GET /api/admin/admins
func AdminListAdmins(c *gin.Context) {
	paging := utils.ParsePaging(c)
	page, err := services.ListAdmins(c.Request.Context(), paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// POST /api/admin/admins
func AdminCreateAdmin(c *gin.Context) {
	var in services.CreateAdminInput
	if bindHasErr(c, &in) {
		return
	}
	creator := middleware.CurrentAdmin(c)
	a, err := services.CreateAdmin(c.Request.Context(), in, creator)
	if hasErr(c, err) {
		return
	}
	utils.Created(c, a)
}

// GET /api/admin/admins/:id
func AdminGetAdmin(c *gin.Context) {
	id, ok := requireUUIDParam(c, "id")
	if !ok {
		return
	}
	a, err := services.GetAdmin(c.Request.Context(), id)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, a)
}

// PUT /api/admin/admins/:id
func AdminEditAdmin(c *gin.Context) {
	id, ok := requireUUIDParam(c, "id")
	if !ok {
		return
	}
	var in services.EditAdminInput
	if bindHasErr(c, &in) {
		return
	}
	editor := middleware.CurrentAdmin(c)
	a, err := services.EditAdmin(c.Request.Context(), id, in, editor)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, a)
}
