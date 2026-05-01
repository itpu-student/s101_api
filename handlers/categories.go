package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// @Summary      List categories
// @Tags         categories
// @Produce      json
// @Success      200 {array} models.Category
// @Router       /categories [get]
func ListCategories(c *gin.Context) {
	cats, err := services.ListCategories(c.Request.Context())
	if hasErr(c, err) {
		return
	}
	utils.OK(c, cats)
}
