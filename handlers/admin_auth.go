package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// POST /api/admin/auth/login
func AdminLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid body")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var a models.Admin
	err := db.Admins().FindOne(ctx, bson.M{"username": req.Username}).Decode(&a)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.Unauthorized(c, "invalid credentials")
			return
		}
		utils.Internal(c, "admin lookup failed")
		return
	}
	if err := utils.ComparePassword(a.PasswordHash, req.Password); err != nil {
		utils.Unauthorized(c, "invalid credentials")
		return
	}
	token, err := utils.IssueJWT(a.ID, utils.TypAdmin)
	if err != nil {
		utils.Internal(c, "token issue failed")
		return
	}
	utils.OK(c, gin.H{"token": token, "admin": a})
}

// GET /api/admin/auth/me
func AdminMe(c *gin.Context) {
	a := middleware.CurrentAdmin(c)
	utils.OK(c, a)
}
