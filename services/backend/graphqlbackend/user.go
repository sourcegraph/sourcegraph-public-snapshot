package graphqlbackend

import (
	"context"
	"errors"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

type currentUserResolver struct{}

func currentUser(ctx context.Context) (*currentUserResolver, error) {
	actor := auth.ActorFromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}
	return &currentUserResolver{}, nil
}

func (r *currentUserResolver) GitHubOrgs(ctx context.Context) ([]string, error) {
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
