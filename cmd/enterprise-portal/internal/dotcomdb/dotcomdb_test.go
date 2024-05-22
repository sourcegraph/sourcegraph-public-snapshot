package dotcomdb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	pgxstdlibv4 "github.com/jackc/pgx/v4/stdlib"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/dotcomproductsubscriptiontest"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func newTestDotcomReader(t *testing.T) (database.DB, *dotcomdb.Reader) {
	ctx := context.Background()

	// Set up a Sourcegraph test database.
	sgtestdb := dbtest.NewDB(t)

	// HACK: Extract the underlying pgx connection from the Sourcegraph test
	// database, so that we can tease out the connection string that was used
	// to connect to it, since we want to use pgx/v5 directly as offered by MSP
	// utilities. This is sneaky but easier than changing the extensive
	// infrastructure around dbtest.
	var connString string
	c, err := sgtestdb.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, c.Raw(func(driverConn any) error {
		// Unwrap all the way down. If something wraps the pgx conn in a
		// non-unwrappable fashion, that would be unfortunate, we'll end up
		// panicking in the next section.
		for {
			if wc, ok := driverConn.(dbconn.UnwrappableConn); ok {
				driverConn = wc.Raw()
			} else {
				break
			}
		}
		// Sourcegraph database uses pgxv4 under the hood, so we cast into the
		// v4 version of pgx internals
		connString = driverConn.(*pgxstdlibv4.Conn).Conn().Config().ConnString()
		return nil
	}))

	// Now create a new connection using the conn string 😎
	t.Logf("pgx.Connect %q", connString)
	conn, err := pgx.Connect(ctx, connString)
	require.NoError(t, err)

	// Make sure it works!
	r := dotcomdb.NewReader(conn)
	require.NoError(t, r.Ping(ctx))

	return database.NewDB(logtest.Scoped(t), sgtestdb), r
}

type mockAccess struct {
	subscriptionID string
	accessTokens   []string
}

func setupDBAndInsertMockLicense(t *testing.T, dotcomdb database.DB, info license.Info, cgAccess graphqlbackend.UpdateCodyGatewayAccessInput) mockAccess {
	ctx := context.Background()
	subdb := dotcomproductsubscriptiontest.NewSubscriptionsDB(t, dotcomdb)
	ldb := dotcomproductsubscriptiontest.NewLicensesDB(t, dotcomdb)
	result := mockAccess{}

	{
		// Create a different subscription and license that's rubbish,
		// created at the same time, to ensure we don't use it
		u, err := dotcomdb.Users().Create(ctx, database.NewUser{Username: "barbaz"})
		require.NoError(t, err)
		sub, err := subdb.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)
		_, err = ldb.Create(ctx, sub, t.Name()+"-barbaz", 2, license.Info{
			CreatedAt: info.CreatedAt,
			ExpiresAt: info.ExpiresAt,
		})
		require.NoError(t, err)
	}

	// Create the subscription we will assert against
	u, err := dotcomdb.Users().Create(ctx, database.NewUser{Username: "user"})
	require.NoError(t, err)
	subid, err := subdb.Create(ctx, u.ID, u.Username)
	require.NoError(t, err)
	result.subscriptionID = subid
	// Insert a rubbish license first, CreatedAt is not used (creation time is
	// inferred from insert time) so we need to do this first
	key1 := t.Name() + "-1"
	result.accessTokens = append(result.accessTokens, license.GenerateLicenseKeyBasedAccessToken(key1))
	_, err = ldb.Create(ctx, subid, key1, 2, license.Info{
		CreatedAt: info.CreatedAt.Add(-time.Hour),
		ExpiresAt: info.ExpiresAt.Add(-time.Hour),
	})
	require.NoError(t, err)
	// Now create the actual requested license, and test that this is the one
	// that gets used.
	key2 := t.Name() + "-2"
	result.accessTokens = append(result.accessTokens, license.GenerateLicenseKeyBasedAccessToken(key2))
	_, err = ldb.Create(ctx, subid, key2, 2, info)
	require.NoError(t, err)
	// Configure Cody Gateway access for this subscription as we do today
	require.NoError(t, subdb.UpdateCodyGatewayAccess(ctx, subid, cgAccess))

	{
		// Create another different subscription and license that's also rubbish,
		// created at the same time, to ensure we don't use it
		u, err := dotcomdb.Users().Create(ctx, database.NewUser{Username: "foobar"})
		require.NoError(t, err)
		sub, err := subdb.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)
		_, err = ldb.Create(ctx, sub, t.Name()+"-foobar", 2, license.Info{
			CreatedAt: info.CreatedAt,
			ExpiresAt: info.ExpiresAt,
		})
		require.NoError(t, err)
	}

	return result
}

func mustWithCtx[T any](t *testing.T, fn func(context.Context) (T, error)) T {
	v, err := fn(context.Background())
	require.NoError(t, err)
	return v
}

func TestGetCodyGatewayAccessAttributes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	for _, tc := range []struct {
		name     string
		info     license.Info
		cgAccess graphqlbackend.UpdateCodyGatewayAccessInput
	}{{
		name: "disabled",
		info: license.Info{
			CreatedAt: time.Now().Add(-30 * time.Minute),
			ExpiresAt: time.Now().Add(30 * time.Minute),
			UserCount: 10,
		},
		cgAccess: graphqlbackend.UpdateCodyGatewayAccessInput{
			Enabled: pointers.Ptr(false),
		},
	}, {
		name: "user and tags",
		info: license.Info{
			CreatedAt: time.Now().Add(-30 * time.Minute),
			ExpiresAt: time.Now().Add(30 * time.Minute),
			UserCount: 321,
			Tags:      []string{licensing.PlanEnterprise1.Tag()},
		},
		cgAccess: graphqlbackend.UpdateCodyGatewayAccessInput{Enabled: pointers.Ptr(true)},
	}, {
		name: "overrides",
		info: license.Info{
			CreatedAt: time.Now().Add(-30 * time.Minute),
			ExpiresAt: time.Now().Add(30 * time.Minute),
			UserCount: 10,
		},
		cgAccess: graphqlbackend.UpdateCodyGatewayAccessInput{
			Enabled:                                 pointers.Ptr(true),
			ChatCompletionsRateLimit:                pointers.Ptr[graphqlbackend.BigInt](123),
			CodeCompletionsRateLimitIntervalSeconds: pointers.Ptr[int32](456),
			EmbeddingsRateLimit:                     pointers.Ptr[graphqlbackend.BigInt](789),
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel() // parallel per newTestDotcomReader
			dotcomdb, dotcomreader := newTestDotcomReader(t)

			// First, set up a subscription and license and some other rubbish
			// data to ensure we only get the license we want.
			mock := setupDBAndInsertMockLicense(t, dotcomdb, tc.info, tc.cgAccess)

			t.Run("by subscription ID", func(t *testing.T) {
				attr, err := dotcomreader.GetCodyGatewayAccessAttributesBySubscription(ctx, mock.subscriptionID)
				require.NoError(t, err)
				validateAccessAttributes(t, dotcomdb, mock, attr, tc.info)
			})

			t.Run("by access token", func(t *testing.T) {
				for i, token := range mock.accessTokens {
					t.Run(fmt.Sprintf("token %d", i), func(t *testing.T) {
						attr, err := dotcomreader.GetCodyGatewayAccessAttributesByAccessToken(ctx, token)
						require.NoError(t, err)
						validateAccessAttributes(t, dotcomdb, mock, attr, tc.info)
					})
				}
			})
		})
	}
}

func validateAccessAttributes(t *testing.T, dotcomdb database.DB, mock mockAccess, attr *dotcomdb.CodyGatewayAccessAttributes, info license.Info) {
	assert.Equal(t, mock.subscriptionID, attr.SubscriptionID)
	assert.Equal(t, int(info.UserCount), *attr.ActiveLicenseUserCount)
	assert.Len(t, attr.LicenseKeyHashes, 2)
	assert.Equal(t, attr.GenerateAccessTokens(), mock.accessTokens)
	limits := attr.EvaluateRateLimits()

	// Validate against the expected values as produced by existing resolvers
	expected := dotcomproductsubscriptiontest.NewCodyGatewayAccessResolver(t, dotcomdb, mock.subscriptionID)
	assert.Equal(t, expected.Enabled(), attr.CodyGatewayEnabled)
	if !expected.Enabled() {
		// We don't care about the rest of the attributes if access is disabled,
		// the resolver returns nil for everything
		return
	}
	for _, compare := range []struct {
		name     string
		expected graphqlbackend.CodyGatewayRateLimit
		got      licensing.CodyGatewayRateLimit
	}{{
		name:     "Chat",
		expected: mustWithCtx(t, expected.ChatCompletionsRateLimit),
		got:      limits.Chat,
	}, {
		name:     "Code",
		expected: mustWithCtx(t, expected.CodeCompletionsRateLimit),
		got:      limits.Code,
	}, {
		name:     "Embeddings",
		expected: mustWithCtx(t, expected.EmbeddingsRateLimit),
		got:      limits.Embeddings,
	}} {
		t.Run(compare.name, func(t *testing.T) {
			// We only care about limit and interval now
			assert.Equal(t, int64(compare.expected.Limit()), compare.got.Limit, "Limit")
			assert.Equal(t, compare.expected.IntervalSeconds(), compare.got.IntervalSeconds, "IntervalSeconds")
		})
	}
}

func TestGetAllCodyGatewayAccessAttributes(t *testing.T) {
	t.Parallel() // parallel per newTestDotcomReader
	dotcomdb, dotcomreader := newTestDotcomReader(t)

	info := license.Info{
		CreatedAt: time.Now().Add(-30 * time.Minute),
		ExpiresAt: time.Now().Add(30 * time.Minute),
		UserCount: 321,
		Tags:      []string{licensing.PlanEnterprise1.Tag()},
	}
	cgAccess := graphqlbackend.UpdateCodyGatewayAccessInput{Enabled: pointers.Ptr(true)}
	mock := setupDBAndInsertMockLicense(t, dotcomdb, info, cgAccess)

	attrs, err := dotcomreader.GetAllCodyGatewayAccessAttributes(context.Background())
	require.NoError(t, err)
	assert.Len(t, attrs, 3) // 3 subscriptions created in setupDBAndInsertMockLicense
	var found bool
	for _, attr := range attrs {
		if attr.SubscriptionID == mock.subscriptionID {
			found = true
			validateAccessAttributes(t, dotcomdb, mock, attr, info)
		} else {
			assert.False(t, attr.CodyGatewayEnabled)
		}
	}
	assert.True(t, found)
}
