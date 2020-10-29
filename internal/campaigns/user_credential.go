package campaigns

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

type UserCredential struct {
	UserID            int32
	ExternalServiceID int64
	Credential        auth.Authenticator
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
