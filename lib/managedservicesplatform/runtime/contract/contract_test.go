package contract_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	mock "github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract/mock"
)

func TestNewContract(t *testing.T) {
	t.Run("sanity check", func(t *testing.T) {
		c, err := mock.NewContract(t, mock.ServiceMetadata{}, "MSP=true")
		if errors.Is(err, mock.MockError{}) {
			t.Fatal(err)
		}
		assert.NotZero(t, c)
		assert.True(t, c.MSP)

		// If our error is not a Mockerror, then it could be an environment validation error
		// Expected to error, as there are missing required env vars.
		envErr := err
		assert.Error(t, envErr)
	})

	// TODO: Add more validation tests
}
