package graphqlbackend

import (
	"context"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestUsers_Activity(t *testing.T) {
	ctx := context.Background()
	store.Mocks.Users.MockGetByAuth0ID_Return(t, &sourcegraph.User{}, nil)
	u := &userResolver{user: &sourcegraph.User{}, actor: actor.FromContext(ctx)}
	_, err := u.Activity(ctx)
	if err == nil {
		t.Errorf("Non-admin can access endpoint")
	}
}
