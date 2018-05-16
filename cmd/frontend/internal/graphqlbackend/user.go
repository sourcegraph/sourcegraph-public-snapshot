package graphqlbackend

import (
	"context"
	"errors"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

func (r *schemaResolver) User(ctx context.Context, args struct{ Username string }) (*userResolver, error) {
	user, err := db.Users.GetByUsername(ctx, args.Username)
	if err != nil {
		return nil, err
	}
	return &userResolver{user: user}, nil
}

// userResolver resolves a Sourcegraph user.
type userResolver struct {
	user *types.User
}

func userByID(ctx context.Context, id graphql.ID) (*userResolver, error) {
	userID, err := unmarshalUserID(id)
	if err != nil {
		return nil, err
	}
	return userByIDInt32(ctx, userID)
}

func userByIDInt32(ctx context.Context, id int32) (*userResolver, error) {
	user, err := db.Users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &userResolver{user: user}, nil
}

func (r *userResolver) ID() graphql.ID { return marshalUserID(r.user.ID) }

func marshalUserID(id int32) graphql.ID { return relay.MarshalID("User", id) }

func unmarshalUserID(id graphql.ID) (userID int32, err error) {
	err = relay.UnmarshalSpec(id, &userID)
	return
}

func (r *userResolver) SourcegraphID() int32 { return r.user.ID }

func (r *userResolver) Email(ctx context.Context) (string, error) {
	// 🚨 SECURITY: Only the user and admins are allowed to access the email address.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err != nil {
		return "", err
	}

	email, _, err := db.UserEmails.GetPrimaryEmail(ctx, r.user.ID)
	if err != nil {
		return "", err
	}

	return email, nil
}

func (r *userResolver) Username() string { return r.user.Username }

func (r *userResolver) DisplayName() *string {
	if r.user.DisplayName == "" {
		return nil
	}
	return &r.user.DisplayName
}

func (r *userResolver) AvatarURL() *string {
	if r.user.AvatarURL == "" {
		return nil
	}
	return &r.user.AvatarURL
}

func (r *userResolver) URL() string {
	return "/users/" + r.user.Username
}

func (r *userResolver) CreatedAt() string {
	return r.user.CreatedAt.Format(time.RFC3339)
}

func (r *userResolver) UpdatedAt() *string {
	t := r.user.UpdatedAt.Format(time.RFC3339) // ISO
	return &t
}

func (r *userResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	// 🚨 SECURITY: Only the user and admins are allowed to access the user's settings, because they
	// may contain secrets or other sensitive data.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err != nil {
		return nil, err
	}

	settings, err := db.Settings.GetLatest(ctx, api.ConfigurationSubject{User: &r.user.ID})
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}
	return &settingsResolver{&configurationSubject{user: r}, settings, nil}, nil
}

func (r *userResolver) SiteAdmin(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: Only the user and admins are allowed to determine if the user is a site admin.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err != nil {
		return false, err
	}

	return r.user.SiteAdmin, nil
}

func (*schemaResolver) UpdateUser(ctx context.Context, args *struct {
	User        graphql.ID
	Username    *string
	DisplayName *string
	AvatarURL   *string
}) (*EmptyResponse, error) {
	userID, err := unmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the user and site admins are allowed to update the user.
	if err := backend.CheckSiteAdminOrSameUser(ctx, userID); err != nil {
		return nil, err
	}

	update := db.UserUpdate{
		DisplayName: args.DisplayName,
		AvatarURL:   args.AvatarURL,
	}
	if args.Username != nil {
		update.Username = *args.Username
	}
	if err := db.Users.Update(ctx, userID, update); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func currentUser(ctx context.Context) (*userResolver, error) {
	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &userResolver{user: user}, nil
}

func (r *userResolver) Orgs(ctx context.Context) ([]*orgResolver, error) {
	orgs, err := db.Orgs.GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	orgResolvers := []*orgResolver{}
	for _, org := range orgs {
		orgResolvers = append(orgResolvers, &orgResolver{org})
	}
	return orgResolvers, nil
}

func (r *userResolver) Tags(ctx context.Context) ([]*userTagResolver, error) {
	// 🚨 SECURITY: Only the user and admins are allowed to access the user's tags.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err != nil {
		return nil, err
	}

	tags, err := db.UserTags.GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	userTagResolvers := []*userTagResolver{}
	for _, tag := range tags {
		userTagResolvers = append(userTagResolvers, &userTagResolver{tag})
	}
	return userTagResolvers, nil
}

func (r *userResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err == backend.ErrNotAuthenticated || err == backend.ErrMustBeSiteAdmin {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (r *schemaResolver) UpdatePassword(ctx context.Context, args *struct {
	OldPassword string
	NewPassword string
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: A user can only change their own password.
	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("no authenticated user")
	}

	if err := db.Users.UpdatePassword(ctx, user.ID, args.OldPassword, args.NewPassword); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
