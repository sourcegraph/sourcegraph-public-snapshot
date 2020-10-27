package campaigns

import (
	"encoding/json"
	"time"
)

type UserToken struct {
	UserID            int32
	ExternalServiceID int64
	Token             *json.RawMessage
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
