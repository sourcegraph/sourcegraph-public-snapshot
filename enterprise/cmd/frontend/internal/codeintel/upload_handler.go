package codeintel

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/httpapi"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// NewCodeIntelUploadHandler creates a new code intel LSIF upload HTTP handler. This is used
// by both the enterprise frontend codeintel init code to install handlers in the frontend API
// as well as the the enterprise frontend executor init code to install handlers in the proxy.
func NewCodeIntelUploadHandler(ctx context.Context, conf conftypes.SiteConfigQuerier, db dbutil.DB, internal bool, services *Services) (http.Handler, error) {
	return httpapi.NewUploadHandler(
		db,
		&httpapi.DBStoreShim{Store: services.dbStore},
		services.uploadStore,
		internal,
		httpapi.DefaultValidatorByCodeHost,
	), nil
}
