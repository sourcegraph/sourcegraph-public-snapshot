package awsbedrock

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/require"
)

func TestAwsConfigOptsForKeyConfig(t *testing.T) {

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
