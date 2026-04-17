package utils

import "github.com/google/uuid"

// NewUUID returns a standard v4 UUID string.
func NewUUID() string {
	return uuid.NewString()
}

// NewUUIDv7 returns a UUIDv7: a standard UUID whose leading 48 bits are the
// Unix millisecond timestamp, so ids sort by insertion time like an ObjectId.
func NewUUIDv7() string {
	return uuid.Must(uuid.NewV7()).String()
}
