package graphqlbackend

import (
	"context"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/suspiciousnames"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/types"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func (r *schemaResolver) Organization(ctx context.Context, args struct{ Name string }) (*orgResolver, error) {
	org, err := db.Orgs.GetByName(ctx, args.Name)
	if err != nil {
		return nil, err
	}
	return &orgResolver{org: org}, nil
}

// Org is DEPRECATED (but still in use by sourcegraph/src). Use Node to look up an org by its
// graphql.ID instead.
func (r *schemaResolver) Org(ctx context.Context, args *struct {
	ID graphql.ID
}) (*orgResolver, error) {
	return orgByID(ctx, args.ID)
}

func orgByID(ctx context.Context, id graphql.ID) (*orgResolver, error) {
	orgID, err := unmarshalOrgID(id)
	if err != nil {
		return nil, err
	}
	return orgByIDInt32(ctx, orgID)
}

func orgByIDInt32(ctx context.Context, orgID int32) (*orgResolver, error) {
	org, err := db.Orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &orgResolver{org}, nil
}

type orgResolver struct {
	org *types.Org
}

func (o *orgResolver) ID() graphql.ID { return marshalOrgID(o.org.ID) }

func marshalOrgID(id int32) graphql.ID { return relay.MarshalID("Org", id) }

func unmarshalOrgID(id graphql.ID) (orgID int32, err error) {
	err = relay.UnmarshalSpec(id, &orgID)
	return
}

func (o *orgResolver) OrgID() int32 {
	return o.org.ID
}

func (o *orgResolver) Name() string {
	return o.org.Name
}

func (o *orgResolver) DisplayName() *string {
	return o.org.DisplayName
}

func (r *orgResolver) URL() string { return "/organizations/" + r.org.Name }

func (r *orgResolver) SettingsURL() string { return r.URL() + "/settings" }

func (o *orgResolver) CreatedAt() string { return o.org.CreatedAt.Format(time.RFC3339) }

func (o *orgResolver) Members(ctx context.Context) (*staticUserConnectionResolver, error) {
	// 🚨 SECURITY: Only org members can list the org members.
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err != nil {
		if err == backend.ErrNotAnOrgMember {
			return nil, errors.New("must be a member of this organization to view members")
		}
		return nil, err
	}

	memberships, err := db.OrgMembers.GetByOrgID(ctx, o.org.ID)
	if err != nil {
		return nil, err
	}
	users := make([]*types.User, len(memberships))
	for i, membership := range memberships {
		user, err := db.Users.GetByID(ctx, membership.UserID)
		if err != nil {
			return nil, err
		}
		users[i] = user
	}
	return &staticUserConnectionResolver{users: users}, nil
}

func (o *orgResolver) configurationSubject() api.ConfigurationSubject {
	return api.ConfigurationSubject{Org: &o.org.ID}
}

func (o *orgResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	// 🚨 SECURITY: Only organization members and site admins may access the settings, because they
	// may contains secrets or other sensitive data.
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err != nil {
		return nil, err
	}

	settings, err := db.Settings.GetLatest(ctx, o.configurationSubject())
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}
	return &settingsResolver{&configurationSubject{org: o}, settings, nil}, nil
}

func (o *orgResolver) ConfigurationCascade() *configurationCascadeResolver {
	return &configurationCascadeResolver{subject: &configurationSubject{org: o}}
}

func (o *orgResolver) ViewerPendingInvitation(ctx context.Context) (*organizationInvitationResolver, error) {
	if actor := actor.FromContext(ctx); actor.IsAuthenticated() {
		orgInvitation, err := db.OrgInvitations.GetPending(ctx, o.org.ID, actor.UID)
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return &organizationInvitationResolver{orgInvitation}, nil
	}
	return nil, nil
}

func (o *orgResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err == backend.ErrNotAuthenticated || err == backend.ErrNotAnOrgMember {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (o *orgResolver) ViewerIsMember(ctx context.Context) (bool, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return false, nil
	}
	if _, err := db.OrgMembers.GetByOrgIDAndUserID(ctx, o.org.ID, actor.UID); err != nil {
		if errcode.IsNotFound(err) {
			err = nil
		}
		return false, err
	}
	return true, nil
}

func (*schemaResolver) CreateOrganization(ctx context.Context, args *struct {
	Name        string
	DisplayName *string
}) (*orgResolver, error) {
	currentUser, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("no current user")
	}

	if err := suspiciousnames.CheckNameAllowedForUserOrOrganization(args.Name); err != nil {
		return nil, err
	}
	newOrg, err := db.Orgs.Create(ctx, args.Name, args.DisplayName)
	if err != nil {
		return nil, err
	}

	// Add the current user as the first member of the new org.
	_, err = db.OrgMembers.Create(ctx, newOrg.ID, currentUser.SourcegraphID())
	if err != nil {
		return nil, err
	}

	return &orgResolver{org: newOrg}, nil
}

func (*schemaResolver) UpdateOrganization(ctx context.Context, args *struct {
	ID          graphql.ID
	DisplayName *string
}) (*orgResolver, error) {
	var orgID int32
	if err := relay.UnmarshalSpec(args.ID, &orgID); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is a member
	// of the org that is being modified.
	if err := backend.CheckOrgAccess(ctx, orgID); err != nil {
		return nil, err
	}

	updatedOrg, err := db.Orgs.Update(ctx, orgID, args.DisplayName)
	if err != nil {
		return nil, err
	}

	return &orgResolver{org: updatedOrg}, nil
}

func (*schemaResolver) RemoveUserFromOrganization(ctx context.Context, args *struct {
	User         graphql.ID
	Organization graphql.ID
}) (*EmptyResponse, error) {
	orgID, err := unmarshalOrgID(args.Organization)
	if err != nil {
		return nil, err
	}
	userID, err := unmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is a member of the org that is being modified, or a
	// site admin.
	if err := backend.CheckOrgAccess(ctx, orgID); err != nil {
		return nil, err
	}

	log15.Info("removing user from org", "user", userID, "org", orgID)
	return nil, db.OrgMembers.Remove(ctx, orgID, userID)
}

func (*schemaResolver) AddUserToOrganization(ctx context.Context, args *struct {
	Organization graphql.ID
	Username     string
}) (*EmptyResponse, error) {
	var orgID int32
	if err := relay.UnmarshalSpec(args.Organization, &orgID); err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Must be a site admin to immediately add a user to an organization (bypassing the
	// invitation step).
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	userToInvite, _, err := getUserToInviteToOrganization(ctx, args.Username, orgID)
	if err != nil {
		return nil, err
	}
	if _, err := db.OrgMembers.Create(ctx, orgID, userToInvite.ID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
