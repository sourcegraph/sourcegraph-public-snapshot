package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestCodeMonitorStoreLastSearched(t *testing.T) {
	ctx := context.Background()
	db := NewEnterpriseDB(database.NewDB(dbtest.NewDB(t)))
	cm := db.CodeMonitors()

	// Create rows to satisfy foreign key constriant
	u, err := db.Users().Create(ctx, database.NewUser{Email: "test", Username: "test", EmailVerificationCode: "test"})
	require.NoError(t, err)
	ctx = actor.WithActor(ctx, actor.FromUser(u.ID))
	m, err := cm.CreateMonitor(ctx, MonitorArgs{NamespaceUserID: &u.ID})
	require.NoError(t, err)

	// Insert
	insertLastSearched := []string{"commit1", "commit2"}
	err = cm.UpsertLastSearched(ctx, m.ID, 3851, insertLastSearched)
	require.NoError(t, err)

	// Get
	lastSearched, err := cm.GetLastSearched(ctx, m.ID, 3851)
	require.NoError(t, err)
	require.Equal(t, insertLastSearched, lastSearched)

	// Update
	updateLastSearched := []string{"commit3", "commit4"}
	err = cm.UpsertLastSearched(ctx, m.ID, 3851, updateLastSearched)
	require.NoError(t, err)

	// Get
	lastSearched, err = cm.GetLastSearched(ctx, m.ID, 3851)
	require.NoError(t, err)
	require.Equal(t, updateLastSearched, lastSearched)
}
