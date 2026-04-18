package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// POST /api/admin/auth/login
func AdminLogin(c *gin.Context) {
	var in services.AdminLoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, "invalid body")
		return
	}
	out, err := services.AdminLogin(c.Request.Context(), in)
	switch {
	case errors.Is(err, services.ErrBadInput):
		utils.BadRequest(c, "username and password are required")
	case errors.Is(err, services.ErrNotFound):
		utils.Unauthorized(c, "invalid credentials")
	case err != nil:
		utils.Internal(c, "login failed")
	default:
		utils.OK(c, out)
	}
}

// GET /api/admin/auth/me
func AdminMe(c *gin.Context) {
	utils.OK(c, middleware.CurrentAdmin(c))
}
