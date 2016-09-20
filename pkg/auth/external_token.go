package auth

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

var (
	// ErrNoExternalAuthToken occurs when no external auth token exists
	// for a given user and host.
	ErrNoExternalAuthToken = errors.New("no external auth token found for user and host")
)

func FetchGitHubToken(ctx context.Context, uid int) (*sourcegraph.ExternalToken, error) {
	return nil, ErrNoExternalAuthToken // TODO
}
