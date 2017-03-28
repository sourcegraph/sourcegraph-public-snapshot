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

type org struct {
	org gogithub.Organization
}

func (o *org) Collaborators() int32 {
	if o.org.Collaborators == nil {
		return 0
	}
	return int32(*o.org.Collaborators)
}

func (o *org) AvatarURL() string {
	return *o.org.AvatarURL
}

func (o *org) Name() string {
	return *o.org.Login
}

func (o *org) Description() string {
	if o.org.Description == nil {
		return ""
	}
	return *o.org.Description
}

func (r *currentUserResolver) GitHubOrgs(ctx context.Context) ([]*org, error) {
	ghOrgs, _, err := github.OrgsFromContext(ctx).List("", &gogithub.ListOptions{PerPage: 100})
	orgs := make([]*org, len(ghOrgs))
	for i, v := range ghOrgs {
		orgs[i] = &org{*v}
	}
	return orgs, err
}
