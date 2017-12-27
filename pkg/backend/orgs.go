package backend

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

var Orgs = &orgs{}

type orgs struct{}

func (s *orgs) List(ctx context.Context) (res []*sourcegraph.Org, err error) {
	if Mocks.Orgs.List != nil {
		return Mocks.Orgs.List(ctx)
	}

	// 🚨 SECURITY:  only admins are allowed to use this endpoint
	if err := CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	return localstore.Orgs.List(ctx)
}
