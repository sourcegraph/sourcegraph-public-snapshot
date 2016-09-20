package testutil

import (
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

const password = "pw"

func CreateAccount(t *testing.T, ctx context.Context, login string) error {
	cl, _ := sourcegraph.NewClientFromContext(ctx)

	created, err := cl.Accounts.Create(ctx, &sourcegraph.NewAccount{
		Login:    login,
		Email:    login + "@example.com",
		Password: password,
	})
	if err != nil {
		return err
	}
	t.Logf("created account %q (UID %d)", login, created.UID)

	return nil
}
