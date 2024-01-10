package productsubscription

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/audit/audittest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/license"
)

func TestProductSubscription_Account(t *testing.T) {
	logger := logtest.Scoped(t)
	t.Run("user not found should be ignored", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefaultReturn(nil, &errcode.Mock{IsNotFound: true})

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		_, err := (&productSubscription{logger: logger, v: &dbSubscription{UserID: 1}, db: db}).Account(context.Background())
		assert.Nil(t, err)
	})
}

// Test cases are very simple for now to expedite assertions that we are
// generating adequate access logs. In the future we can extend this to
// better cover more scenarios.
func TestProductSubscriptionActiveLicense(t *testing.T) {
	ctx := context.Background()
	db := database.NewDB(logtest.Scoped(t), dbtest.NewDB(t))
	subscriptionsDB := dbSubscriptions{db: db}
	licensesDB := dbLicenses{db: db}

	// Set global feature flag so we can override it per-user
	_, err := db.FeatureFlags().CreateBool(ctx, featureFlagProductSubscriptionsServiceAccount, false)
	require.NoError(t, err)

	// Site admin
	adminUser, err := db.Users().Create(ctx, database.NewUser{Username: "admin"})
	require.NoError(t, err)
	err = db.Users().SetIsSiteAdmin(ctx, adminUser.ID, true)
	require.NoError(t, err)

	// User owning the subscription in question
	ownerUser, err := db.Users().Create(ctx, database.NewUser{Username: "verified"})
	require.NoError(t, err)
	sub, err := subscriptionsDB.Create(ctx, ownerUser.ID, "subscription")
	require.NoError(t, err)
	_, err = licensesDB.Create(ctx, sub, "license-key", 1, license.Info{})
	require.NoError(t, err)

	// Service account user
	serviceAccountUser, err := db.Users().Create(ctx, database.NewUser{Username: "serviceaccount"})
	require.NoError(t, err)
	_, err = db.FeatureFlags().CreateOverride(ctx, &featureflag.Override{
		UserID:   &serviceAccountUser.ID,
		FlagName: featureFlagProductSubscriptionsServiceAccount,
	})
	require.NoError(t, err)

	// Test cases
	for _, test := range []struct {
		name           string
		actor          *actor.Actor
		subscriptionID graphql.ID
	}{
		{
			name:           "site admin",
			actor:          actor.FromActualUser(adminUser),
			subscriptionID: marshalProductSubscriptionID(sub),
		},
		{
			name:           "subscription owner",
			actor:          actor.FromActualUser(adminUser),
			subscriptionID: marshalProductSubscriptionID(sub),
		},
		{
			name:           "service account",
			actor:          actor.FromActualUser(adminUser),
			subscriptionID: marshalProductSubscriptionID(sub),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			logger, exportLogs := logtest.Captured(t)

			requestCtx := actor.WithActor(ctx, test.actor)

			r := ProductSubscriptionLicensingResolver{Logger: logger, DB: db}

			// Resolve the subscription and then the active license of the subscription
			sub, err := r.ProductSubscriptionByID(requestCtx, test.subscriptionID)
			require.NoError(t, err)
			_, err = sub.ActiveLicense(requestCtx)
			require.NoError(t, err)

			// A subscription was resolved in this test case, we should have an
			// audit log
			assert.True(t, exportLogs().Contains(func(l logtest.CapturedLog) bool {
				fields, ok := audittest.ExtractAuditFields(l)
				if !ok {
					return ok
				}
				return fields.Entity == auditEntityProductSubscriptions &&
					fields.Action == "access"
			}))
		})
	}
}
