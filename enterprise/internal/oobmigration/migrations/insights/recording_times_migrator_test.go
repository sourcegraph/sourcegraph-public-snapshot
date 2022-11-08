package insights

import (
	"testing"

	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestRecordingTimesMigrator(t *testing.T) {
	logger := logtest.Scoped(t)
	//postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))

	//permStore := store.NewInsightPermissionStore(postgres)
	insightsStore := basestore.NewWithHandle(insightsDB.Handle())

	migrator := NewRecordingTimesMigrator(insightsStore)
}
