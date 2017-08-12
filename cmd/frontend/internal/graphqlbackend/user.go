package graphqlbackend

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

type currentUserResolver struct{}

func currentUser(ctx context.Context) (*currentUserResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}
	return &currentUserResolver{}, nil
}

func (r *currentUserResolver) GitHubInstallations(ctx context.Context) ([]*installationResolver, error) {
	ghInstalls, err := github.ListAllAccessibleInstallations(ctx)
	if err != nil {
		return nil, err
	}
	installs := make([]*installationResolver, len(ghInstalls))
	for i, v := range ghInstalls {
		installs[i] = &installationResolver{v}
	}
	return installs, nil
}
