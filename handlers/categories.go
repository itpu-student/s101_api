package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// GET /api/categories
func ListCategories(c *gin.Context) {
	cats, err := services.ListCategories(c.Request.Context())
	if hasErr(c, err) {
		return
	}
	utils.OK(c, cats)
}
