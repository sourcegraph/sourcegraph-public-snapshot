package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getLimitsURL(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	u.Path = "v1/limits"
	return u.String(), nil
}

func getLimitsRequest(cc *conftypes.CompletionsConfig, ec *conftypes.EmbeddingsConfig) (*http.Request, error) {
	// It's possible the user is only using sourcegraph gateway for completions or embeddings
	// make sure to get the url/token for the sourcegraph provider
	// start with the embeddings since there are fewer options
	endPoint := ec.Endpoint
	token := ec.AccessToken
	if ec.Provider != conftypes.EmbeddingsProviderNameSourcegraph {
		endPoint = cc.Endpoint
		token = cc.AccessToken
	}
	url, err := getLimitsURL(endPoint)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return req, nil
}

func (r *appResolver) CodyGatewayRateLimitStatus(ctx context.Context) ([]graphqlbackend.RateLimitStatus, error) {
	// 🚨 SECURITY: Only site admins may check rate limits.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	config := conf.Get().SiteConfig()
	cc := conf.GetCompletionsConfig(config)
	ec := conf.GetEmbeddingsConfig(config)

	// If the user doesn't have an dotcom auth token
	// or isn't using the cody gateway, there are no limits
	if (config.App == nil || len(config.App.DotcomAuthToken) == 0) ||
		(cc.Provider != conftypes.CompletionsProviderNameSourcegraph &&
			ec.Provider != conftypes.EmbeddingsProviderNameSourcegraph) {
		return []graphqlbackend.RateLimitStatus{}, nil
	}

	req, err := getLimitsRequest(cc, ec)
	if err != nil {
		return nil, err
	}
	resp, err := r.doer.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("request failed with status: %d", errors.Safe(resp.StatusCode))
	}
	var featureLimits map[string]rateLimit
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&featureLimits); err != nil {
		return nil, err
	}

	rateLimits := make([]graphqlbackend.RateLimitStatus, 0, len(featureLimits))
	for featureName, limit := range featureLimits {
		rateLimits = append(rateLimits, &codyRateLimit{
			feature:   featureName,
			rateLimit: limit,
		})
	}

	sort.Slice(rateLimits, func(i, j int) bool {
		return rateLimits[i].Feature() < rateLimits[j].Feature()
	})

	return rateLimits, nil

}

type rateLimit struct {
	IntervalLimit int64      `json:"limit"`
	IntervalUsage int64      `json:"usage"`
	TimeInterval  string     `json:"interval"`
	Expiry        *time.Time `json:"expiry"`
}

var featureDisplayNames map[string]string = map[string]string{"chat_completions": "Chat", "code_completions": "Autocomplete", "embeddings": "Embeddings"}

type codyRateLimit struct {
	feature string
	rateLimit
}

func (c *codyRateLimit) Feature() string {
	display, ok := featureDisplayNames[c.feature]
	if !ok {
		return c.feature
	}
	return display

}

func (c *codyRateLimit) Limit() graphqlbackend.BigInt {
	return graphqlbackend.BigInt(c.IntervalLimit)
}

func (c *codyRateLimit) Usage() graphqlbackend.BigInt {
	return graphqlbackend.BigInt(c.IntervalUsage)
}

func (c *codyRateLimit) PercentUsed() int32 {
	if c.IntervalLimit == 0 {
		return 100
	}
	return int32(math.Ceil(float64(c.IntervalUsage) / float64(c.IntervalLimit) * 100))
}

func (c *codyRateLimit) NextLimitReset() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(c.Expiry)
}

func (c *codyRateLimit) Interval() string {
	return c.TimeInterval
}
