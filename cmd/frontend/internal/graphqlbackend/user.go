package graphqlbackend

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

type currentUserResolver struct {
	actor *actor.Actor
}

func currentUser(ctx context.Context) (*currentUserResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}
	return &currentUserResolver{
		actor: actor,
	}, nil
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

func (r *currentUserResolver) ID(ctx context.Context) (string, error) {
	if r.actor != nil {
		return r.actor.UID, nil
	}
	return "", errors.New("no current user")
}

func (r *currentUserResolver) Handle(ctx context.Context) (*string, error) {
	if r.actor != nil {
		return &r.actor.Login, nil
	}
	return nil, errors.New("no current user")
}

func (r *currentUserResolver) AvatarURL(ctx context.Context) (*string, error) {
	if r.actor != nil {
		return &r.actor.AvatarURL, nil
	}
	return nil, errors.New("no current user")
}

func (r *currentUserResolver) Email(ctx context.Context) (*string, error) {
	if r.actor != nil {
		return &r.actor.Email, nil
	}
	return nil, errors.New("no current user")
}
