package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/handlers"
	"github.com/itpu-student/s101_api/middleware"
)

func Register(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// public static file serving for uploaded assets
	r.Static("/static", "./static")

	api := r.Group("/api")

	// ---- public auth ----
	api.POST("/auth/verify-code", handlers.VerifyCode)
	api.GET("/auth/me", middleware.RequireUser(), handlers.Me)

	// ---- admin auth ----
	api.POST("/admin/auth/login", handlers.AdminLogin)
	api.GET("/admin/auth/me", middleware.RequireAdmin(), handlers.AdminMe)

	// ---- users ----
	api.GET("/users/:id", handlers.GetPublicUser)
	api.GET("/users/:id/reviews", handlers.UserReviews)
	api.PUT("/users/me", middleware.RequireUser(), handlers.UpdateMe)
	api.DELETE("/users/me", middleware.RequireUser(), handlers.DeleteMe)

	// ---- categories ----
	api.GET("/categories", handlers.ListCategories)

	// ---- places ----
	api.GET("/places", handlers.ListPlaces)
	// OptionalAuth lets creators/claimants see their own non-approved places.
	api.GET("/places/:id", middleware.OptionalAuth(), handlers.GetPlace)
	api.POST("/places/create", middleware.RequireUser(), handlers.CreatePlace)
	api.PUT("/places/:id", middleware.RequireUser(), handlers.EditPlace)

	// ---- reviews ----
	api.GET("/places/:id/reviews", handlers.ListPlaceReviews)
	api.POST("/places/:id/reviews", middleware.RequireUser(), handlers.CreateReview)
	api.DELETE("/reviews/:id", middleware.RequireUser(), handlers.DeleteReview)

	// ---- bookmarks (user-private) ----
	bm := api.Group("/bookmarks", middleware.RequireUser())
	bm.GET("", handlers.ListBookmarks)
	bm.POST("/:placeId", handlers.AddBookmark)
	bm.DELETE("/:placeId", handlers.RemoveBookmark)

	// ---- claims ----
	api.POST("/claims", middleware.RequireUser(), handlers.SubmitClaim)
	api.GET("/claims/mine", middleware.RequireUser(), handlers.MyClaims)

	// ---- reports ----
	api.GET("/reports/meta", handlers.ReportMeta)
	rep := api.Group("/reports", middleware.RequireUser())
	rep.POST("", handlers.SubmitReport)
	rep.GET("/mine", handlers.MyReports)
	rep.GET("/:id", handlers.GetMyReport)
	rep.PUT("/:id", handlers.EditMyReport)
	rep.DELETE("/:id", handlers.DeleteMyReport)

	// ---- files ----
	api.POST("/files/upload", middleware.RequireUser(), handlers.UploadFile)

	// ---- admin ----
	admin := api.Group("/admin", middleware.RequireAdmin())
	rw := middleware.RequireWritePrivilege()

	admin.GET("/places", handlers.AdminListPlaces)
	admin.PUT("/places/:id/status", rw, handlers.AdminSetPlaceStatus)
	admin.PUT("/places/:id", rw, handlers.AdminEditPlace)
	admin.DELETE("/places/:id", rw, handlers.AdminDeletePlace)

	admin.GET("/reviews", handlers.AdminListReviews)
	admin.DELETE("/reviews/:id", rw, handlers.AdminDeleteReview)

	admin.GET("/users", handlers.AdminListUsers)
	admin.GET("/users/:id", handlers.AdminGetUser)
	admin.PUT("/users/:id/block", rw, handlers.AdminBlockUser)

	admin.GET("/claims", handlers.AdminListClaims)
	admin.PUT("/claims/:id", rw, handlers.AdminReviewClaim)

	admin.GET("/categories", handlers.AdminListCategories)
	admin.PUT("/categories/:id", rw, handlers.AdminEditCategory)

	admin.GET("/reports", handlers.AdminListReports)
	admin.GET("/reports/:id", handlers.AdminGetReport)
	admin.PUT("/reports/:id", rw, handlers.AdminReviewReport)

	admin.GET("/admins", handlers.AdminListAdmins)
	admin.POST("/admins", rw, handlers.AdminCreateAdmin)
	admin.GET("/admins/:id", handlers.AdminGetAdmin)
	admin.PUT("/admins/:id", rw, handlers.AdminEditAdmin)

}
