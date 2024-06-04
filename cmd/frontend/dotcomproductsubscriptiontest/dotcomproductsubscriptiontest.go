// Package dotcomproductsubscriptiontest provides a shim for internals from
// cmd/frontend/internal/dotcom/productsubscription, to be used in tests in
// packages outside of cmd/frontend.
//
// 👷 This package is intended to be a short-lived mechanism, and should be
// removed as part of https://linear.app/sourcegraph/project/12f1d5047bd2/overview.
// It needs to exist so that we can TEST reading data from the Sourcegraph.com
// database from Enterprise Portal. No other code should import this package.
package dotcomproductsubscriptiontest

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type subscriptionsDB struct{ db database.DB }

func (s *subscriptionsDB) Create(ctx context.Context, userID int32, username string) (id string, err error) {
	return productsubscription.NewSubscriptionsDB(s.db).Create(ctx, userID, username)
}

func (s *subscriptionsDB) UpdateCodyGatewayAccess(ctx context.Context, id string, access graphqlbackend.UpdateCodyGatewayAccessInput) error {
	return productsubscription.NewSubscriptionsDB(s.db).Update(ctx, id, productsubscription.DBSubscriptionUpdate{
		CodyGatewayAccess: &access,
	})
}

func (s *subscriptionsDB) Archive(ctx context.Context, id string) error {
	return productsubscription.NewSubscriptionsDB(s.db).Archive(ctx, id)
}

// NewSubscriptionsDB returns a new SubscriptionsDB backed by the given database.DB.
// It requires testing.T to indicate that it should only be used in tests.
//
// See package docs for more details.
func NewSubscriptionsDB(t *testing.T, db database.DB) *subscriptionsDB {
	t.Helper()
	return &subscriptionsDB{db}
}

type LicensesDB interface {
	Create(ctx context.Context, subscriptionID, licenseKey string, version int, info license.Info) (id string, err error)
}

// NewLicensesDB returns a new LicensesDB backed by the given database.DB.
// It requires testing.T to indicate that it should only be used in tests.
//
// See package docs for more details.
func NewLicensesDB(t *testing.T, db database.DB) LicensesDB {
	t.Helper()
	return productsubscription.NewLicensesDB(db)
}

type TokensDB interface {
	LookupProductSubscriptionIDByAccessToken(ctx context.Context, token string) (string, error)
}

// NewTokensDB returns a new TokensDB backed by the given database.DB.
// It requires testing.T to indicate that it should only be used in tests.
//
// See package docs for more details.
func NewTokensDB(t *testing.T, db database.DB) TokensDB {
	t.Helper()
	return productsubscription.NewTokensDB(db)
}

type mockAdminFetcher struct{}

func (mockAdminFetcher) GetByID(context.Context, int32) (*types.User, error) {
	return &types.User{ID: 999, SiteAdmin: true}, nil
}

// NewCodyGatewayAccessResolver returns the resolver code for Cody Gateway
// access, solely for the purpose of asserting against existing behaviour.
//
// See package docs for more details.
func NewCodyGatewayAccessResolver(t *testing.T, db database.DB, subscriptionID string) graphqlbackend.CodyGatewayAccess {
	t.Helper()
	// Hydrate the actor with a mock site-admin
	a := &actor.Actor{UID: 999}
	_, err := a.User(context.Background(), mockAdminFetcher{})
	require.NoError(t, err)
	// Hydrate the context with the site-admin actor
	ctx := actor.WithActor(context.Background(), a)
	r, err := productsubscription.NewCodyGatewayAccessResolver(ctx, logtest.Scoped(t), db, subscriptionID)
	require.NoError(t, err)
	return r
}
