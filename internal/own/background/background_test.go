package background

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	dbworker "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func Test_Handle(t *testing.T) {
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	err := db.Repos().Create(ctx, &types.Repo{Name: "fakerepo", ID: 1})
	require.NoError(t, err)

	config, err := loadConfig(ctx, IndexJobType{Name: "recent-contributors"}, db.OwnSignalConfigurations())
	require.NoError(t, err)

	insertJob(t, db, 1, config, "queued", time.Time{})

	err = db.OwnSignalConfigurations().UpdateConfiguration(ctx, database.UpdateSignalConfigurationArgs{Name: config.Name, Enabled: true})
	require.NoError(t, err)

	t.Run("verify signal that is enabled shows up in queue", func(t *testing.T) {
		store := makeWorkerStore(db, obsCtx)
		count, err := store.CountByState(ctx, dbworker.StateQueued|dbworker.StateErrored)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

	})
	t.Run("verify signal that is disabled doesn't show up in queue", func(t *testing.T) {
		store := makeWorkerStore(db, obsCtx)
		count, err := store.CountByState(ctx, dbworker.StateQueued|dbworker.StateErrored)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		err = db.OwnSignalConfigurations().UpdateConfiguration(ctx, database.UpdateSignalConfigurationArgs{Name: config.Name, Enabled: false})
		require.NoError(t, err)

		count, err = store.CountByState(ctx, dbworker.StateQueued|dbworker.StateErrored)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func Test_JanitorTable(t *testing.T) {
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	store := basestore.NewWithHandle(db.Handle())

	clearTable := func(t *testing.T) {
		t.Helper()
		err := store.Exec(ctx, sqlf.Sprintf("truncate %s", sqlf.Sprintf(tableName)))
		if err != nil {
			t.Fatal(err)
		}
	}

	countRows := func(t *testing.T) int {
		t.Helper()
		val, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf("select count(*) from %s", sqlf.Sprintf(tableName))))
		if err != nil {
			t.Fatal(err)
		}
		return val
	}

	config, err := loadConfig(ctx, IndexJobType{Name: "recent-contributors"}, db.OwnSignalConfigurations())
	require.NoError(t, err)

	tests := []struct {
		name               string
		isSignalEnabled    bool
		jobStates          []string
		jobFinsishedAt     time.Time
		shouldGetJanitored bool
	}{
		{
			name:               "signal enabled jobs in progress or queued should not expire",
			isSignalEnabled:    true,
			jobStates:          []string{"queued", "processing", "errored"},
			jobFinsishedAt:     time.Time{},
			shouldGetJanitored: false,
		},
		{
			name:               "signal enabled completed jobs not old enough to expire",
			isSignalEnabled:    true,
			jobStates:          []string{"failed", "completed"},
			jobFinsishedAt:     time.Now(),
			shouldGetJanitored: false,
		},
		{
			name:               "signal enabled completed jobs sufficient age to expire",
			isSignalEnabled:    true,
			jobStates:          []string{"failed", "completed"},
			jobFinsishedAt:     time.Now().AddDate(0, 0, -2),
			shouldGetJanitored: true,
		},
		{
			name:               "signal disabled all jobs should expire",
			isSignalEnabled:    false,
			jobStates:          []string{"queued", "processing", "errored", "completed", "failed"},
			jobFinsishedAt:     time.Time{},
			shouldGetJanitored: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, state := range test.jobStates {
				t.Run(state, func(t *testing.T) {
					clearTable(t)
					err := db.OwnSignalConfigurations().UpdateConfiguration(ctx, database.UpdateSignalConfigurationArgs{Name: config.Name, Enabled: test.isSignalEnabled})
					require.NoError(t, err)

					insertJob(t, db, 1, config, state, test.jobFinsishedAt)

					_, _, err = janitorFunc(db, time.Hour*24)(ctx)
					require.NoError(t, err)

					expected := 1
					if test.shouldGetJanitored {
						expected = 0
					}
					assert.Equal(t, expected, countRows(t), "rows in table equal to expected")
				})
			}
		})
	}
}
