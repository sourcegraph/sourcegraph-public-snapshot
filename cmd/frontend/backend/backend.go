package backend

import (
	"context"

	"github.com/inconshreveable/log15"
)

// Warmup runs some backend functions which fill the cache. This is a kludge,
// we have work to remove this need.
func Warmup(ctx context.Context) error {
	// Pre-heat default repo cache. This loads the default repos at startup,
	// so the cache is warm when the first search query comes in. This query
	// can take ~20s on sourcegraph.com, so this avoids users hitting a cold
	// cache. Issue https://github.com/sourcegraph/sourcegraph/issues/20651
	// tracks removing this again.
	//
	// TODO: Remove this.
	if _, err := Repos.ListDefault(ctx); err != nil {
		log15.Error("Failed to pre-populate default repos cache", "err", err)
	}

	return nil
}
