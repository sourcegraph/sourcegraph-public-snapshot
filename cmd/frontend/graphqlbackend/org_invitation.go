package graphqlbackend

import (
	"context"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

// organizationInvitationResolver implements the GraphQL type OrganizationInvitation.
type organizationInvitationResolver struct {
	v *db.OrgInvitation
}

func orgInvitationByID(ctx context.Context, id graphql.ID) (*organizationInvitationResolver, error) {
	orgInvitationID, err := unmarshalOrgInvitationID(id)
	if err != nil {
		return nil, err
	}
	return orgInvitationByIDInt64(ctx, orgInvitationID)
}

func orgInvitationByIDInt64(ctx context.Context, id int64) (*organizationInvitationResolver, error) {
	orgInvitation, err := db.OrgInvitations.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &organizationInvitationResolver{v: orgInvitation}, nil
}

func (r *organizationInvitationResolver) ID() graphql.ID {
	return marshalOrgInvitationID(r.v.ID)
}

func marshalOrgInvitationID(id int64) graphql.ID { return relay.MarshalID("OrgInvitation", id) }

func unmarshalOrgInvitationID(id graphql.ID) (orgInvitationID int64, err error) {
	err = relay.UnmarshalSpec(id, &orgInvitationID)
	return
}

func (r *organizationInvitationResolver) Organization(ctx context.Context) (*OrgResolver, error) {
	return OrgByIDInt32(ctx, r.v.OrgID)
}
func (r *organizationInvitationResolver) Sender(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.v.SenderUserID)
}
func (r *organizationInvitationResolver) Recipient(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.v.RecipientUserID)
}
func (r *organizationInvitationResolver) CreatedAt() string { return r.v.CreatedAt.Format(time.RFC3339) }
func (r *organizationInvitationResolver) NotifiedAt() *string {
	if r.v.NotifiedAt == nil {
		return nil
	}
	s := r.v.NotifiedAt.Format(time.RFC3339)
	return &s
}
func (r *organizationInvitationResolver) RespondedAt() *string {
	if r.v.RespondedAt == nil {
		return nil
	}
	s := r.v.RespondedAt.Format(time.RFC3339)
	return &s
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
		org, err := db.Orgs.GetByID(ctx, r.v.OrgID)
		if err != nil {
			return nil, err
		}
		url := orgInvitationURL(org).String()
		return &url, nil
	}
	return nil, nil
}
func (r *organizationInvitationResolver) RevokedAt() *string {
	if r.v.RevokedAt == nil {
		return nil
	}
	s := r.v.RevokedAt.Format(time.RFC3339)
	return &s
}

func strptr(s string) *string { return &s }
