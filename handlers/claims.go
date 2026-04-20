package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// POST /api/claims
// Body: { place_id, phone, note? }
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

// GET /api/claims/mine
func MyClaims(c *gin.Context) {
	u := middleware.CurrentUser(c)
	paging := utils.ParsePaging(c)

	page, err := services.ListClaimsForUser(c.Request.Context(), u.ID, paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}
