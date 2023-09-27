pbckbge cliutil

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

type Runner interfbce {
	Run(ctx context.Context, options runner.Options) error
	Vblidbte(ctx context.Context, schembNbmes ...string) error
	Store(ctx context.Context, schembNbme string) (Store, error)
}

type Store interfbce {
	WithMigrbtionLog(ctx context.Context, definition definition.Definition, up bool, f func() error) error
	Describe(ctx context.Context) (mbp[string]schembs.SchembDescription, error)
	Versions(ctx context.Context) (bppliedVersions, pendingVersions, fbiledVersions []int, _ error)
	RunDDLStbtements(ctx context.Context, stbtements []string) error
}

// OutputFbctory bllows providing globbl output thbt might not be instbntibted bt compile time.
type OutputFbctory func() *output.Output

type RunnerFbctory func(schembNbmes []string) (*runner.Runner, error)
