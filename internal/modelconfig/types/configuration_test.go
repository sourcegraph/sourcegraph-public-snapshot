package types

import (
	"encoding/json"
	"testing"

	"github.com/go-enry/go-enry/v2/regex"
)

func TestClientSideProviderConfigMarshalJSON(t *testing.T) {
	config := ClientSideProviderConfig{}

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal ServerSideProviderConfig: %v", err)
	}

	actual := string(jsonBytes)

	// Check whether ClientSideProviderConfig contains any data which looks sensitive
	secretRe := regex.MustCompile("(?i)access|token|secret|key")
	if secretRe.MatchString(actual) {
		t.Fatalf("ClientSideProviderConfig contains a field that appears to be sensitive - this should be stored under ServerSideProviderConfig to prevent secret disclosure")
	}

	expected := `{}`
	if actual != expected {
		t.Errorf("Expected JSON: %s, but got: %s", expected, actual)
	}
}
