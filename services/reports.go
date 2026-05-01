package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/itpu-student/s101_api/bot"
	"github.com/itpu-student/s101_api/config"
	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
	. "github.com/itpu-student/s101_api/utils/api_err"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SubmitReport(ctx context.Context, reporterID string, in SubmitReportInput) (*ReportView, error) {
	if _, ok := models.ParseReportTargetType(string(in.TargetType)); !ok {
		return nil, NewApiErr(AetBadInput, "target_type must be 'review' or 'place'")
	}
	if in.TargetID == "" {
		return nil, NewApiErr(AetBadInput, "target_id is required")
	}
	if in.Type != nil {
		if _, ok := models.ParseReportType(string(*in.Type)); !ok {
			return nil, NewApiErr(AetBadInput, "invalid report type: %s", *in.Type)
		}
	}
	if in.Type == nil && in.Text == "" {
		return nil, NewApiErr(AetBadInput, "either type or text is required")
	}
	if len(in.Text) > config.Cfg.TextInputLimit {
		return nil, NewApiErrS(400, AetTextTooLong, "text exceeds %d chars", config.Cfg.TextInputLimit)
	}

	reportedUserID, err := loadTargetOwner(ctx, in.TargetType, in.TargetID)
	if err != nil {
		return nil, err
	}

	existing := db.Reports().FindOne(ctx, bson.M{
		"user_id":     reporterID,
		"target_type": in.TargetType,
		"target_id":   in.TargetID,
		"status":      models.ReportStatusPending,
	})
	if existing.Err() == nil {
		return nil, NewApiErrS(409, AetDuplicateOpenReport, "you already have an open report for this target")
	}

	now := time.Now().UTC()
	r := models.Report{
		ID:             utils.NewUUIDv7(),
		UserID:         reporterID,
		TargetType:     in.TargetType,
		TargetID:       in.TargetID,
		ReportedUserID: reportedUserID,
		Type:           in.Type,
		Text:           in.Text,
		Status:         models.ReportStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if _, err := db.Reports().InsertOne(ctx, r); err != nil {
		return nil, err
	}
	return buildReportView(ctx, &r, false), nil
}

func EditMyReport(ctx context.Context, reporterID, reportID string, in EditReportInput) (*ReportView, error) {
	var r models.Report
	if err := db.Reports().FindOne(ctx, bson.M{"_id": reportID}).Decode(&r); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewApiErrS(404, AetNotFound, "report not found: %s", reportID)
		}
		return nil, err
	}
	if r.UserID != reporterID {
		return nil, NewApiErrS(403, AetForbidden, "not your report")
	}
	if r.Status != models.ReportStatusPending {
		return nil, NewApiErrS(409, AetReportLocked, "report is no longer pending")
	}

	nextType := r.Type
	if in.Type != nil {
		if string(*in.Type) == "" {
			nextType = nil
		} else {
			if _, ok := models.ParseReportType(string(*in.Type)); !ok {
				return nil, NewApiErr(AetBadInput, "invalid report type: %s", *in.Type)
			}
			nextType = in.Type
		}
	}
	nextText := r.Text
	if in.Text != nil {
		nextText = *in.Text
	}
	if nextType == nil && nextText == "" {
		return nil, NewApiErr(AetBadInput, "either type or text is required")
	}
	if len(nextText) > config.Cfg.TextInputLimit {
		return nil, NewApiErrS(400, AetTextTooLong, "text exceeds %d chars", config.Cfg.TextInputLimit)
	}

	update := bson.M{"updated_at": time.Now().UTC()}
	if in.Type != nil {
		if nextType == nil {
			update["type"] = nil
		} else {
			update["type"] = *nextType
		}
	}
	if in.Text != nil {
		update["text"] = nextText
	}
	if _, err := db.Reports().UpdateByID(ctx, reportID, bson.M{"$set": update}); err != nil {
		return nil, err
	}
	if err := db.Reports().FindOne(ctx, bson.M{"_id": reportID}).Decode(&r); err != nil {
		return nil, err
	}
	return buildReportView(ctx, &r, false), nil
}

func DeleteMyReport(ctx context.Context, reporterID, reportID string) error {
	var r models.Report
	if err := db.Reports().FindOne(ctx, bson.M{"_id": reportID}).Decode(&r); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return NewApiErrS(404, AetNotFound, "report not found: %s", reportID)
		}
		return err
	}
	if r.UserID != reporterID {
		return NewApiErrS(403, AetForbidden, "not your report")
	}
	if r.Status != models.ReportStatusPending {
		return NewApiErrS(409, AetReportLocked, "only pending reports can be withdrawn")
	}
	_, err := db.Reports().DeleteOne(ctx, bson.M{"_id": reportID})
	return err
}

// GetReport returns a single report. If isAdmin is true, no ownership check
// and reporter/reported_user are populated. Otherwise the caller must be the
// reporter.
func GetReport(ctx context.Context, reportID, viewerUserID string, isAdmin bool) (*ReportView, error) {
	var r models.Report
	if err := db.Reports().FindOne(ctx, bson.M{"_id": reportID}).Decode(&r); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewApiErrS(404, AetNotFound, "report not found: %s", reportID)
		}
		return nil, err
	}
	if !isAdmin && r.UserID != viewerUserID {
		return nil, NewApiErrS(404, AetNotFound, "report not found: %s", reportID)
	}
	return buildReportView(ctx, &r, isAdmin), nil
}

func ListMyReports(ctx context.Context, userID string, status *models.ReportStatus, paging utils.Paging) (*Page[ReportView], error) {
	filter := bson.M{"user_id": userID}
	if status != nil {
		filter["status"] = *status
	}
	cur, err := db.Reports().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var items []models.Report
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}
	total, _ := db.Reports().CountDocuments(ctx, filter)
	return NewPage(buildReportViews(ctx, items, false), paging, total), nil
}

func ListReportsAdmin(ctx context.Context, f ReportFilter, paging utils.Paging) (*Page[ReportView], error) {
	filter := bson.M{}
	if f.Status != nil {
		filter["status"] = *f.Status
	}
	if f.Type != nil {
		filter["type"] = *f.Type
	}
	if f.TargetType != nil {
		filter["target_type"] = *f.TargetType
	}
	if f.TargetID != nil {
		filter["target_id"] = *f.TargetID
	}
	if f.ReportedUserID != nil {
		filter["reported_user_id"] = *f.ReportedUserID
	}
	if f.AdminID != nil {
		filter["admin_id"] = *f.AdminID
	}
	cur, err := db.Reports().Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetSkip(paging.Skip).SetLimit(int64(paging.Limit)))
	if err != nil {
		return nil, err
	}
	var items []models.Report
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}
	total, _ := db.Reports().CountDocuments(ctx, filter)
	return NewPage(buildReportViews(ctx, items, true), paging, total), nil
}

func ReviewReport(ctx context.Context, reportID, adminID string, in ReviewReportInput) (*ReportView, error) {
	if _, ok := models.ParseReportStatus(string(in.Status)); !ok {
		return nil, NewApiErr(AetBadInput, "invalid status: %s", in.Status)
	}
	if in.AdminResponse != nil && len(*in.AdminResponse) > config.Cfg.TextInputLimit {
		return nil, NewApiErrS(400, AetTextTooLong, "admin_response exceeds %d chars", config.Cfg.TextInputLimit)
	}

	var r models.Report
	if err := db.Reports().FindOne(ctx, bson.M{"_id": reportID}).Decode(&r); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewApiErrS(404, AetNotFound, "report not found: %s", reportID)
		}
		return nil, err
	}

	if in.DeleteTargetReview && r.TargetType != models.ReportTargetReview {
		return nil, NewApiErr(AetBadInput, "delete_target_review only valid when target_type is 'review'")
	}
	if in.BlockReportedUser && r.ReportedUserID == nil {
		return nil, NewApiErr(AetBadInput, "block_reported_user requires a known reported_user_id")
	}

	if in.DeleteTargetReview {
		if err := AdminDeleteReview(ctx, r.TargetID); err != nil {
			var ae *ApiErr
			if !(errors.As(err, &ae) && ae.Typ == AetNotFound) {
				return nil, err
			}
		}
	}
	if in.BlockReportedUser {
		_, err := db.Users().UpdateByID(ctx, *r.ReportedUserID, bson.M{"$set": bson.M{
			"blocked":    true,
			"updated_at": time.Now().UTC(),
		}})
		if err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	update := bson.M{
		"status":      in.Status,
		"admin_id":    adminID,
		"reviewed_at": now,
		"updated_at":  now,
	}
	if in.AdminResponse != nil {
		update["admin_response"] = *in.AdminResponse
	}
	if _, err := db.Reports().UpdateByID(ctx, reportID, bson.M{"$set": update}); err != nil {
		return nil, err
	}
	if err := db.Reports().FindOne(ctx, bson.M{"_id": reportID}).Decode(&r); err != nil {
		return nil, err
	}

	notifyReporter(ctx, &r)
	return buildReportView(ctx, &r, true), nil
}

// GetReportMeta returns the static metadata frontends need to render report
// forms — the list of report types with bilingual labels, and the shared text
// input length limit.
func GetReportMeta() *ReportMeta {
	types := []models.ReportType{
		models.ReportTypeSpam,
		models.ReportTypeMisleading,
		models.ReportTypeInappropriate,
		models.ReportTypeProfanity,
	}
	out := make([]ReportTypeMeta, 0, len(types))
	for _, t := range types {
		out = append(out, ReportTypeMeta{Value: t, Label: t.Label()})
	}
	return &ReportMeta{Types: out, TextInputLimit: config.Cfg.TextInputLimit}
}

// loadTargetOwner verifies the report target exists and returns its owner id
// (denormalized as reported_user_id). Owner may be nil for anonymous reviews
// or places without a creator.
func loadTargetOwner(ctx context.Context, tt models.ReportTargetType, id string) (*string, error) {
	switch tt {
	case models.ReportTargetReview:
		var rv models.Review
		if err := db.Reviews().FindOne(ctx, bson.M{"_id": id}).Decode(&rv); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, NewApiErrS(404, AetReportTargetMissing, "review not found: %s", id)
			}
			return nil, err
		}
		return rv.UserID, nil
	case models.ReportTargetPlace:
		var p models.Place
		if err := db.Places().FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, NewApiErrS(404, AetReportTargetMissing, "place not found: %s", id)
			}
			return nil, err
		}
		return p.CreatedBy, nil
	}
	return nil, NewApiErr(AetBadInput, "unknown target_type: %s", tt)
}

// buildReportView wraps a Report with a target preview card. When includeUsers
// is true, also embeds reporter, reported_user, and reviewing admin (admin views).
// Missing related rows (e.g. target deleted post-action) leave nested fields
// nil rather than failing.
func buildReportView(ctx context.Context, r *models.Report, includeUsers bool) *ReportView {
	v := &ReportView{Report: *r}
	v.Target = loadReportTarget(ctx, r.TargetType, r.TargetID)
	if includeUsers {
		v.ReporterUser = lookupUserMini(ctx, &r.UserID)
		v.ReportedUser = lookupUserMini(ctx, r.ReportedUserID)
		v.Admin = lookupAdminMini(ctx, r.AdminID)
	}
	return v
}

func buildReportViews(ctx context.Context, rs []models.Report, includeUsers bool) []ReportView {
	out := make([]ReportView, 0, len(rs))
	for i := range rs {
		out = append(out, *buildReportView(ctx, &rs[i], includeUsers))
	}
	return out
}

// loadReportTarget builds the uniform card for the report's target. For a
// review, the card is anchored to the place the review is on (so the user
// sees what context the review lives in) with the review text as content.
// Returns nil if the target row is gone.
func loadReportTarget(ctx context.Context, tt models.ReportTargetType, id string) *ReportTarget {
	switch tt {
	case models.ReportTargetPlace:
		var p models.Place
		if err := db.Places().FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
			return nil
		}
		var avatar *string
		if p.LogoKey != "" {
			lk := p.LogoKey
			avatar = &lk
		}
		return &ReportTarget{
			ID:        p.ID,
			Type:      models.ReportTargetPlace,
			Name:      p.Name,
			AvatarKey: avatar,
			Content:   pickI18n(p.Description),
		}
	case models.ReportTargetReview:
		var rv models.Review
		if err := db.Reviews().FindOne(ctx, bson.M{"_id": id}).Decode(&rv); err != nil {
			return nil
		}
		t := &ReportTarget{
			ID:      rv.ID,
			Type:    models.ReportTargetReview,
			Content: rv.Text,
		}
		var p models.Place
		if err := db.Places().FindOne(ctx, bson.M{"_id": rv.PlaceID}).Decode(&p); err == nil {
			t.Name = p.Name
			if p.LogoKey != "" {
				lk := p.LogoKey
				t.AvatarKey = &lk
			}
		}
		return t
	}
	return nil
}

// pickI18n returns the uz field, falling back to en. Used for content snippets
// where we collapse a bilingual field into a single string.
func pickI18n(t models.I18nText) string {
	if t.UZ != "" {
		return t.UZ
	}
	return t.EN
}

// lookupPublicUser returns the public projection for a user id, or nil if the
// id is nil or the user no longer exists.
func lookupPublicUser(ctx context.Context, id *string) *models.PublicUser {
	if id == nil || *id == "" {
		return nil
	}
	var u models.User
	if err := db.Users().FindOne(ctx, bson.M{"_id": *id}).Decode(&u); err != nil {
		return nil
	}
	return u.Public()
}

func lookupUserMini(ctx context.Context, id *string) *models.UserMini {
	if id == nil || *id == "" {
		return nil
	}
	var u models.User
	if err := db.Users().FindOne(ctx, bson.M{"_id": *id}).Decode(&u); err != nil {
		return nil
	}
	return u.Mini()
}

func lookupAdminMini(ctx context.Context, id *string) *models.AdminMini {
	if id == nil || *id == "" {
		return nil
	}
	var a models.Admin
	if err := db.Admins().FindOne(ctx, bson.M{"_id": *id}).Decode(&a); err != nil {
		return nil
	}
	return a.Mini()
}

// notifyReporter best-effort sends an Uz Telegram DM to the reporter when an
// admin transitions the report status. Failures are logged and swallowed.
func notifyReporter(ctx context.Context, r *models.Report) {
	var u models.User
	if err := db.Users().FindOne(ctx, bson.M{"_id": r.UserID}).Decode(&u); err != nil {
		return
	}
	if u.TelegramID == "" {
		return
	}
	resp := ""
	if r.AdminResponse != nil {
		resp = *r.AdminResponse
	}
	var text string
	switch r.Status {
	case models.ReportStatusInProgress:
		text = "Hisobotingiz ko'rib chiqilmoqda."
	case models.ReportStatusActioned:
		text = "Hisobotingiz bo'yicha chora ko'rildi."
		if resp != "" {
			text += fmt.Sprintf("\nJavob: %s", resp)
		}
	case models.ReportStatusDismissed:
		text = "Hisobotingiz rad etildi."
		if resp != "" {
			text += fmt.Sprintf("\nJavob: %s", resp)
		}
	default:
		return
	}
	if err := bot.SendDM(ctx, u.TelegramID, text); err != nil {
		log.Printf("report tg notify failed: %v", err)
	}
}
