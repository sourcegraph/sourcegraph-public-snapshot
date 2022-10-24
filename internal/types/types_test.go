package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExternalServiceFields will fail if a new field is added to
// ExternalService to ensure that we update ToAPIService
func TestExternalServiceFields(t *testing.T) {
	v := reflect.ValueOf(ExternalService{})
	// If this test fails it means that fields have changed on types.ExternalService.
	// Ensure that types.ExternalService and api.ExternalService are consistent and
	// also that types.ExternalService.ToAPIService has been updated.
	wantFieldCount := 15
	if wantFieldCount != v.NumField() {
		t.Fatalf("Expected %d fields, got %d. See comments in failing test", wantFieldCount, v.NumField())
	}
}

func TestCodeHostURN(t *testing.T) {
	t.Run("normalize URL", func(t *testing.T) {
		const url = "https://github.com"
		urn, err := ParseCodeHostURN(url)
		require.NoError(t, err)

		assert.Equal(t, "https://github.com/", urn.String())
	})

	t.Run(`empty CodeHostURN.String() returns ""`, func(t *testing.T) {
		urn := CodeHostURN{}
		assert.Equal(t, "", urn.String())
	})
}
