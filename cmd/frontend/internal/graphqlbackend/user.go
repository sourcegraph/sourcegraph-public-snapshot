package graphqlbackend

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"

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

func (r *currentUserResolver) OrgMemberships(ctx context.Context) ([]*orgMemberResolver, error) {
	actor := actor.FromContext(ctx)
	members, err := localstore.OrgMembers.GetByUserID(ctx, actor.UID)
	if err != nil {
		return nil, err
	}
	orgMemberResolvers := []*orgMemberResolver{}
	for _, member := range members {
		orgMemberResolvers = append(orgMemberResolvers, &orgMemberResolver{nil, member})
	}
	return orgMemberResolvers, nil
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

func (r *currentUserResolver) ID(ctx context.Context) string {
	return r.actor.UID
}

// TODO(Dan): since this will likely be a mutable property on the webapp
// frontend, don't just return the actor's value (set at sign up/sign in)
func (r *currentUserResolver) AvatarURL(ctx context.Context) *string {
	return &r.actor.AvatarURL
}

// TODO(Dan): since this will likely be a mutable property on the webapp
// frontend, don't just return the actor's value (set at sign up/sign in)
func (r *currentUserResolver) Email(ctx context.Context) *string {
	return &r.actor.Email
}
