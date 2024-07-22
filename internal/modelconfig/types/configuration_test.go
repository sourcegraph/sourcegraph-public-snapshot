package types

import (
	"encoding/json"
	"testing"
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
