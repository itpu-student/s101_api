package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.com/itpu-student/s101_api/utils/api_err"
)

func hasErr(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	var ae *ApiErr
	if errors.As(err, &ae) {
		c.AbortWithStatusJSON(ae.Status(), ae)
		return true
	}
	fmt.Println("unknown err:", err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, &ApiErr{
		Typ: AetUnknown, Msg: err.Error(),
	})
	return true
}

func bindHasErr(c *gin.Context, obj any) bool {
	if err := c.ShouldBind(obj); err != nil {
		return hasErr(c, NewApiErr(AetBadInput, "%s", err.Error()))
	}
	return false
}
