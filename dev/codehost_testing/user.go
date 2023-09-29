package codehost_scenario

import (
	"context"

	"github.com/google/go-github/v53/github"
)

type User struct {
	s    *GithubScenario
	name string
}

func (u *User) Get(ctx context.Context) (*github.User, error) {
	if u.s.IsApplied() {
		return u.get(ctx)
	}
	panic("cannot retrieve user before scenario is applied")
}

func (u *User) get(ctx context.Context) (*github.User, error) {
	return u.s.client.GetUser(ctx, u.name)
}
