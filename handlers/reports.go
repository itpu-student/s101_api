package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// GET /api/reports/meta — public. Lists report types + char limit so the
// frontend can render the form without hardcoding the enum.
func ReportMeta(c *gin.Context) {
	utils.OK(c, services.GetReportMeta())
}

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

// GET /api/reports/:id — must be the reporter's own.
func GetMyReport(c *gin.Context) {
	u := middleware.CurrentUser(c)
	r, err := services.GetReport(c.Request.Context(), c.Param("id"), u.ID, false)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, r)
}

// GET /api/admin/reports/:id — admin sees any report with reporter + reported_user.
func AdminGetReport(c *gin.Context) {
	r, err := services.GetReport(c.Request.Context(), c.Param("id"), "", true)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, r)
}

// GET /api/reports/mine?status=
func MyReports(c *gin.Context) {
	u := middleware.CurrentUser(c)
	page, err := services.ListMyReports(c.Request.Context(), u.ID,
		parseReportStatusQuery(c.Query("status")), utils.ParsePaging(c))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// GET /api/admin/reports?status=&type=&target_type=&target_id=&reported_user_id=&admin_id=
func AdminListReports(c *gin.Context) {
	f := services.ReportFilter{
		Status:     parseReportStatusQuery(c.Query("status")),
		Type:       parseReportTypeQuery(c.Query("type")),
		TargetType: parseReportTargetTypeQuery(c.Query("target_type")),
	}
	if v := c.Query("target_id"); v != "" {
		f.TargetID = &v
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
