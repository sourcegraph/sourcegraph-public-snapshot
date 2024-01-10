package syncer

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	ghastore "github.com/sourcegraph/sourcegraph/internal/github_apps/store"
)

type SyncStore interface {
	ListCodeHosts(ctx context.Context, opts store.ListCodeHostsOpts) ([]*btypes.CodeHost, error)
	ListChangesets(ctx context.Context, opts store.ListChangesetsOpts) (btypes.Changesets, int64, error)
	ListChangesetSyncData(context.Context, store.ListChangesetSyncDataOpts) ([]*btypes.ChangesetSyncData, error)
	GetChangeset(context.Context, store.GetChangesetOpts) (*btypes.Changeset, error)
	UpdateChangesetCodeHostState(ctx context.Context, cs *btypes.Changeset) error
	UpsertChangesetEvents(ctx context.Context, cs ...*btypes.ChangesetEvent) error
	GetSiteCredential(ctx context.Context, opts store.GetSiteCredentialOpts) (*btypes.SiteCredential, error)
	Transact(context.Context) (*store.Store, error)
	Repos() database.RepoStore
	ExternalServices() database.ExternalServiceStore
	Clock() func() time.Time
	DatabaseDB() database.DB
	GetExternalServiceIDs(ctx context.Context, opts store.GetExternalServiceIDsOpts) ([]int64, error)
	UserCredentials() database.UserCredentialsStore
	GetBatchChange(ctx context.Context, opts store.GetBatchChangeOpts) (*btypes.BatchChange, error)
	GitHubAppsStore() ghastore.GitHubAppsStore
	GetChangesetSpecByID(ctx context.Context, id int64) (*btypes.ChangesetSpec, error)
}
