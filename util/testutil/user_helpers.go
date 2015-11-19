package testutil

import (
	"testing"
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

const Password = "pw"

func CreateAccount(t *testing.T, ctx context.Context, login string) (*sourcegraph.UserSpec, error) {
	cl := sourcegraph.NewClientFromContext(ctx)

	user, err := cl.Accounts.Create(ctx, &sourcegraph.NewAccount{
		Login:    login,
		Email:    login + "@example.com",
		Password: Password,
	})
	if err != nil {
		return nil, err
	}
	t.Logf("created account %q (domain %q, UID %d)", user.Login, user.Domain, user.UID)

	return user, nil
}

func EnsureUserExists(t *testing.T, ctx context.Context, login string) int {
	cl := sourcegraph.NewClientFromContext(ctx)

	user, err := cl.Users.Get(ctx, &sourcegraph.UserSpec{Login: login})
	if err != nil {
		t.Fatal(err)
	}
	return int(user.UID)
}

func WaitForUserEmailToExist(t *testing.T, ctx context.Context, login string, wantEmail string) {
	cl := sourcegraph.NewClientFromContext(ctx)

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
