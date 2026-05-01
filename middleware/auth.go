package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	. "github.com/itpu-student/s101_api/utils/api_err"
	"github.com/itpu-student/s101_api/utils"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	CTX_KEY_USER   = "claims"
	CTX_KEY_ADMIN  = "user"
	CTX_KEY_CLAIMS = "admin"
)

func abortErr(c *gin.Context, e *ApiErr) {
	c.AbortWithStatusJSON(e.Status(), e)
}

// extract the bearer token from the Authorization header.
func bearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}

// OptionalAuth parses the JWT if present but does not require it.
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := bearer(c)
		if tok == "" {
			c.Next()
			return
		}
		claims, err := utils.ParseJWT(tok)
		if err != nil {
			c.Next()
			return
		}
		c.Set(CTX_KEY_USER, claims)
		c.Next()
	}
}

// RequireUser ensures the request carries a valid typ:user JWT, loads the
// user, and rejects blocked users.
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := bearer(c)
		if tok == "" {
			abortErr(c, NewApiErrS(401, AetUnauthorized, "missing token"))
			return
		}
		claims, err := utils.ParseJWT(tok)
		if err != nil {
			abortErr(c, NewApiErrS(401, AetUnauthorized, "invalid token"))
			return
		}
		if claims.Typ != utils.TypUser {
			abortErr(c, NewApiErrS(403, AetForbidden, "user token required"))
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		var u models.User
		if err := db.Users().FindOne(ctx, bson.M{"_id": claims.Subject}).Decode(&u); err != nil {
			abortErr(c, NewApiErrS(401, AetUnauthorized, "user not found"))
			return
		}
		if u.Blocked {
			abortErr(c, NewApiErrS(403, AetForbidden, "account is blocked"))
			return
		}
		c.Set(CTX_KEY_CLAIMS, claims)
		c.Set(CTX_KEY_USER, &u)
		c.Next()
	}
}

// RequireAdmin ensures the request carries a valid typ:admin JWT.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := bearer(c)
		if tok == "" {
			abortErr(c, NewApiErrS(401, AetUnauthorized, "missing token"))
			return
		}
		claims, err := utils.ParseJWT(tok)
		if err != nil {
			abortErr(c, NewApiErrS(401, AetUnauthorized, "invalid token"))
			return
		}
		if claims.Typ != utils.TypAdmin {
			abortErr(c, NewApiErrS(403, AetForbidden, "admin token required"))
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		var a models.Admin
		if err := db.Admins().FindOne(ctx, bson.M{"_id": claims.Subject}).Decode(&a); err != nil {
			abortErr(c, NewApiErrS(401, AetUnauthorized, "admin not found"))
			return
		}
		c.Set(CTX_KEY_CLAIMS, claims)
		c.Set(CTX_KEY_ADMIN, &a)
		c.Next()
	}
}

// RequireWritePrivilege rejects admins with power == 0 (read-only).
// Must be used after RequireAdmin.
func RequireWritePrivilege() gin.HandlerFunc {
	return func(c *gin.Context) {
		a := CurrentAdmin(c)
		if a == nil || a.Power == 0 {
			abortErr(c, NewApiErrS(403, AetForbidden, "read-only admin"))
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) *models.User {
	v, ok := c.Get(CTX_KEY_USER)
	if !ok {
		return nil
	}
	u, _ := v.(*models.User)
	return u
}

func CurrentAdmin(c *gin.Context) *models.Admin {
	v, ok := c.Get(CTX_KEY_ADMIN)
	if !ok {
		return nil
	}
	a, _ := v.(*models.Admin)
	return a
}

func CurrentClaims(c *gin.Context) *utils.Claims {
	v, ok := c.Get(CTX_KEY_CLAIMS)
	if !ok {
		return nil
	}
	cl, _ := v.(*utils.Claims)
	return cl
}
