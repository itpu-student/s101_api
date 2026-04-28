package api_err

type ApiErrTyp string

const (
	// ApiErr.Typ enum
	// Aet - Api Error Type
	AetNotFound           ApiErrTyp = "not_found"
	AetBadInput            ApiErrTyp = "bad_input"
	AetForbidden          ApiErrTyp = "forbidden"
	AetUnauthorized       ApiErrTyp = "unauthorized"
	AetAlreadyClaimed     ApiErrTyp = "already_claimed"
	AetPendingClaimExists ApiErrTyp = "pending_claim_exists"
	AetUsernameInvalid     ApiErrTyp = "username_invalid"
	AetUsernameTaken       ApiErrTyp = "username_taken"

	AetReportTargetMissing ApiErrTyp = "report_target_missing"
	AetReportLocked        ApiErrTyp = "report_locked"
	AetDuplicateOpenReport ApiErrTyp = "duplicate_open_report"
	AetTextTooLong         ApiErrTyp = "text_too_long"

	AetUnknown ApiErrTyp = "unknown"
)
