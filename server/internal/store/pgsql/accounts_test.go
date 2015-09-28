package pgsql

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func mustCreateUser(ctx context.Context, t *testing.T, users ...*sourcegraph.User) []*sourcegraph.User {
	var createdUsers []*sourcegraph.User
	for _, user := range users {
		createdUser, err := (&Accounts{}).Create(ctx, user)
		if err != nil {
			t.Fatal(err)
		}
		createdUsers = append(createdUsers, createdUser)
	}
	return createdUsers
}
