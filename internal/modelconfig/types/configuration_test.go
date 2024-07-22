package types

import (
	"encoding/json"
	"testing"

	"github.com/go-enry/go-enry/v2/regex"
)

func TestServerSideProviderConfigMarshalJSON(t *testing.T) {
	config := ServerSideProviderConfig{
		AWSBedrock: &AWSBedrockProviderConfig{
			AccessToken: "foobar",
			Endpoint:    "https://example.com",
			Region:      "us-west-2",
		},
		AzureOpenAI: &AzureOpenAIProviderConfig{
			AccessToken: "foobar",
			Endpoint:    "https://example.com",
			User:        "example-user",
		},
		OpenAICompatible: &OpenAICompatibleProviderConfig{
			Endpoints: []OpenAICompatibleEndpoint{
				{
					URL:         "https://example.com/v1",
					AccessToken: "foobar",
				},
			},
			EnableVerboseLogs: true,
		},
		GenericProvider: &GenericProviderConfig{
			ServiceName: GenericServiceProviderAnthropic,
			AccessToken: "foobar",
			Endpoint:    "https://example.com",
		},
		SourcegraphProvider: &SourcegraphProviderConfig{
			AccessToken: "foobar",
			Endpoint:    "https://example.com/api/cody",
		},
	}

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal ServerSideProviderConfig: %v", err)
	}

	expected := "{}"
	actual := string(jsonBytes)

	if actual != expected {
		t.Errorf("Expected JSON: %s, but got: %s", expected, actual)
	}
}

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
