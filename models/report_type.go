package models

type ReportTargetType string

const (
	ReportTargetReview ReportTargetType = "review"
	ReportTargetPlace  ReportTargetType = "place"
)

func ParseReportTargetType(s string) (ReportTargetType, bool) {
	switch s {
	case "review":
		return ReportTargetReview, true
	case "place":
		return ReportTargetPlace, true
	}
	return "", false
}

type ReportType string

const (
	ReportTypeSpam          ReportType = "spam"
	ReportTypeMisleading    ReportType = "misleading"
	ReportTypeInappropriate ReportType = "inappropriate"
	ReportTypeProfanity     ReportType = "profanity"
)

func ParseReportType(s string) (ReportType, bool) {
	switch s {
	case "spam":
		return ReportTypeSpam, true
	case "misleading":
		return ReportTypeMisleading, true
	case "inappropriate":
		return ReportTypeInappropriate, true
	case "profanity":
		return ReportTypeProfanity, true
	}
	return "", false
}

// Label returns the en/uz display label for this report type.
func (t ReportType) Label() I18nText {
	switch t {
	case ReportTypeSpam:
		return I18nText{EN: "Spam", UZ: "Spam"}
	case ReportTypeMisleading:
		return I18nText{EN: "Misleading", UZ: "Chalg'ituvchi"}
	case ReportTypeInappropriate:
		return I18nText{EN: "Inappropriate", UZ: "Nomaqbul"}
	case ReportTypeProfanity:
		return I18nText{EN: "Profanity", UZ: "Haqoratli so'zlar"}
	}
	return I18nText{}
}

type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusInProgress ReportStatus = "in_progress"
	ReportStatusDismissed  ReportStatus = "dismissed"
	ReportStatusActioned   ReportStatus = "actioned"
)

func ParseReportStatus(s string) (ReportStatus, bool) {
	switch s {
	case "pending":
		return ReportStatusPending, true
	case "in_progress":
		return ReportStatusInProgress, true
	case "dismissed":
		return ReportStatusDismissed, true
	case "actioned":
		return ReportStatusActioned, true
	}
	return "", false
}
