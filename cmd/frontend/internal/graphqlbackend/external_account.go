package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
)

type externalAccountResolver struct {
	user      *userResolver
	serviceID string
	accountID string
}

func externalAccountByID(ctx context.Context, id graphql.ID) (*externalAccountResolver, error) {
	externalAccountID, err := unmarshalExternalAccountID(id)
	if err != nil {
		return nil, err
	}

	userID := externalAccountID // TEMPORARY: users each have 0 or 1 external account, so just use user ID

	// 🚨 SECURITY: Only the user and site admins should be able to see a user's external accounts.
	if err := backend.CheckSiteAdminOrSameUser(ctx, userID); err != nil {
		return nil, err
	}

	user, err := userByIDInt32(ctx, userID)
	if err != nil {
		return nil, err
	}
	var externalID string
	if user.user.ExternalID != nil {
		externalID = *user.user.ExternalID
	}
	return &externalAccountResolver{
		user:      user,
		serviceID: user.user.ExternalProvider,
		accountID: externalID,
	}, nil
}

func marshalExternalAccountID(repo int32) graphql.ID { return relay.MarshalID("ExternalAccount", repo) }

func unmarshalExternalAccountID(id graphql.ID) (externalAccountID int32, err error) {
	err = relay.UnmarshalSpec(id, &externalAccountID)
	return
}

func (r *externalAccountResolver) ID() graphql.ID      { return marshalExternalAccountID(r.user.user.ID) }
func (r *externalAccountResolver) User() *userResolver { return r.user }
func (r *externalAccountResolver) ServiceType() string { return "" }
func (r *externalAccountResolver) ServiceID() string   { return r.serviceID }
func (r *externalAccountResolver) AccountID() string   { return r.accountID }
func (r *externalAccountResolver) CreatedAt() string   { return r.user.CreatedAt() }
func (r *externalAccountResolver) UpdatedAt() string   { return r.user.CreatedAt() }
