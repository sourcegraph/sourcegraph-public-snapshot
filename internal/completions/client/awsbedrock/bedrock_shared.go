package awsbedrock

import (
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

const (
	clientID = "sourcegraph/1.0"
)

const BedrockAnthropic = "aws-bedrock-anthropic"
const BedrockTitan = "aws-bedrock-titan"

func NewClient(cli httpcli.Doer, endpoint, accessToken string, tokenManager tokenusage.Manager) types.CompletionsClient {
	return &awsBedrockAnthropicCompletionStreamClient{
		cli:          cli,
		accessToken:  accessToken,
		endpoint:     endpoint,
		tokenManager: tokenManager,
	}
}

type awsEventStreamPayload struct {
	Bytes []byte `json:"bytes"`
}

func awsConfigOptsForKeyConfig(endpoint string, accessToken string) []func(*config.LoadOptions) error {
	configOpts := []func(*config.LoadOptions) error{}
	if endpoint != "" {
		apiURL, err := url.Parse(endpoint)
		if err != nil || apiURL.Scheme == "" { // this is not a url assume it is a region
			configOpts = append(configOpts, config.WithRegion(endpoint))
		} else { // this is a url just use it directly
			configOpts = append(configOpts, config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: endpoint}, nil
				})))
		}
	}

	// We use the accessToken field to provide multiple values.
	// If it consists of two parts, separated by a `:`, the first part is
	// the aws access key, and the second is the aws secret key.
	// If there are three parts, the third part is the aws session token.
	// If no access token is given, we default to the AWS default credential provider
	// chain, which supports all basic known ways of connecting to AWS.
	if accessToken != "" {
		parts := strings.SplitN(accessToken, ":", 3)
		if len(parts) == 2 {
			configOpts = append(configOpts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(parts[0], parts[1], "")))
		} else if len(parts) == 3 {
			configOpts = append(configOpts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(parts[0], parts[1], parts[2])))
		}
	}

	return configOpts
}

func removeWhitespaceOnlySequences(sequences []string) []string {
	var result []string
	for _, sequence := range sequences {
		if len(strings.TrimSpace(sequence)) > 0 {
			result = append(result, sequence)
		}
	}
	return result
}
