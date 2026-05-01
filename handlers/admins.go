package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// @Summary      Admin: list admins
// @Tags         admin-admins
// @Security     BearerAuth
// @Produce      json
// @Param        page  query int false "Page number"
// @Param        limit query int false "Page size"
// @Success      200 {object} services.Page[models.Admin]
// @Router       /admin/admins [get]
func AdminListAdmins(c *gin.Context) {
	paging := utils.ParsePaging(c)
	page, err := services.ListAdmins(c.Request.Context(), paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// @Summary      Admin: create admin
// @Tags         admin-admins
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body services.CreateAdminInput true "Admin data"
// @Success      201 {object} models.Admin
// @Failure      400 {object} api_err.ApiErr
// @Router       /admin/admins [post]
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

// @Summary      Admin: get admin by ID
// @Tags         admin-admins
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Admin UUID"
// @Success      200 {object} models.Admin
// @Failure      404 {object} api_err.ApiErr
// @Router       /admin/admins/{id} [get]
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

// @Summary      Admin: edit admin
// @Tags         admin-admins
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Admin UUID"
// @Param        body body services.EditAdminInput true "Fields to update"
// @Success      200 {object} models.Admin
// @Failure      400 {object} api_err.ApiErr
// @Router       /admin/admins/{id} [put]
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
