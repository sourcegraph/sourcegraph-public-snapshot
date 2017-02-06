package accesscontrol

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

// Mocks are the currently mocked permissions checking functions. If
// any fields are defined, those will take precedence over the actual
// permissions checking functions. This should only be used in tests.
var Mocks MockPerms

// MockPerms provides stubs for mocking permissions checking functions.
type MockPerms struct {
	VerifyUserHasReadAccess     func(ctx context.Context, method string, repoID int32) error
	VerifyActorHasRepoURIAccess func(ctx context.Context, actor *auth.Actor, method string, repoURI string) bool
	VerifyUserHasReadAccessAll  func(ctx context.Context, method string, repos []*sourcegraph.Repo) (allowed []*sourcegraph.Repo, err error)
}
