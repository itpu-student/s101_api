package services

import "errors"

// Sentinel errors returned by service functions. Handlers map these to HTTP
// statuses; other callers (e.g. bot, cron jobs) can match them with errors.Is.
var (
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
	ErrBadInput  = errors.New("bad input")
	ErrForbidden          = errors.New("forbidden")
	ErrAlreadyClaimed     = errors.New("already claimed")
	ErrPendingClaimExists = errors.New("pending claim exists")
)
