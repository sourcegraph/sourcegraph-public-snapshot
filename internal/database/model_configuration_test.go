package database

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestModelConfigurationStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	// Create the first user, with ID 1.
	_, err := db.Users().Create(ctx, NewUser{Username: "sa"})
	require.NoError(t, err)

	modelCfgStore := db.ModelConfiguration()

	// Verify that GetLatest works, even without any existing rows in the test database.
	t.Run("GetLatest", func(t *testing.T) {
		latest, err := modelCfgStore.GetLatest(ctx)
		require.NoError(t, err)

		// Confirm the initial/default configuration is what we expect.
		assert.WithinDuration(t, time.Now(), latest.CreatedAt, time.Second)
		assert.Nil(t, latest.BaseConfigurationJSON)

		require.NotNil(t, latest.CreatedBy)
		assert.EqualValues(t, 1, *latest.CreatedBy)

		assert.Equal(t, "{}", latest.RedactedConfigurationPatchJSON)
		assert.Equal(t, "{}", latest.EncryptedConfigurationPatchJSON)
		assert.Equal(t, "", latest.EncryptionKeyID)
		assert.EqualValues(t, 0, latest.Flags)

		// Confirm that calling it a second time will return the same data,
		// and not create a new row.
		secondCall, err := modelCfgStore.GetLatest(ctx)
		require.NoError(t, err)
		assert.EqualValues(t, *latest, *secondCall)
	})
}
