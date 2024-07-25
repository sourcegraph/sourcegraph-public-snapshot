package adminanalytics

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/adminanalytics"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

func TestRefreshAnalyticsCache(t *testing.T) {
	adminanalytics.MockStore = rcache.SetupForTest(t)
	defer func() {
		adminanalytics.MockStore = nil
	}()
	db := database.NewDB(logtest.NoOp(t), dbtest.NewDB(t))
	err := refreshAnalyticsCache(context.Background(), db)
	require.NoError(t, err)
}
