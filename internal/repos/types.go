package repos

import (
	"context"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// pick deterministically chooses between a and b a repo to keep and
// discard. It is used when resolving conflicts on sourced repositories.
func pick(a *types.Repo, b *types.Repo) (keep, discard *types.Repo) {
	if a.Less(b) {
		return a, b
	}
	return b, a
}

type externalServiceLister interface {
	List(context.Context, database.ExternalServicesListOptions) ([]*types.ExternalService, error)
}

// RateLimitSyncer syncs rate limits based on external service configuration
type RateLimitSyncer struct {
	registry      *ratelimit.Registry
	serviceLister externalServiceLister
	// How many services to fetch in each DB call
	limit int64
}

// NewRateLimitSyncer returns a new syncer
func NewRateLimitSyncer(registry *ratelimit.Registry, serviceLister externalServiceLister) *RateLimitSyncer {
	r := &RateLimitSyncer{
		registry:      registry,
		serviceLister: serviceLister,
		limit:         500,
	}
	return r
}

// SyncRateLimiters syncs all rate limiters using current config.
// We sync them all as we need to pick the most restrictive configured limit per code host
// and rate limits can be defined in multiple external services for the same host.
func (r *RateLimitSyncer) SyncRateLimiters(ctx context.Context) error {
	byURL := make(map[string]extsvc.RateLimitConfig)

	cursor := database.LimitOffset{
		Limit: int(r.limit),
	}

	for {
		services, err := r.serviceLister.List(ctx, database.ExternalServicesListOptions{
			LimitOffset: &cursor,
		})
		if err != nil {
			return errors.Wrap(err, "listing external services")
		}

		if len(services) == 0 {
			break
		}

		cursor.Offset += len(services)

		for _, svc := range services {
			rlc, err := extsvc.ExtractRateLimitConfig(svc.Config, svc.Kind, svc.DisplayName)
			if err != nil {
				if _, ok := err.(extsvc.ErrRateLimitUnsupported); ok {
					continue
				}
				return errors.Wrap(err, "getting rate limit configuration")
			}

			current, ok := byURL[rlc.BaseURL]
			if !ok || (ok && current.IsDefault) {
				byURL[rlc.BaseURL] = rlc
				continue
			}
			// Use the lower limit, but a default value should not override
			// a limit that has been configured
			if rlc.Limit < current.Limit && !rlc.IsDefault {
				byURL[rlc.BaseURL] = rlc
			}
		}

		if len(services) < int(r.limit) {
			break
		}
	}

	for u, rl := range byURL {
		l := r.registry.Get(u)
		l.SetLimit(rl.Limit)
	}

	return nil
}

type ScopeCache interface {
	Get(string) ([]byte, bool)
	Set(string, []byte)
}

// GrantedScopes returns a slice of scopes granted by the service based on the token
// provided in the config. It makes a request to the code host.
//
// Currently only GitHub is supported.
func GrantedScopes(ctx context.Context, cache ScopeCache, kind string, rawConfig string) ([]string, error) {
	if kind != extsvc.KindGitHub {
		return nil, fmt.Errorf("only GitHub supported")
	}
	config, err := extsvc.ParseConfig(kind, rawConfig)
	if err != nil {
		return nil, errors.Wrap(err, "parsing config")
	}
	switch v := config.(type) {
	case *schema.GitHubConnection:
		if v.Token == "" {
			return nil, errors.New("missing token")
		}
		key, err := hashToken(v.Token)
		if err != nil {
			return nil, err
		}
		if result, ok := cache.Get(key); ok && len(result) > 0 {
			return strings.Split(string(result), ","), nil
		}

		// Slow path
		u, err := url.Parse(v.Url)
		if err != nil {
			return nil, errors.Wrap(err, "parsing URL")
		}
		client := github.NewV3Client(u, &auth.OAuthBearerToken{Token: v.Token}, nil)
		scopes, err := client.GetAuthenticatedUserOAuthScopes(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "getting scopes")
		}
		cache.Set(key, []byte(strings.Join(scopes, ",")))
		return scopes, nil
	default:
		return nil, fmt.Errorf("unsupported config type: %T", v)
	}
}

func hashToken(token string) (string, error) {
	h := fnv.New32()
	_, err := h.Write([]byte(token))
	if err != nil {
		return "", errors.Wrap(err, "hashing token")
	}
	b := h.Sum(nil)
	return hex.EncodeToString(b), nil
}
