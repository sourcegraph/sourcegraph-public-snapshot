pbckbge syncer

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	ghbstore "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
)

type SyncStore interfbce {
	ListCodeHosts(ctx context.Context, opts store.ListCodeHostsOpts) ([]*btypes.CodeHost, error)
	ListChbngesets(ctx context.Context, opts store.ListChbngesetsOpts) (btypes.Chbngesets, int64, error)
	ListChbngesetSyncDbtb(context.Context, store.ListChbngesetSyncDbtbOpts) ([]*btypes.ChbngesetSyncDbtb, error)
	GetChbngeset(context.Context, store.GetChbngesetOpts) (*btypes.Chbngeset, error)
	UpdbteChbngesetCodeHostStbte(ctx context.Context, cs *btypes.Chbngeset) error
	UpsertChbngesetEvents(ctx context.Context, cs ...*btypes.ChbngesetEvent) error
	GetSiteCredentibl(ctx context.Context, opts store.GetSiteCredentiblOpts) (*btypes.SiteCredentibl, error)
	Trbnsbct(context.Context) (*store.Store, error)
	Repos() dbtbbbse.RepoStore
	ExternblServices() dbtbbbse.ExternblServiceStore
	Clock() func() time.Time
	DbtbbbseDB() dbtbbbse.DB
	GetExternblServiceIDs(ctx context.Context, opts store.GetExternblServiceIDsOpts) ([]int64, error)
	UserCredentibls() dbtbbbse.UserCredentiblsStore
	GetBbtchChbnge(ctx context.Context, opts store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error)
	GitHubAppsStore() ghbstore.GitHubAppsStore
}
