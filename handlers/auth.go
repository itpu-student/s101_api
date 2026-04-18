package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// POST /api/auth/verify-code
// Body: { "code": "123456" }
func VerifyCode(c *gin.Context) {
	var in services.VerifyCodeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, "invalid body")
		return
	}
	out, err := services.VerifyCode(c.Request.Context(), in)
	switch {
	case errors.Is(err, services.ErrBadInput):
		utils.BadRequest(c, "code must be 6 digits")
	case errors.Is(err, services.ErrNotFound):
		utils.Unauthorized(c, "invalid or expired code")
	case errors.Is(err, services.ErrForbidden):
		utils.Forbidden(c, "account is blocked")
	case err != nil:
		utils.Internal(c, "verify failed")
	default:
		utils.OK(c, out)
	}
}

// GET /api/auth/me
func Me(c *gin.Context) {
	u := middleware.CurrentUser(c)
	out, err := services.GetMe(c.Request.Context(), u)
	if err != nil {
		utils.Internal(c, "me lookup failed")
		return
	}
	utils.OK(c, out)
}
