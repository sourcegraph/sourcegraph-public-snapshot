package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

type SiteCredential struct {
	ID                  int64
	ExternalServiceType string
	ExternalServiceID   string
	Credential          auth.Authenticator
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
