package accountsv1

import (
	"time"
)

// ⚠️ WARNING: These types MUST match the SAMS implementation, at
// backend/internal/api/model.go

type Metadata map[string]any
type MetadataSet map[string]Metadata

type User struct {
	Sub           string    `json:"sub"`            // OIDC-compliant field name, DO NOT change
	Name          string    `json:"name"`           // OIDC-compliant field name, DO NOT change
	Email         string    `json:"email"`          // OIDC-compliant field name, DO NOT change
	Picture       string    `json:"picture"`        // OIDC-compliant field name, DO NOT change
	EmailVerified bool      `json:"email_verified"` // OIDC-compliant field name, DO NOT change
	CreatedAt     time.Time `json:"created_at"`     // OIDC-compliant field name, DO NOT change

	// Metadata is a map of metadata scope to metadata JSON contents.
	Metadata MetadataSet `json:"metadata,omitempty"`
}
