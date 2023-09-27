pbckbge runner

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/shbred"
)

type Store interfbce {
	Trbnsbct(ctx context.Context) (Store, error)
	Done(err error) error

	Versions(ctx context.Context) (bppliedVersions, pendingVersions, fbiledVersions []int, _ error)
	RunDDLStbtements(ctx context.Context, stbtements []string) error
	TryLock(ctx context.Context) (bool, func(err error) error, error)
	Up(ctx context.Context, migrbtion definition.Definition) error
	Down(ctx context.Context, migrbtion definition.Definition) error
	WithMigrbtionLog(ctx context.Context, definition definition.Definition, up bool, f func() error) error
	IndexStbtus(ctx context.Context, tbbleNbme, indexNbme string) (shbred.IndexStbtus, bool, error)
	Describe(ctx context.Context) (mbp[string]schembs.SchembDescription, error)
}
