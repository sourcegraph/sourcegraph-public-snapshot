package auth

import (
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CheckUnauthenticatedVisitorAccess returns nil if unauthenticated clients are allowed to access
// the API, and a non-nil error otherwise.
//
// Only Sourcegraph.com allows unauthenticated clients to access the API right now.
func CheckUnauthenticatedVisitorAccess() error {
	if dotcom.SourcegraphDotComMode() {
		return nil
	}

	// 🚨 SECURITY: Unauthenticated clients should have no access to the API (except on
	// Sourcegraph.com). While other protections (such as HTTP middleware) *should* prevent
	// unauthenticated clients from even reaching the API, this is an extra check just in case there
	// is a mistake elsewhere.
	return ErrNoUnauthenticatedVisitorAccess
}

var ErrNoUnauthenticatedVisitorAccess = errors.New("unauthenticated visitors are forbidden")
