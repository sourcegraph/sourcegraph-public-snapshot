package backend

import (
	"context"

	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

var Orgs = &orgs{}

type orgs struct{}

func (s *orgs) List(ctx context.Context) (res []*sourcegraph.Org, err error) {
	if Mocks.Orgs.List != nil {
		return Mocks.Orgs.List(ctx)
	}

	actor := actor.FromContext(ctx)
	// 🚨 SECURITY:  only admins are allowed to use this endpoint
	if !actor.IsAdmin() {
		return nil, errors.New("Must be an admin")
	}
	return localstore.Orgs.List(ctx)
}
