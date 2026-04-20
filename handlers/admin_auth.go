package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// POST /api/admin/auth/login
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

// GET /api/admin/auth/me
func AdminMe(c *gin.Context) {
	utils.OK(c, middleware.CurrentAdmin(c))
}
