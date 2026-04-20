package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// POST /api/auth/verify-code
// Body: { "code": "123456" }
func VerifyCode(c *gin.Context) {
	var in services.VerifyCodeInput
	if bindHasErr(c, &in) {
		return
	}
	out, err := services.VerifyCode(c.Request.Context(), in)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, out)
}

// GET /api/auth/me
func Me(c *gin.Context) {
	u := middleware.CurrentUser(c)
	out, err := services.GetMe(c.Request.Context(), u)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, out)
}
