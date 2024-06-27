package contract_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
)

type mockServiceMetadata struct{}

func (m mockServiceMetadata) Name() string    { return "mock-name" }
func (m mockServiceMetadata) Version() string { return "mock-version" }

func TestNewContract(t *testing.T) {
	t.Run("service sanity check", func(t *testing.T) {
		e, err := contract.ParseEnv([]string{"MSP=true"})
		require.NoError(t, err)

		c := contract.NewService(logtest.Scoped(t), mockServiceMetadata{}, e)
		assert.NotZero(t, c)
		assert.True(t, c.MSP)

		// Expected to error, as there are missing required env vars.
		err = e.Validate()
		assert.Error(t, err)
	})

	t.Run("job sanity check", func(t *testing.T) {
		e, err := contract.ParseEnv([]string{"MSP=true"})
		require.NoError(t, err)
		c := contract.NewJob(logtest.Scoped(t), mockServiceMetadata{}, e)
		assert.NotZero(t, c)
		assert.True(t, c.MSP)

		// Not expected to error, as there are no required env vars.
		err = e.Validate()
		assert.NoError(t, err)
	})
	// TODO: Add more validation tests
}
