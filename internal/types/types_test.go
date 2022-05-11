package types

import (
	"reflect"
	"testing"
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
