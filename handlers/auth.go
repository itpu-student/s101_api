package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// @Summary      Verify OTP code and get token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body services.VerifyCodeInput true "OTP code"
// @Success      200 {object} services.VerifyCodeOutput
// @Failure      400 {object} api_err.ApiErr
// @Router       /auth/verify-code [post]
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

// @Summary      Get current user profile
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} services.MeView
// @Failure      401 {object} api_err.ApiErr
// @Router       /auth/me [get]
func Me(c *gin.Context) {
	u := middleware.CurrentUser(c)
	out, err := services.GetMe(c.Request.Context(), u)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, out)
}
