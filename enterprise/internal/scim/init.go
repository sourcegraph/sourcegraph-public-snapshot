package scim

import (
	"context"
	"net/http"
	"strings"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Init sets SCIMHandler to a real handler.
func Init(ctx context.Context, observationCtx *observation.Context, db database.DB, _ codeintel.Services, _ conftypes.UnifiedWatchable, s *enterprise.Services) error {
	s.SCIMHandler = NewHandler(ctx, db, observationCtx)

	return nil
}

// NewHandler creates and returns a new SCIM 2.0 handler.
func NewHandler(ctx context.Context, db database.DB, observationCtx *observation.Context) http.Handler {
	config := scim.ServiceProviderConfig{
		DocumentationURI: optional.NewString("www.example.com/scim"),
	}

	var userResourceHandler = NewUserResourceHandler(ctx, observationCtx, db)

	resourceTypes := []scim.ResourceType{createUserResourceType(userResourceHandler)}

	server := scim.Server{
		Config:        config,
		ResourceTypes: resourceTypes,
	}

	// wrap server into logger handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if conf.Get().ScimAuthToken != strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/.api/scim")
		observationCtx.Logger.Error("SCIM request", logger.String("method", r.Method), logger.String("path", r.URL.Path)) // TODO for debugging
		server.ServeHTTP(w, r)
	})

	return handler
}
