package ratelimit

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type handler struct {
	codeHostStore  database.CodeHostStore
	redisKeyPrefix string
	kv             redispool.KeyValue
}

func (h *handler) Handle(ctx context.Context, observationCtx *observation.Context) error {
	codeHosts, next, err := h.codeHostStore.List(ctx, database.ListCodeHostsOpts{
		LimitOffset: database.LimitOffset{
			Limit: 10,
		},
	})
	for {
		for _, codeHost := range codeHosts {
			// only the last error gets recorded, but we don't want to stop on the first error, since other code hosts could be
			// needing config updates.
			err = h.processCodeHost(ctx, codeHost.URL)
			if err != nil {
				observationCtx.Logger.Error("error setting rate limit configuration", log.String("url", codeHost.URL), log.Error(err))
			}
		}
		codeHosts, next, err = h.codeHostStore.List(ctx, database.ListCodeHostsOpts{
			LimitOffset: database.LimitOffset{
				Limit: 10,
			},
			Cursor: next,
		})

		// TODO: @varsanojidan fix this when next is fixed
		if next <= 0 {
			break
		}
	}
	return err
}

func (h *handler) processCodeHost(ctx context.Context, codeHostURL string) error {
	// Retrieve all the code host rate limit config keys in Redis.
	apiCapKey, apiReplenishmentKey, gitCapKey, gitReplenishmentKey := redispool.GetCodeHostRateLimiterConfigKeys(h.redisKeyPrefix, codeHostURL)

	// Retrieve the actual rate limit values from the source of truth (Postgres).
	ch, err := h.codeHostStore.GetByURL(ctx, codeHostURL)
	if err != nil {
		return errors.Wrapf(err, "rate limit config worker unable to get code host by URL: %s", codeHostURL)
	}

	// Set all of the rate limit config options in Redis.
	if ch.APIRateLimitQuota != nil {
		err = h.kv.Set(apiCapKey, *ch.APIRateLimitQuota)
		if err != nil {
			return errors.Wrapf(err, "rate limit config worker unable to set config key: %s for code host URL: %s", apiCapKey, codeHostURL)

		}
	}
	if ch.APIRateLimitIntervalSeconds != nil {
		err = h.kv.Set(apiReplenishmentKey, *ch.APIRateLimitIntervalSeconds)
		if err != nil {
			return errors.Wrapf(err, "rate limit config worker unable to set config key: %s for code host URL: %s", apiReplenishmentKey, codeHostURL)
		}
	}
	if ch.GitRateLimitQuota != nil {
		err = h.kv.Set(gitCapKey, *ch.GitRateLimitQuota)
		if err != nil {
			return errors.Wrapf(err, "rate limit config worker unable to set config key: %s for code host URL: %s", gitCapKey, codeHostURL)
		}
	}
	if ch.GitRateLimitIntervalSeconds != nil {
		err = h.kv.Set(gitReplenishmentKey, *ch.GitRateLimitIntervalSeconds)
		if err != nil {
			return errors.Wrapf(err, "rate limit config worker unable to set config key: %s for code host URL: %s", gitReplenishmentKey, codeHostURL)
		}
	}
	return nil
}
