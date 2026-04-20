package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// POST /api/files/upload — multipart: file + usage.
func UploadFile(c *gin.Context) {
	u := middleware.CurrentUser(c)
	usage := c.PostForm("usage")
	fh, err := c.FormFile("file")
	if err != nil {
		hasErr(c, services.NewApiErr("bad_input", "file is required"))
		return
	}

	f, err := services.UploadFile(c.Request.Context(), u.ID, usage, fh)
	if hasErr(c, err) {
		return
	}

	key := f.ID + "." + f.Ext
	utils.Created(c, gin.H{
		"file_id": f.ID,
		"key":     key,
		"url":     "/static/" + key,
		"usage":   f.Usage,
	})
}
