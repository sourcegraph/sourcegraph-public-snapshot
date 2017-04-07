package graphqlbackend

import (
	"context"
	"errors"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
)

type currentUserResolver struct{}

func currentUser(ctx context.Context) (*currentUserResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}
	return &currentUserResolver{}, nil
}

type orgResolver struct {
	org gogithub.Organization
}

func (o *orgResolver) Collaborators() int32 {
	if o.org.Collaborators == nil {
		return 0
	}
	return int32(*o.org.Collaborators)
}

func (o *orgResolver) AvatarURL() string {
	return *o.org.AvatarURL
}

func (o *orgResolver) Name() string {
	return *o.org.Login
}

func (o *orgResolver) Description() string {
	if o.org.Description == nil {
		return ""
	}
	return *o.org.Description
}

func (r *currentUserResolver) GitHubOrgs(ctx context.Context) ([]*orgResolver, error) {
	ghOrgs, _, err := github.Client(ctx).Organizations.List("", &gogithub.ListOptions{PerPage: 100})
	orgs := make([]*orgResolver, len(ghOrgs))
	for i, v := range ghOrgs {
		orgs[i] = &orgResolver{*v}
	}
	return orgs, err
}
