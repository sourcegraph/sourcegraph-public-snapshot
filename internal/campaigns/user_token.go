package campaigns

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/secret"
)

type UserToken struct {
	UserID            int32
	ExternalServiceID int64
	Token             secret.StringValue
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
