package handlers

import (
	"errors"

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
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	cr, err := services.SubmitClaim(c.Request.Context(), u.ID, in)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			utils.NotFound(c, "place not found")
		case errors.Is(err, services.ErrAlreadyClaimed):
			utils.Conflict(c, "this place is already claimed")
		case errors.Is(err, services.ErrPendingClaimExists):
			utils.Conflict(c, "you already have a pending claim for this place")
		case errors.Is(err, services.ErrBadInput):
			utils.BadRequest(c, "place_id and phone are required")
		default:
			utils.Internal(c, "claim insert failed")
		}
		return
	}
	utils.Created(c, cr)
}

// GET /api/claims/mine
func MyClaims(c *gin.Context) {
	u := middleware.CurrentUser(c)
	paging := utils.ParsePaging(c)

	page, err := services.ListClaimsForUser(c.Request.Context(), u.ID, paging)
	if err != nil {
		utils.Internal(c, "claim list failed")
		return
	}
	utils.OK(c, page)
}
