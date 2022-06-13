package permissions

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := database.NewDB(dbtest.NewDB(t))

	ctx := context.Background()
	jobID, err := db.BitbucketProjectPermissions().Enqueue(ctx, "project1", 2, []types.UserPermission{
		{BindID: "user1", Permission: "read"},
		{BindID: "user2", Permission: "admin"},
	}, false)
	require.NoError(t, err)
	require.NotZero(t, jobID)

	store := createBitbucketProjectPermissionsStore(db)
	count, err := store.QueuedCount(ctx, true, nil)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func intPtr(v int) *int              { return &v }
func stringPtr(v string) *string     { return &v }
func timePtr(v time.Time) *time.Time { return &v }

func mustParseTime(v string) time.Time {
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		panic(err)
	}
	return t
}

func TestGetBitbucketClient(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx := context.Background()

	var c schema.BitbucketServerConnection
	c.Token = "secret"
	c.Url = "http://some-url"
	c.Username = "username"

	cfg, err := json.Marshal(&c)
	require.NoError(t, err)

	svc := types.ExternalService{
		Config: string(cfg),
	}

	var handler bitbucketProjectPermissionsHandler
	client, err := handler.getBitbucketClient(ctx, &svc)
	require.NoError(t, err)
	require.NotNil(t, client)
}
