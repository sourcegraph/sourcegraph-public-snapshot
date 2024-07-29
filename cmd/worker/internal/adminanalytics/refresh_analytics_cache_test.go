package adminanalytics

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

func TestRefreshAnalyticsCache(t *testing.T) {
	cache := rcache.SetupForTest(t)
	db := database.NewDB(logtest.NoOp(t), dbtest.NewDB(t))
	err := refreshAnalyticsCache(context.Background(), cache, db)
	require.NoError(t, err)
}
