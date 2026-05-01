package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// @Summary      Admin login
// @Tags         admin-auth
// @Accept       json
// @Produce      json
// @Param        body body services.AdminLoginInput true "Credentials"
// @Success      200 {object} services.AdminLoginOutput
// @Failure      401 {object} api_err.ApiErr
// @Router       /admin/auth/login [post]
func AdminLogin(c *gin.Context) {
	var in services.AdminLoginInput
	if bindHasErr(c, &in) {
		return
	}
	out, err := services.AdminLogin(c.Request.Context(), in)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, out)
}

// @Summary      Get current admin
// @Tags         admin-auth
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} models.Admin
// @Failure      401 {object} api_err.ApiErr
// @Router       /admin/auth/me [get]
func AdminMe(c *gin.Context) {
	utils.OK(c, middleware.CurrentAdmin(c))
}
