package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GET /api/categories
func ListCategories(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	cur, err := db.Categories().Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "slug", Value: 1}}))
	if err != nil {
		utils.Internal(c, "categories list failed")
		return
	}
	var out []models.Category
	if err := cur.All(ctx, &out); err != nil {
		utils.Internal(c, "categories decode failed")
		return
	}
	utils.OK(c, out)
}
