package testutil

import (
	"testing"
	"time"

	"golang.org/x/net/context"

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

func EnsureUserExists(t *testing.T, ctx context.Context, login string) int {
	cl, _ := sourcegraph.NewClientFromContext(ctx)

	user, err := cl.Users.Get(ctx, &sourcegraph.UserSpec{Login: login})
	if err != nil {
		t.Fatal(err)
	}
	return int(user.UID)
}

func WaitForUserEmailToExist(t *testing.T, ctx context.Context, login string, wantEmail string) {
	cl, _ := sourcegraph.NewClientFromContext(ctx)

	d := time.Second * 10
	timeout := time.After(d)
	errc := make(chan error)
	go func() {
		for {
			emails, err := cl.Users.ListEmails(ctx, &sourcegraph.UserSpec{Login: login})
			if err != nil {
				errc <- err
				break
			}
			for _, email := range emails.EmailAddrs {
				if email.Email == wantEmail {
					errc <- nil
					break
				}
			}
		}
	}()
	select {
	case err := <-errc:
		if err != nil {
			t.Fatalf("while waiting for user %q email %q to exist: %s", login, wantEmail, err)
		}
	case <-timeout:
		t.Fatalf("user %q email %q does not exist, even after waiting %s", login, wantEmail, d)
		panic("unreachable")
	}
}
