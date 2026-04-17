package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Paging struct {
	Page  int
	Limit int
	Skip  int64
}

func ParsePaging(c *gin.Context) Paging {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return Paging{Page: page, Limit: limit, Skip: int64((page - 1) * limit)}
}
