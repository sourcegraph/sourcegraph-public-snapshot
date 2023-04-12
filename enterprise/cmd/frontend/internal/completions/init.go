package completions

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	_ database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	enterpriseServices.CompletionsResolver = &resolver{}
	enterpriseServices.NewCompletionsStreamHandler = codyAPIRedirectorFactory
	return nil
}

type resolver struct{}

func (c *resolver) Completions(ctx context.Context, args graphqlbackend.CompletionsArgs) (string, error) {
	// Gate via feature flags on Sourcegraph.com
	if envvar.SourcegraphDotComMode() && !cody.IsCodyExperimentalFeatureFlagEnabled(ctx) {
		return "", errors.New("cody experimental feature flag is not enabled for current user")
	}

	base, ok := codyURL()
	if !ok {
		return "", errors.New("no cody-api address are configured")
	}
	url := base + "/completions/final"

	content, _ := json.Marshal(graphqlbackend.CompletionsInput{
		Messages:          args.Input.Messages,
		Temperature:       args.Input.Temperature,
		MaxTokensToSample: args.Input.MaxTokensToSample,
		TopK:              args.Input.TopK,
		TopP:              args.Input.TopP,
	})
	resp, err := http.Post(url, "", bytes.NewReader(content))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Newf("unexpected status code %d", resp.StatusCode)
	}

	var payload struct {
		Completion string `json:"completion"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	return payload.Completion, nil
}

func codyURL() (string, bool) {
	addrs := conf.Get().ServiceConnectionConfig.CodyAPIs
	if len(addrs) == 0 {
		return "", false
	}

	return addrs[0], true
}

func codyAPIRedirectorFactory() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", "/.cody"+r.URL.Path)
		w.WriteHeader(http.StatusPermanentRedirect)
	})
}
