package accesscontrol

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

var Mocks MockPerms

type MockPerms struct {
	VerifyUserHasReadAccess              func(ctx context.Context, method string, repoID int32) error
	VerifyUserHasWriteAccess             func(ctx context.Context, method string, repo int32) error
	VerifyActorHasRepoURIAccess          func(ctx context.Context, actor *auth.Actor, method string, repoURI string) bool
	VerifyActorHasGCPRepoAccess          func(ctx context.Context, actor *auth.Actor, repoURI string) bool
	VerifyUserHasReadAccessAll           func(ctx context.Context, method string, repos []*sourcegraph.Repo) (allowed []*sourcegraph.Repo, err error)
	VerifyUserHasReadAccessToDefRepoRefs func(ctx context.Context, method string, repoRefs []*sourcegraph.DeprecatedDefRepoRef) ([]*sourcegraph.DeprecatedDefRepoRef, error)
}
