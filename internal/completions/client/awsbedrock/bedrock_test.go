package awsbedrock

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/require"
)


func Test_BedrockProvisionedThroughputModel(t *testing.T) {
	tests := []struct {
		want           string
		endpoint       string
		model          string
		fallbackRegion string
		stream         bool
	}{
		{
			want:           "https://bedrock-runtime.us-west-2.amazonaws.com/model/amazon.titan-text-express-v1/invoke",
			endpoint:       "",
			model:          "amazon.titan-text-express-v1",
			fallbackRegion: "us-west-2",
			stream:         false,
		},
		{
			want:           "https://bedrock-runtime.us-west-2.amazonaws.com/model/anthropic.claude-3-sonnet-20240229-v1:0:200k/invoke",
			endpoint:       "",
			model:          "anthropic.claude-3-sonnet-20240229-v1:0:200k",
			fallbackRegion: "us-west-2",
			stream:         false,
		},
		{
			want:           "https://vpce-12345678910.bedrock-runtime.us-west-2.vpce.amazonaws.com/model/arn%3Aaws%3Abedrock%3Aus-west-2%3A012345678901%3Aprovisioned-model%2Fabcdefghijkl/invoke-with-response-stream",
			endpoint:       "https://vpce-12345678910.bedrock-runtime.us-west-2.vpce.amazonaws.com",
			model:          "anthropic.claude-instant-v1/arn:aws:bedrock:us-west-2:012345678901:provisioned-model/abcdefghijkl",
			fallbackRegion: "us-east-1",
			stream:         true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.want), func(t *testing.T) {
			got := buildApiUrl(tt.endpoint, tt.model, tt.stream, tt.fallbackRegion)
			if got.String() != tt.want {
				t.Logf("got %q but wanted %q", got, tt.want)
				t.Fail()
			}
		})
	}
}

func Test_AwsConfigOptsForKeyConfig(t *testing.T) {

	t.Run("With endpoint as URL", func(t *testing.T) {
		endpoint := "https://example.com"
		accessToken := "key:secret"

		defaultConfig, err := config.LoadDefaultConfig(context.Background(), awsConfigOptsForKeyConfig(endpoint, accessToken)...)
		require.NoError(t, err)
		// The endpoint resolver should be set if the endpoint is a URL
		require.NotNil(t, defaultConfig.EndpointResolverWithOptions)
		// The endpoint for any service should be the URL
		awsEndpoint, err := defaultConfig.EndpointResolverWithOptions.ResolveEndpoint("test", "some-region", nil)
		require.NoError(t, err)
		require.Equal(t, awsEndpoint.URL, endpoint)

	})

	t.Run("With endpoint as region", func(t *testing.T) {
		endpoint := "us-east-1"
		accessToken := "key:secret"

		defaultConfig, err := config.LoadDefaultConfig(context.Background(), awsConfigOptsForKeyConfig(endpoint, accessToken)...)
		require.NoError(t, err)
		// The endpoint resolver should not be set if the endpoint is a region
		require.Nil(t, defaultConfig.EndpointResolverWithOptions)
		// The region should be set if the endpoint is a region
		require.Equal(t, defaultConfig.Region, endpoint)

	})

}
