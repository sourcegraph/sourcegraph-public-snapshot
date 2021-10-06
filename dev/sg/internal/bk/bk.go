package bk

import (
	"context"
	"fmt"
	"strings"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// https://buildkite.com/sourcegraph
const buildkiteOrg = "sourcegraph"

type buildkiteSecrets struct {
	Token string `json:"token"`
}

// retrieveToken obtains a token either from the cached configuration or by asking the user for it.
func retrieveToken(ctx context.Context, out *output.Output) (string, error) {
	sec := secrets.FromContext(ctx)
	bkSecrets := buildkiteSecrets{}
	err := sec.Get("buildkite", &bkSecrets)
	if errors.Is(err, secrets.ErrSecretNotFound) {
		str, err := getTokenFromUser(out)
		if err != nil {
			return "", nil
		}
		if err := sec.PutAndSave("buildkite", buildkiteSecrets{Token: str}); err != nil {
			return "", err
		}
		return str, nil
	}
	if err != nil {
		return "", err
	}
	return bkSecrets.Token, nil
}

// getTokenFromUser prompts the user for a slack OAuth token.
func getTokenFromUser(out *output.Output) (string, error) {
	out.WriteLine(output.Linef(output.EmojiLightbulb, output.StylePending, `Please create and copy a new token from https://buildkite.com/user/api-access-tokens with the following scopes:

- Organization access to %q
- read_artifacts
- read_builds
- read_build_logs
- read_pipelines
`, buildkiteOrg))
	fmt.Printf("Paste it here: ")
	var token string
	if _, err := fmt.Scan(&token); err != nil {
		return "", err
	}
	return token, nil
}

type Client struct {
	bk *buildkite.Client
}

func NewClient(ctx context.Context, out *output.Output) (*Client, error) {
	token, err := retrieveToken(ctx, out)
	if err != nil {
		return nil, err
	}
	config, err := buildkite.NewTokenConfig(token, false)
	if err != nil {
		return nil, fmt.Errorf("failed to init buildkite config: %w", err)
	}
	return &Client{bk: buildkite.NewClient(config.Client())}, nil
}

func (c *Client) GetMostRecentBuild(ctx context.Context, pipeline, branch string) (*buildkite.Build, error) {
	builds, _, err := c.bk.Builds.ListByPipeline(buildkiteOrg, pipeline, &buildkite.BuildsListOptions{
		Branch: branch,
	})
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil, errors.New("no build found")
		}
		return nil, err
	}
	if len(builds) == 0 {
		return nil, errors.New("no builds found")
	}
	// Newest is returned first https://buildkite.com/docs/apis/rest-api/builds#list-builds-for-a-pipeline
	return &builds[0], nil
}
