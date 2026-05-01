package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itpu-student/s101_api/middleware"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/services"
	"github.com/itpu-student/s101_api/utils"
)

// @Summary      Get report type metadata
// @Tags         reports
// @Produce      json
// @Success      200 {object} services.ReportMeta
// @Router       /reports/meta [get]
func ReportMeta(c *gin.Context) {
	utils.OK(c, services.GetReportMeta())
}

// @Summary      Submit a report
// @Tags         reports
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body services.SubmitReportInput true "Report data"
// @Success      201 {object} services.ReportView
// @Failure      400 {object} api_err.ApiErr
// @Router       /reports [post]
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

// @Summary      Edit own report
// @Tags         reports
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Report ID"
// @Param        body body services.EditReportInput true "Fields to update"
// @Success      200 {object} services.ReportView
// @Failure      403 {object} api_err.ApiErr
// @Router       /reports/{id} [put]
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

// @Summary      Delete own report
// @Tags         reports
// @Security     BearerAuth
// @Param        id path string true "Report ID"
// @Success      204
// @Failure      403 {object} api_err.ApiErr
// @Router       /reports/{id} [delete]
func DeleteMyReport(c *gin.Context) {
	u := middleware.CurrentUser(c)
	if hasErr(c, services.DeleteMyReport(c.Request.Context(), u.ID, c.Param("id"))) {
		return
	}
	utils.NoContent(c)
}

// @Summary      Get own report by ID
// @Tags         reports
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Report ID"
// @Success      200 {object} services.ReportView
// @Failure      404 {object} api_err.ApiErr
// @Router       /reports/{id} [get]
func GetMyReport(c *gin.Context) {
	u := middleware.CurrentUser(c)
	r, err := services.GetReport(c.Request.Context(), c.Param("id"), u.ID, false)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, r)
}

// @Summary      Admin: get report by ID
// @Tags         admin-reports
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Report ID"
// @Success      200 {object} services.ReportView
// @Failure      404 {object} api_err.ApiErr
// @Router       /admin/reports/{id} [get]
func AdminGetReport(c *gin.Context) {
	r, err := services.GetReport(c.Request.Context(), c.Param("id"), "", true)
	if hasErr(c, err) {
		return
	}
	utils.OK(c, r)
}

// @Summary      List own reports
// @Tags         reports
// @Security     BearerAuth
// @Produce      json
// @Param        status query string false "Filter by status"
// @Param        page   query int    false "Page number"
// @Param        limit  query int    false "Page size"
// @Success      200 {object} object
// @Failure      401 {object} api_err.ApiErr
// @Router       /reports/mine [get]
func MyReports(c *gin.Context) {
	u := middleware.CurrentUser(c)
	page, err := services.ListMyReports(c.Request.Context(), u.ID,
		parseReportStatusQuery(c.Query("status")), utils.ParsePaging(c))
	if hasErr(c, err) {
		return
	}
	utils.OK(c, page)
}

// @Summary      Admin: list reports
// @Tags         admin-reports
// @Security     BearerAuth
// @Produce      json
// @Param        status           query string false "Filter by status"
// @Param        type             query string false "Filter by type"
// @Param        target_type      query string false "Filter by target type"
// @Param        target_id        query string false "Filter by target ID"
// @Param        reported_user_id query string false "Filter by reported user"
// @Param        admin_id         query string false "Filter by admin"
// @Param        page             query int    false "Page number"
// @Param        limit            query int    false "Page size"
// @Success      200 {object} object
// @Router       /admin/reports [get]
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

// @Summary      Admin: review a report
// @Tags         admin-reports
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Report ID"
// @Param        body body services.ReviewReportInput true "Review decision"
// @Success      200 {object} services.ReportView
// @Failure      400 {object} api_err.ApiErr
// @Router       /admin/reports/{id} [put]
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
