package backend

import (
	"context"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func TestOrgs_List(t *testing.T) {
	ctx := testContext()

	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*sourcegraph.User, error) { return nil, nil }

	_, err := Orgs.List(ctx)
	if err == nil {
		t.Errorf("Non-admin can access endpoint")
	}
}
