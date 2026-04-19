package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// GET /api/places?query=&sort=top|recent|nearest&category=<slug-or-id>&near=lat,lon&page=
func ListPlaces(c *gin.Context) {
	paging := utils.ParsePaging(c)

	filter := services.PlaceFilter{}
	if cat := c.Query("category"); cat != "" {
		filter.Category = &cat
	}
	if q := strings.TrimSpace(c.Query("query")); q != "" {
		filter.Query = &q
	}
	sort := c.DefaultQuery("sort", "top")
	filter.Sort = &sort

	if lat, lon, ok := parseNear(c.Query("near")); ok {
		filter.NearLat = &lat
		filter.NearLon = &lon
	} else if sort == "nearest" {
		utils.BadRequest(c, "sort=nearest requires near=lat,lon")
		return
	}

	page, err := services.ListPlaces(c.Request.Context(), filter, paging)
	if err != nil {
		utils.Internal(c, "place list failed")
		return
	}
	utils.OK(c, page)
}

// GET /api/places/:id   (accepts UUID or slug)
func GetPlace(c *gin.Context) {
	var viewerID *string
	var viewerTyp *string
	if claims := middleware.CurrentClaims(c); claims != nil {
		viewerID = &claims.Subject
		t := string(claims.Typ)
		viewerTyp = &t
	}

	view, err := services.GetPlaceView(c.Request.Context(), c.Param("id"), viewerID, viewerTyp)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			utils.NotFound(c, "place not found")
			return
		}
		utils.Internal(c, "place lookup failed")
		return
	}
	utils.OK(c, view)
}

// POST /api/places/create
func CreatePlace(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in services.CreatePlaceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	p, err := services.CreatePlace(c.Request.Context(), u.ID, in)
	if err != nil {
		if errors.Is(err, services.ErrBadInput) {
			utils.BadRequest(c, "invalid category")
			return
		}
		utils.Internal(c, "place creation failed")
		return
	}
	utils.Created(c, p)
}

// PUT /api/places/:id   — only the claimant may edit
func EditPlace(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in services.EditPlaceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	view, err := services.EditPlace(c.Request.Context(), u.ID, c.Param("id"), in)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			utils.NotFound(c, "place not found")
		case errors.Is(err, services.ErrForbidden):
			utils.Forbidden(c, "only the claimant can edit this place")
		default:
			utils.Internal(c, "place update failed")
		}
		return
	}
	utils.OK(c, view)
}

// ---- helpers ----

func parseNear(s string) (lat, lon float64, ok bool) {
	if s == "" {
		return 0, 0, false
	}
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, 0, false
	}
	var err error
	if lat, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err != nil {
		return 0, 0, false
	}
	if lon, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err != nil {
		return 0, 0, false
	}
	return lat, lon, true
}
