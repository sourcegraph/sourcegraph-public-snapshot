package backend

import (
	"testing"

	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockOrgs struct {
	List func(context.Context) ([]*sourcegraph.Org, error)
}

func (s *MockOrgs) MockList(t *testing.T, wantNames ...string) (called *bool) {
	called = new(bool)
	s.List = func(ctx context.Context) ([]*sourcegraph.Org, error) {
		*called = true
		orgs := make([]*sourcegraph.Org, len(wantNames))
		for i, name := range wantNames {
			orgs[i] = &sourcegraph.Org{Name: name}
		}
		return orgs, nil
	}
	return
}
