package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
	. "github.com/itpu-student/s101_api/utils/api_err"
)

// @Summary      List places
// @Tags         places
// @Produce      json
// @Param        query       query string false "Search query"
// @Param        sort        query string false "Sort: top|recent|nearest"
// @Param        category_id query string false "Category UUID"
// @Param        near        query string false "lat,lon for nearest sort"
// @Param        page        query int    false "Page number"
// @Param        limit       query int    false "Page size"
// @Success      200 {object} object
// @Router       /places [get]
func ListPlaces(c *gin.Context) {
	paging := utils.ParsePaging(c)

	filter := services.PlaceFilter{}
	if cat := c.Query("category_id"); cat != "" {
		filter.CategoryId = cat
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
		hasErr(c, NewApiErr(AetBadInput, "sort=nearest requires near=lat,lon"))
		return
	}

	page, err := services.ListPlaces(c.Request.Context(), filter, paging)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// @Summary      Get place by ID or slug
// @Tags         places
// @Produce      json
// @Param        id path string true "Place UUID or slug"
// @Success      200 {object} services.PlaceView
// @Failure      404 {object} api_err.ApiErr
// @Router       /places/{id} [get]
func GetPlace(c *gin.Context) {
	var viewerID *string
	var viewerTyp *string
	if claims := middleware.CurrentClaims(c); claims != nil {
		viewerID = &claims.Subject
		t := string(claims.Typ)
		viewerTyp = &t
	}

	view, err := services.GetPlaceView(c.Request.Context(), c.Param("id"), viewerID, viewerTyp)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, view)
}

// @Summary      Create a place
// @Tags         places
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body services.CreatePlaceInput true "Place data"
// @Success      201 {object} services.PlaceView
// @Failure      400 {object} api_err.ApiErr
// @Router       /places/create [post]
func CreatePlace(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in services.CreatePlaceInput
	if bindHasErr(c, &in) {
		return
	}

	p, err := services.CreatePlace(c.Request.Context(), u.ID, in)
	if hasErr(c, err) {
		return
	}
	utils.Created(c, p)
}

// @Summary      Edit own place (claimant only)
// @Tags         places
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Place UUID"
// @Param        body body services.EditPlaceInput true "Fields to update"
// @Success      200 {object} services.PlaceView
// @Failure      403 {object} api_err.ApiErr
// @Router       /places/{id} [put]
func EditPlace(c *gin.Context) {
	u := middleware.CurrentUser(c)
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		hasErr(c, NewApiErr(AetBadInput, "id must be a UUID"))
		return
	}
	var in services.EditPlaceInput
	if bindHasErr(c, &in) {
		return
	}

	view, err := services.EditPlace(c.Request.Context(), u.ID, id, in)
	if hasErr(c, err) {
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
