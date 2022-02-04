package compute

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/compute/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(ctx context.Context, db database.DB, _ conftypes.UnifiedWatchable, enterpriseServices *enterprise.Services, observationContext *observation.Context) error {
	enterpriseServices.ComputeResolver = resolvers.NewResolver(db)
	enterpriseServices.NewComputeStreamHandler = newComputeStreamHandler
	return nil
}

// newComputeStreamHandler implements the HTTP endpoint for the Compute stream.
// TODO(rvantonder): #30527
func newComputeStreamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		_, _ = w.Write([]byte("compute stream endpoint unimplemented"))
	})
}
