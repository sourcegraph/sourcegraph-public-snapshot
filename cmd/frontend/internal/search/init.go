package search

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/httpapi"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
)

func LoadConfig() {
	search.ObjectStorageConfigInst.Load()
}

// Init initializes the given enterpriseServices to include the required resolvers for search.
func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	if !exhaustive.IsEnabled(conf.Get()) {
		return nil
	}
	logger := observationCtx.Logger
	store := store.New(db, observationCtx)

	uploadStore, err := search.NewObjectStorage(ctx, observationCtx, search.ObjectStorageConfigInst)
	if err != nil {
		return err
	}

	searchClient := client.New(logger, db, gitserver.NewClient("http.search"))
	newSearcher := service.FromSearchClient(searchClient)

	svc := service.New(observationCtx, store, uploadStore, newSearcher)

	enterpriseServices.SearchJobsResolver = resolvers.New(logger, db, svc)
	enterpriseServices.SearchJobsDataExportHandler = httpapi.ServeSearchJobDownload(logger, svc)
	enterpriseServices.SearchJobsLogsHandler = httpapi.ServeSearchJobLogs(logger, svc)

	return nil
}
