package repos

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// NewSourcer returns a Sourcer that converts the given ExternalService
// into a Source that uses the provided httpcli.Factory to create the
// http.Clients needed to contact the respective upstream code host APIs.
//
// The provided decorator functions will be applied to the Source.
func NewSourcer(logger log.Logger, db database.DB, cf *httpcli.Factory, decs ...func(repos.Source) repos.Source) repos.Sourcer {
	return func(ctx context.Context, svc *types.ExternalService) (repos.Source, error) {
		src, err := repos.NewSource(ctx, logger.Scoped("source", ""), db, svc, cf)
		if err != nil {
			return nil, err
		}

		for _, dec := range decs {
			src = dec(src)
		}

		return src, nil
	}
}
