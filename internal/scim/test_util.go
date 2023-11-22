package scim

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

// createDummyRequest creates a dummy request with a body that is not empty.
func createDummyRequest() *http.Request {
	return &http.Request{Body: io.NopCloser(strings.NewReader("test"))}
}

// makeEmail creates a new UserEmail with the given parameters.
func makeEmail(userID int32, address string, primary, verified bool) *database.UserEmail {
	var vDate *time.Time
	if verified {
		vDate = &verifiedDate
	}
	return &database.UserEmail{UserID: userID, Email: address, VerifiedAt: vDate, Primary: primary}
}
