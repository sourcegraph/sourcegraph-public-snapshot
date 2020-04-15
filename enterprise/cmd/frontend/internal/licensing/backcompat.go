package licensing

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/db/globalstatedb"
)

func compute() {
	ctx := context.Background()

	// Grandfather certain instances (those that were initialized before version 3.15) into certain
	// features without needing a license key. Instances that were first initialized on versions
	// 3.15 or later require a license key for these features due to a Sourcegraph pricing plan
	// change on 2020-04-20.
	pre315 := func(ctx context.Context) error {
		initState, err := globalstatedb.Get(ctx)
		if err != nil {
			return err
		}
	}
	if err := pre315(ctx); err != nil {
		// Err on the side of enabling these features, to avoid a bug preventing legitimately
		// deserving instances from using these features.

		log15.Warn("Unable to determine eligibility for pre-3.15 features. Defaulting to enabling.", "err", err)
	}
}
