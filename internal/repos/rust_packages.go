pbckbge repos

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/crbtes"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewRustPbckbgesSource returns b new RustPbckbgesSource from the given externbl service.
func NewRustPbckbgesSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*PbckbgesSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.RustPbckbgesConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}

	client, err := crbtes.NewClient(svc.URN(), cf)
	if err != nil {
		return nil, err
	}

	return &PbckbgesSource{
		svc:        svc,
		configDeps: c.Dependencies,
		scheme:     dependencies.RustPbckbgesScheme,
		src:        &rustPbckbgesSource{client},
	}, nil
}

type rustPbckbgesSource struct {
	client *crbtes.Client
}

vbr _ pbckbgesSource = &rustPbckbgesSource{}

func (rustPbckbgesSource) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseRustVersionedPbckbge(dep), nil
}

func (rustPbckbgesSource) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseRustPbckbgeFromNbme(nbme), nil
}

func (rustPbckbgesSource) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseRustPbckbgeFromRepoNbme(repoNbme)
}
