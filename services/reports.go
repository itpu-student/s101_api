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

func SubmitReport(ctx context.Context, reporterID string, in SubmitReportInput) (*models.Report, error) {
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
	// Rule: at least one of (Type, Text) must be set.
	if in.Type == nil && in.Text == "" {
		return nil, NewApiErr(AetBadInput, "either type or text is required")
	}
	if len(in.Text) > config.Cfg.TextInputLimit {
		return nil, NewApiErrS(400, AetTextTooLong, "text exceeds %d chars", config.Cfg.TextInputLimit)
	}

	// Load target to verify it exists, and to denormalize reported_user_id.
	reportedUserID, err := loadTargetOwner(ctx, in.TargetType, in.TargetID)
	if err != nil {
		return nil, err
	}

	// Reject duplicate pending report from same user against same target.
	existing := db.Reports().FindOne(ctx, bson.M{
		"user_id":     reporterID,
		"target_type": in.TargetType,
		"target_id":   in.TargetID,
		"status":      models.ReportStatusPending,
	})
	if existing.Err() == nil {
		return nil, NewApiErrS(409, AetDuplicateOpenReport,
			"you already have an open report for this target")
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
	return &r, nil
}

func EditMyReport(ctx context.Context, reporterID, reportID string, in EditReportInput) (*models.Report, error) {
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

	// Compute resulting state to enforce "at least one of (Type, Text)".
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
	return &r, nil
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

func ListMyReports(ctx context.Context, userID string, paging utils.Paging) (*Page[models.Report], error) {
	filter := bson.M{"user_id": userID}
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
	return NewPage(items, paging, total), nil
}

func ListReportsAdmin(ctx context.Context, f ReportFilter, paging utils.Paging) (*Page[models.Report], error) {
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
	return NewPage(items, paging, total), nil
}

func ReviewReport(ctx context.Context, reportID, adminID string, in ReviewReportInput) (*models.Report, error) {
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
			// AdminDeleteReview returns ApiErr; bubble up unless target already gone.
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
	return &r, nil
}

// loadTargetOwner verifies the report target exists and returns the owner id
// to denormalize as reported_user_id. Owner may be nil for anonymous reviews
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
	var text string
	resp := ""
	if r.AdminResponse != nil {
		resp = *r.AdminResponse
	}
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
