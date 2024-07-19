package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// organizationInvitationResolver implements the GraphQL type OrganizationInvitation.
type organizationInvitationResolver struct {
	db database.DB
	v  *database.OrgInvitation
}

func NewOrganizationInvitationResolver(db database.DB, v *database.OrgInvitation) *organizationInvitationResolver {
	return &organizationInvitationResolver{db, v}
}

func orgInvitationByID(ctx context.Context, db database.DB, id graphql.ID) (*organizationInvitationResolver, error) {
	orgInvitationID, err := UnmarshalOrgInvitationID(id)
	if err != nil {
		return nil, err
	}
	return orgInvitationByIDInt64(ctx, db, orgInvitationID)
}

func orgInvitationByIDInt64(ctx context.Context, db database.DB, id int64) (*organizationInvitationResolver, error) {
	orgInvitation, err := db.OrgInvitations().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &organizationInvitationResolver{db: db, v: orgInvitation}, nil
}

func (r *organizationInvitationResolver) ID() graphql.ID {
	return MarshalOrgInvitationID(r.v.ID)
}

func MarshalOrgInvitationID(id int64) graphql.ID { return relay.MarshalID("OrgInvitation", id) }

func UnmarshalOrgInvitationID(id graphql.ID) (orgInvitationID int64, err error) {
	err = relay.UnmarshalSpec(id, &orgInvitationID)
	return
}

func (r *organizationInvitationResolver) Organization(ctx context.Context) (*OrgResolver, error) {
	return OrgByIDInt32(ctx, r.db, r.v.OrgID)
}

func (r *organizationInvitationResolver) Sender(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.db, r.v.SenderUserID)
}

func (r *organizationInvitationResolver) Recipient(ctx context.Context) (*UserResolver, error) {
	if r.v.RecipientUserID == 0 {
		return nil, nil
	}
	return UserByIDInt32(ctx, r.db, r.v.RecipientUserID)
}
func (r *organizationInvitationResolver) RecipientEmail() (*string, error) {
	if r.v.RecipientEmail == "" {
		return nil, nil
	}
	return &r.v.RecipientEmail, nil
}
func (r *organizationInvitationResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.v.CreatedAt}
}
func (r *organizationInvitationResolver) NotifiedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.v.NotifiedAt)
}

func (r *organizationInvitationResolver) RespondedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.v.RespondedAt)
}

func (r *organizationInvitationResolver) ResponseType() *string {
	if r.v.ResponseType == nil {
		return nil
	}
	if *r.v.ResponseType {
		return strptr("ACCEPT")
	}
	return strptr("REJECT")
}

func (r *organizationInvitationResolver) RespondURL(ctx context.Context) (*string, error) {
	if r.v.Pending() {
		var url string
		var err error
		if orgInvitationConfigDefined() {
			url, err = orgInvitationURL(*r.v, true)
		} else { // TODO: remove this fallback once signing key is enforced for on-prem instances
			org, err := r.db.Orgs().GetByID(ctx, r.v.OrgID)
			if err != nil {
				return nil, err
			}
			url = orgInvitationURLLegacy(org, true)
		}
		if err != nil {
			return nil, err
		}
		return &url, nil
	}
	return nil, nil
}

func (r *organizationInvitationResolver) RevokedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.v.RevokedAt)
}

func (r *organizationInvitationResolver) ExpiresAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.v.ExpiresAt)
}

func (r *organizationInvitationResolver) IsVerifiedEmail() *bool {
	return &r.v.IsVerifiedEmail
}

func strptr(s string) *string { return &s }
