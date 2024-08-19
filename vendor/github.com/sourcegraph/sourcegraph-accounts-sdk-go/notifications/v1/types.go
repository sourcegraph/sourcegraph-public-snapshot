package v1

import "go.opentelemetry.io/otel"

// ⚠️ WARNING: These types MUST match the SAMS implementation, at
// backend/internal/notification/types.go

const (
	nameUserDeleted = "UserDeleted"
)

// UserDeletedData contains information of a "UserDeleted" notification.
type UserDeletedData struct {
	// AccountID is the SAMS external ID of the deleted user.
	AccountID string `json:"account_id"`
	// Email is the email address of the deleted user.
	Email string `json:"email"`
}

var tracer = otel.Tracer("sams.notifications.v1")
