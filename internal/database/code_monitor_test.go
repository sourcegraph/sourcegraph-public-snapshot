package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type testFixtures struct {
	monitor    *Monitor
	query      *QueryTrigger
	emails     [2]*EmailAction
	recipients [2]*Recipient
}

func (s *codeMonitorStore) insertTestMonitor(ctx context.Context, t *testing.T) *testFixtures {
	t.Helper()

	fixtures := testFixtures{}

	actions := []*EmailActionArgs{
		{
			Enabled:        true,
			IncludeResults: false,
			Priority:       "NORMAL",
			Header:         "test header 1",
		},
		{
			Enabled:        true,
			IncludeResults: false,
			Priority:       "CRITICAL",
			Header:         "test header 2",
		},
	}
	// Create monitor.
	uid := actor.FromContext(ctx).UID
	var err error
	fixtures.monitor, err = s.CreateMonitor(ctx, MonitorArgs{
		Description:     testDescription,
		Enabled:         true,
		NamespaceUserID: &uid,
	})
	require.NoError(t, err)

	// Create trigger.
	fixtures.query, err = s.CreateQueryTrigger(ctx, fixtures.monitor.ID, testQuery)
	require.NoError(t, err)

	for i, a := range actions {
		fixtures.emails[i], err = s.CreateEmailAction(ctx, fixtures.monitor.ID, &EmailActionArgs{
			Enabled:        a.Enabled,
			IncludeResults: a.IncludeResults,
			Priority:       a.Priority,
			Header:         a.Header,
		})
		require.NoError(t, err)

		fixtures.recipients[i], err = s.CreateRecipient(ctx, fixtures.emails[i].ID, &uid, nil)
		require.NoError(t, err)
		// TODO(camdencheek): add other action types (webhooks) here
	}
	return &fixtures
}

type codeMonitorTestFixtures struct {
	User    *types.User
	Monitor *Monitor
	Query   *QueryTrigger
	Repo    *types.Repo
}

func populateCodeMonitorFixtures(t *testing.T, db DB) codeMonitorTestFixtures {
	ctx := context.Background()
	u, err := db.Users().Create(ctx, NewUser{Email: "test", Username: "test", EmailVerificationCode: "test"})
	require.NoError(t, err)
	err = db.Repos().Create(ctx, &types.Repo{Name: "test"})
	require.NoError(t, err)
	r, err := db.Repos().GetByName(ctx, "test")
	require.NoError(t, err)
	ctx = actor.WithActor(ctx, actor.FromUser(u.ID))
	m, err := db.CodeMonitors().CreateMonitor(ctx, MonitorArgs{NamespaceUserID: &u.ID, Enabled: true})
	require.NoError(t, err)
	q, err := db.CodeMonitors().CreateQueryTrigger(ctx, m.ID, "type:commit repo:.")
	require.NoError(t, err)
	return codeMonitorTestFixtures{User: u, Monitor: m, Query: q, Repo: r}
}
