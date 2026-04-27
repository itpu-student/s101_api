package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// POST /api/reports
func SubmitReport(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in services.SubmitReportInput
	if bindHasErr(c, &in) {
		return
	}
	r, err := services.SubmitReport(c.Request.Context(), u.ID, in)
	if hasErr(c, err) {
		return
	}
	utils.Created(c, r)
}

// PUT /api/reports/:id
func EditMyReport(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var in services.EditReportInput
	if bindHasErr(c, &in) {
		return
	}
	r, err := services.EditMyReport(c.Request.Context(), u.ID, c.Param("id"), in)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, r)
}

// DELETE /api/reports/:id
func DeleteMyReport(c *gin.Context) {
	u := middleware.CurrentUser(c)
	if hasErr(c, services.DeleteMyReport(c.Request.Context(), u.ID, c.Param("id"))) {
		return
	}
	utils.NoContent(c)
}

// GET /api/reports/mine
func MyReports(c *gin.Context) {
	u := middleware.CurrentUser(c)
	page, err := services.ListMyReports(c.Request.Context(), u.ID, utils.ParsePaging(c))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// GET /api/admin/reports?status=&type=&target_type=&reported_user_id=&admin_id=
func AdminListReports(c *gin.Context) {
	f := services.ReportFilter{
		Status:     parseReportStatusQuery(c.Query("status")),
		Type:       parseReportTypeQuery(c.Query("type")),
		TargetType: parseReportTargetTypeQuery(c.Query("target_type")),
	}
	if v := c.Query("reported_user_id"); v != "" {
		f.ReportedUserID = &v
	}
	if v := c.Query("admin_id"); v != "" {
		f.AdminID = &v
	}
	page, err := services.ListReportsAdmin(c.Request.Context(), f, utils.ParsePaging(c))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// PUT /api/admin/reports/:id
func AdminReviewReport(c *gin.Context) {
	a := middleware.CurrentAdmin(c)
	var in services.ReviewReportInput
	if bindHasErr(c, &in) {
		return
	}
	r, err := services.ReviewReport(c.Request.Context(), c.Param("id"), a.ID, in)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, r)
}

func parseReportStatusQuery(s string) *models.ReportStatus {
	if v, ok := models.ParseReportStatus(s); ok {
		return &v
	}
	return nil
}

func parseReportTypeQuery(s string) *models.ReportType {
	if v, ok := models.ParseReportType(s); ok {
		return &v
	}
	return nil
}

func parseReportTargetTypeQuery(s string) *models.ReportTargetType {
	if v, ok := models.ParseReportTargetType(s); ok {
		return &v
	}
	return nil
}
