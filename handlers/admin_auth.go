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
	if err := c.ShouldBindJSON(&in); err != nil {
		hasErr(c, services.NewApiErr("bad_input", "%s", err.Error()))
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
