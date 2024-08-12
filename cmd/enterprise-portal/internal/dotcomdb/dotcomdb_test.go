package dotcomdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	pgxstdlibv4 "github.com/jackc/pgx/v4/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/dotcomproductsubscriptiontest"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	sgdatabase "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
)

func newTestDotcomReader(t *testing.T, opts dotcomdb.ReaderOptions) (sgdatabase.DB, *dotcomdb.Reader) {
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
	conn, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	// Make sure it works!
	r := dotcomdb.NewReader(conn, opts)
	require.NoError(t, r.Ping(ctx))

	return sgdatabase.NewDB(logtest.Scoped(t), sgtestdb), r
}

type mockedData struct {
	targetSubscriptionID  string
	accessTokens          []string
	createdLicenses       int
	createdSubscriptions  int
	archivedSubscriptions int
}

func setupDBAndInsertMockLicense(t *testing.T, dotcomdb sgdatabase.DB, info license.Info, cgAccess *graphqlbackend.UpdateCodyGatewayAccessInput) mockedData {
	start := time.Now()

	ctx := context.Background()
	subscriptionsdb := dotcomproductsubscriptiontest.NewSubscriptionsDB(t, dotcomdb)
	licensesdb := dotcomproductsubscriptiontest.NewLicensesDB(t, dotcomdb)
	result := mockedData{}

	{
		// Create a different subscription and license that's rubbish,
		// created at the same time, to ensure we don't use it
		u, err := dotcomdb.Users().Create(ctx, sgdatabase.NewUser{Username: "barbaz"})
		require.NoError(t, err)
		sub, err := subscriptionsdb.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)
		result.createdSubscriptions += 1
		_, err = licensesdb.Create(ctx, sub, t.Name()+"-barbaz", 2, license.Info{
			CreatedAt: info.CreatedAt,
			ExpiresAt: info.ExpiresAt,
			Tags:      []string{licensing.DevTag},
		})
		require.NoError(t, err)
		result.createdLicenses += 1
	}

	{
		// Create a different subscription and license that's archived,
		// created at the same time, to ensure we don't use it
		u, err := dotcomdb.Users().Create(ctx, sgdatabase.NewUser{Username: "archived"})
		require.NoError(t, err)
		sub, err := subscriptionsdb.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)
		result.createdSubscriptions += 1
		_, err = licensesdb.Create(ctx, sub, t.Name()+"-archived", 2, license.Info{
			CreatedAt: info.CreatedAt,
			ExpiresAt: info.ExpiresAt,
			Tags:      []string{licensing.DevTag},
		})
		require.NoError(t, err)
		result.createdLicenses += 1
		// Archive the subscription
		require.NoError(t, subscriptionsdb.Archive(ctx, sub))
		result.archivedSubscriptions += 1
	}

	{
		// Create a different subscription and license that's not a dev tag,
		// created at the same time, to ensure we don't use it
		u, err := dotcomdb.Users().Create(ctx, sgdatabase.NewUser{Username: "not-devlicense"})
		require.NoError(t, err)
		sub, err := subscriptionsdb.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)
		result.createdSubscriptions += 1
		_, err = licensesdb.Create(ctx, sub, t.Name()+"-not-devlicense", 2, license.Info{
			CreatedAt: info.CreatedAt,
			ExpiresAt: info.ExpiresAt,
			Tags:      []string{},
		})
		require.NoError(t, err)
	}

	// Create the subscription we will assert against
	u, err := dotcomdb.Users().Create(ctx, sgdatabase.NewUser{Username: "user"})
	require.NoError(t, err)
	subid, err := subscriptionsdb.Create(ctx, u.ID, u.Username)
	require.NoError(t, err)
	result.createdSubscriptions += 1
	result.targetSubscriptionID = subid
	// Insert a rubbish license first, CreatedAt is not used (creation time is
	// inferred from insert time) so we need to do this first
	key1 := t.Name() + "-1"
	result.accessTokens = append(result.accessTokens, license.GenerateLicenseKeyBasedAccessToken(key1))
	_, err = licensesdb.Create(ctx, subid, key1, 2, license.Info{
		CreatedAt: info.CreatedAt.Add(-time.Hour),
		ExpiresAt: info.ExpiresAt.Add(-time.Minute), // should expire first, but not be expired
		Tags:      []string{licensing.DevTag},
	})
	require.NoError(t, err)
	result.createdLicenses += 1
	// Now create the actual requested license, and test that this is the one
	// that gets used.
	key2 := t.Name() + "-2"
	result.accessTokens = append(result.accessTokens, license.GenerateLicenseKeyBasedAccessToken(key2))
	_, err = licensesdb.Create(ctx, subid, key2, 2, info)
	require.NoError(t, err)
	result.createdLicenses += 1
	// Configure Cody Gateway access for this subscription as we do today
	if cgAccess != nil {
		require.NoError(t, subscriptionsdb.UpdateCodyGatewayAccess(ctx, subid, *cgAccess))
	}

	{
		// Create another different subscription and license that's also rubbish,
		// created at the same time, to ensure we don't use it
		u, err := dotcomdb.Users().Create(ctx, sgdatabase.NewUser{Username: "foobar"})
		require.NoError(t, err)
		sub, err := subscriptionsdb.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)
		result.createdSubscriptions += 1
		_, err = licensesdb.Create(ctx, sub, t.Name()+"-foobar", 2, license.Info{
			CreatedAt: info.CreatedAt,
			ExpiresAt: info.ExpiresAt,
			Tags:      []string{licensing.DevTag},
		})
		require.NoError(t, err)
		result.createdLicenses += 1
	}

	t.Logf("Setup complete in %s", time.Since(start).String())
	return result
}

func TestListEnterpriseSubscriptionLicenses(t *testing.T) {
	t.Parallel()

	db, dotcomreader := newTestDotcomReader(t, dotcomdb.ReaderOptions{
		DevOnly: true,
	})
	info := license.Info{
		ExpiresAt: time.Now().Add(30 * time.Minute),
		UserCount: 321,
		Tags:      []string{licensing.PlanEnterprise1.Tag(), licensing.DevTag},
	}
	rootTestName := t.Name()
	mock := setupDBAndInsertMockLicense(t, db, info, nil)

	assertMockLicense := func(t *testing.T, l *dotcomdb.LicenseAttributes) {
		assert.Equal(t, mock.targetSubscriptionID, l.SubscriptionID)
		assert.Equal(t, info.UserCount, uint(*l.UserCount))
		assert.NotEmpty(t, l.CreatedAt)
		assert.Equal(t,
			info.ExpiresAt.Format(time.DateTime),
			l.ExpiresAt.Format(time.DateTime))
		assert.Equal(t, info.Tags, l.Tags)
	}

	ctx := context.Background()
	for _, tc := range []struct {
		name     string
		filters  []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter
		pageSize int
		expect   func(t *testing.T, licenses []*dotcomdb.LicenseAttributes)
	}{{
		name:    "no filters",
		filters: nil,
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			// Only unarchived subscriptions
			assert.Len(t, licenses, mock.createdLicenses-mock.archivedSubscriptions)
		},
	}, {
		name: "filter by subscription ID",
		filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId{
				SubscriptionId: mock.targetSubscriptionID,
			},
		}},
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			assert.Len(t, licenses, 2) // setupDBAndInsertMockLicense adds 2
			// The first one is our most recently created one, i.e. the one we
			// requested in setupDBAndInsertMockLicense
			assertMockLicense(t, licenses[0])
		},
	}, {
		name: "filter by subscription ID and limit 1",
		filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId{
				SubscriptionId: mock.targetSubscriptionID,
			},
		}},
		pageSize: 1,
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			assert.Len(t, licenses, 1)
			assertMockLicense(t, licenses[0])
		},
	}, {
		name: "filter by subscription ID and not archived",
		filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId{
				SubscriptionId: mock.targetSubscriptionID,
			},
		}, {
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_IsRevoked{
				IsRevoked: false,
			},
		}},
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			// setupDBAndInsertMockLicense adds 2 for the target subscription,
			// both unarchived
			assert.Len(t, licenses, 2)
			for _, l := range licenses {
				assert.Equal(t, mock.targetSubscriptionID, l.SubscriptionID)
			}
		},
	}, {
		name: "filter by is archived",
		filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_IsRevoked{
				IsRevoked: true,
			},
		}},
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			assert.Len(t, licenses, mock.archivedSubscriptions)
		},
	}, {
		name: "filter by not archived",
		filters: []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter{{
			Filter: &subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_IsRevoked{
				IsRevoked: false,
			},
		}},
		expect: func(t *testing.T, licenses []*dotcomdb.LicenseAttributes) {
			assert.Len(t, licenses, mock.createdLicenses-mock.archivedSubscriptions)
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			licenses, err := dotcomreader.ListEnterpriseSubscriptionLicenses(ctx, tc.filters, tc.pageSize)
			require.NoError(t, err)
			for _, l := range licenses {
				// Each mock license key contains a variation of the root test name
				assert.Contains(t, l.LicenseKey, rootTestName)
			}
			if tc.expect != nil {
				tc.expect(t, licenses)
			}
		})
	}
}

func TestListEnterpriseSubscriptions(t *testing.T) {
	t.Run("devonly", func(t *testing.T) {
		t.Parallel()

		db, dotcomreader := newTestDotcomReader(t, dotcomdb.ReaderOptions{
			DevOnly: true,
		})
		info := license.Info{
			ExpiresAt: time.Now().Add(30 * time.Minute),
			UserCount: 321,
			Tags:      []string{licensing.PlanEnterprise1.Tag(), licensing.DevTag},
		}
		mock := setupDBAndInsertMockLicense(t, db, info, nil)

		ss, err := dotcomreader.ListEnterpriseSubscriptions(
			context.Background(),
			dotcomdb.ListEnterpriseSubscriptionsOptions{})
		require.NoError(t, err)
		// We expect 1 less subscription because one of the subscriptions does not
		// have a dev/internal license
		assert.Len(t, ss, mock.createdSubscriptions-mock.archivedSubscriptions-1)
		for _, s := range ss {
			s.CreatedAt = time.Time{} // zero time for autogold
		}
		autogold.Expect("foobar - 0001-01-01 00:00:00").Equal(t, ss[0].GenerateDisplayName())
		autogold.Expect("user - 0001-01-01 00:00:00").Equal(t, ss[1].GenerateDisplayName())
		autogold.Expect("barbaz - 0001-01-01 00:00:00").Equal(t, ss[2].GenerateDisplayName())

		var found bool
		for _, s := range ss {
			if s.ID == mock.targetSubscriptionID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("not devonly", func(t *testing.T) {
		t.Parallel()

		db, dotcomreader := newTestDotcomReader(t, dotcomdb.ReaderOptions{
			DevOnly: false,
		})
		info := license.Info{
			ExpiresAt: time.Now().Add(30 * time.Minute),
			UserCount: 321,
			Tags:      []string{licensing.PlanEnterprise1.Tag(), licensing.DevTag},
		}
		mock := setupDBAndInsertMockLicense(t, db, info, nil)

		ss, err := dotcomreader.ListEnterpriseSubscriptions(
			context.Background(),
			dotcomdb.ListEnterpriseSubscriptionsOptions{})
		require.NoError(t, err)
		// all subscriptions included, minus the ones without any license
		assert.Len(t, ss, mock.createdSubscriptions-1)
	})
}
