package batchescodemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	batchesresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers"
	codemonitorsresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codemonitors/resolvers"
	batchesstore "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(
	ctx context.Context,
	db database.DB,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
	observationContext *observation.Context,
) error {
	enterpriseServices.BatchChangesCodeMonitorsResolver = &resolver{
		edb:    edb.NewEnterpriseDB(db),
		bstore: batchesstore.New(db, observationContext, keyring.Default().BatchChangesCredentialKey),
	}

	return nil
}

type resolver struct {
	bstore *batchesstore.Store
	edb    edb.EnterpriseDB
}

func (r *resolver) BatchChangeCodeMonitor(ctx context.Context, monitorID int64) (graphqlbackend.MonitorResolver, error) {
	return codemonitorsresolvers.MonitorByID(ctx, r.edb, monitorID)
}

func (r *resolver) CodeMonitorBatchChange(ctx context.Context, batchChangeID int64) (graphqlbackend.BatchChangeResolver, error) {
	return batchesresolvers.BatchChangeByID(ctx, r.bstore, batchChangeID)
}
