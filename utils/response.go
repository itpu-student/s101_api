package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Err(c *gin.Context, status int, code, msg string) {
	c.AbortWithStatusJSON(status, ErrorResponse{Error: code, Message: msg})
}

func BadRequest(c *gin.Context, msg string)   { Err(c, http.StatusBadRequest, "bad_request", msg) }
func Unauthorized(c *gin.Context, msg string) { Err(c, http.StatusUnauthorized, "unauthorized", msg) }
func Forbidden(c *gin.Context, msg string)    { Err(c, http.StatusForbidden, "forbidden", msg) }
func NotFound(c *gin.Context, msg string)     { Err(c, http.StatusNotFound, "not_found", msg) }
func Conflict(c *gin.Context, msg string)     { Err(c, http.StatusConflict, "conflict", msg) }
func Internal(c *gin.Context, msg string)     { Err(c, http.StatusInternalServerError, "internal_error", msg) }
