package graphqlbackend

import (
	"context"
	"errors"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

type userResolver struct {
	uid string
}

func currentUser(ctx context.Context) (*userResolver, error) {
	actor := auth.ActorFromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}
	return &userResolver{
		uid: actor.UID,
	}, nil
}

func (r *userResolver) UID() string {
	return r.uid
}

func (r *userResolver) GitHubOrgs(ctx context.Context) ([]string, error) {
	orgs, _, err := github.OrgsFromContext(ctx).List("", &gogithub.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}
	orgNames := make([]string, len(orgs))
	for i, org := range orgs {
		orgNames[i] = *org.Login
	}
	return orgNames, nil
}
