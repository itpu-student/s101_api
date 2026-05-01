package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
	. "github.com/itpu-student/s101_api/utils/api_err"
)

// @Summary      Upload a file
// @Tags         files
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData file   true  "File to upload"
// @Param        usage formData string false "Usage hint (e.g. avatar, place_logo)"
// @Success      201 {object} object{file_id=string,key=string,url=string,usage=string}
// @Failure      400 {object} api_err.ApiErr
// @Router       /files/upload [post]
func UploadFile(c *gin.Context) {
	u := middleware.CurrentUser(c)
	usage := c.PostForm("usage")
	fh, err := c.FormFile("file")
	if err != nil {
		hasErr(c, NewApiErr(AetBadInput, "file is required"))
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
