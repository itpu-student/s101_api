package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/services"
)

func hasErr(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	var ae *services.ApiErr
	if errors.As(err, &ae) {
		c.AbortWithStatusJSON(ae.Status(), ae)
		return true
	}
	fmt.Println("unknown err:", err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, &services.ApiErr{
		Typ: "unknown", Msg: err.Error(),
	})
	return true
}
