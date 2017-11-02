package graphqlbackend

import (
	"context"
	"errors"
	"time"

	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// userResolver resolves a graphql user using an optional a
// sourcegraph user and falling back to the provided actor.
// Sourcegraph users should contain a superset of the data
// stored in actors, but it is possible for a user to have an Auth0 account (and
// thus and actor) and not a User for backwards compatibility reasons we must
// support an actor-only userResolver as well.
type userResolver struct {
	user  *sourcegraph.User
	actor *actor.Actor
}

// deprecated use Auth0ID
func (r *userResolver) ID() string {
	return r.Auth0ID()
}

func (r *userResolver) Auth0ID() string {
	if r.user == nil {
		return r.actor.UID
	}
	return r.user.Auth0ID
}

func (r *userResolver) SourcegraphID() *int32 {
	if r.user == nil {
		return nil
	}
	return &r.user.ID
}

func (r *userResolver) Email() string {
	if r.user == nil {
		return r.actor.Email
	}
	return r.user.Email
}

func (r *userResolver) Username() *string {
	if r.user == nil {
		return nil
	}
	return &r.user.Username
}

func (r *userResolver) DisplayName() *string {
	if r.user == nil {
		return nil
	}
	return &r.user.DisplayName
}

func (r *userResolver) AvatarURL() *string {
	if r.user == nil {
		return &r.actor.AvatarURL
	}
	return r.user.AvatarURL
}

func (u *userResolver) CreatedAt() *string {
	if u.user == nil {
		return nil
	}
	t := u.user.CreatedAt.Format(time.RFC3339) // ISO
	return &t
}

func (u *userResolver) UpdatedAt() *string {
	if u.user == nil {
		return nil
	}
	t := u.user.CreatedAt.Format(time.RFC3339) // ISO
	return &t
}

// HasSourcegraphUser indicates whether the current user has a Sourcegraph user
// associated with their account, as opposed to only having a registered Auth0 user.
func (r *userResolver) HasSourcegraphUser() bool {
	if r.user == nil {
		return false
	}
	return true
}

func (*schemaResolver) CreateUser(ctx context.Context, args *struct {
	Username    string
	DisplayName string
	AvatarURL   *string
}) (*userResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		// For now, we are creating users from an existing Auth0 user. An active login
		// must exist to create a new Sourcegraph user.
		return nil, errors.New("no current user")
	}

	// If this step fails, we are still in a recoverable state (it is safe to execute
	// this codepath again, or the next user login will be possible with username and
	// afterwards we will add a row to the DB).
	newUser, err := store.Users.Create(actor.UID, actor.Email, args.Username, args.DisplayName, "", args.AvatarURL)
	if err != nil {
		return nil, err
	}
	return &userResolver{actor: actor, user: newUser}, nil
}

func (*schemaResolver) UpdateUser(ctx context.Context, args *struct {
	Username    *string
	DisplayName *string
	AvatarURL   *string
}) (*userResolver, error) {
	user, err := store.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	updatedUser, err := store.Users.Update(user.ID, args.Username, args.DisplayName, args.AvatarURL)
	if err != nil {
		return nil, err
	}

	return &userResolver{actor: actor.FromContext(ctx), user: updatedUser}, nil
}

func currentUser(ctx context.Context) (*userResolver, error) {
	user, err := store.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		if _, ok := err.(store.ErrUserNotFound); !ok {
			return nil, err
		}
	}
	return &userResolver{actor: actor.FromContext(ctx), user: user}, nil
}

func (r *userResolver) Orgs(ctx context.Context) ([]*orgResolver, error) {
	actor := actor.FromContext(ctx)
	orgs, err := store.Orgs.GetByUserID(ctx, actor.UID)
	if err != nil {
		return nil, err
	}
	orgResolvers := []*orgResolver{}
	for _, org := range orgs {
		orgResolvers = append(orgResolvers, &orgResolver{org})
	}
	return orgResolvers, nil
}

func (r *userResolver) OrgMemberships(ctx context.Context) ([]*orgMemberResolver, error) {
	actor := actor.FromContext(ctx)
	members, err := store.OrgMembers.GetByUserID(ctx, actor.UID)
	if err != nil {
		return nil, err
	}
	orgMemberResolvers := []*orgMemberResolver{}
	for _, member := range members {
		orgMemberResolvers = append(orgMemberResolvers, &orgMemberResolver{nil, member, nil})
	}
	return orgMemberResolvers, nil
}

func (r *userResolver) Tags(ctx context.Context) ([]*userTagResolver, error) {
	if r.user == nil {
		return nil, errors.New("Could not resolve tags on nil user")
	}
	tags, err := store.UserTags.GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	userTagResolvers := []*userTagResolver{}
	for _, tag := range tags {
		userTagResolvers = append(userTagResolvers, &userTagResolver{tag})
	}
	return userTagResolvers, nil
}
