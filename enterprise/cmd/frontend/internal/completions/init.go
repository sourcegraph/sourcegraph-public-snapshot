package completions

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/resolvers"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	_ database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	logger := log.Scoped("completions", "")
	enterpriseServices.NewCompletionsStreamHandler = func() http.Handler { return streaming.NewCompletionsStreamHandler(logger) }
	enterpriseServices.NewCodeCompletionsHandler = func() http.Handler { return streaming.NewCodeCompletionsHandler(logger) }
	enterpriseServices.CompletionsResolver = resolvers.NewCompletionsResolver()

	return nil
}
