package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// @Summary      Submit a claim for a place
// @Tags         claims
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body services.SubmitClaimInput true "Claim data"
// @Success      201 {object} services.ClaimView
// @Failure      400 {object} api_err.ApiErr
// @Router       /claims [post]
func SubmitClaim(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in services.SubmitClaimInput
	if bindHasErr(c, &in) {
		return
	}

	cr, err := services.SubmitClaim(c.Request.Context(), u.ID, in)
	if hasErr(c, err) {
		return
	}
	utils.Created(c, cr)
}

// @Summary      List own claims
// @Tags         claims
// @Security     BearerAuth
// @Produce      json
// @Param        page  query int false "Page number"
// @Param        limit query int false "Page size"
// @Success      200 {object} services.Page[models.ClaimRequest]
// @Failure      401 {object} api_err.ApiErr
// @Router       /claims/mine [get]
func MyClaims(c *gin.Context) {
	u := middleware.CurrentUser(c)
	paging := utils.ParsePaging(c)

	page, err := services.ListClaimsForUser(c.Request.Context(), u.ID, paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}
