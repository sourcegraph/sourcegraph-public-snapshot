package cody

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/accesstoken"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/hashutil"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RefreshGatewayRateLimits refreshes the rate limits for the user on Cody Gateway.
func RefreshGatewayRateLimits(ctx context.Context, userID int32, db database.DB) error {
	completionsConfig := conf.GetCompletionsConfig(conf.Get().SiteConfig())
	// We don't need to do anything if the target is not Cody Gateway, but it's not an error either.
	if completionsConfig.Provider != conftypes.CompletionsProviderNameSourcegraph {
		return nil
	}

	apiTokenSha256, err := db.AccessTokens().GetOrCreateInternalToken(ctx, userID, []string{"user:all"})
	if err != nil {
		return errors.Wrap(err, "getting internal access token")
	}
	gatewayToken := accesstoken.DotcomUserGatewayAccessTokenPrefix + hex.EncodeToString(hashutil.ToSHA256Bytes(apiTokenSha256))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, completionsConfig.Endpoint+"/v1/limits/refresh", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gatewayToken))

	resp, err := httpcli.UncachedExternalDoer.Do(req)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("non-200 response refreshing Gateway limits: %d", resp.StatusCode)
	}

	return nil
}
