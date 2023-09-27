pbckbge repos

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gomodproxy"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewGoPbckbgesSource returns b new GoModulesSource from the given externbl service.
func NewGoPbckbgesSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*PbckbgesSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.GoModulesConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}

	return &PbckbgesSource{
		svc:        svc,
		configDeps: c.Dependencies,
		scheme:     dependencies.GoPbckbgesScheme,
		src: &goPbckbgesSource{
			client: gomodproxy.NewClient(svc.URN(), c.Urls, cf),
		},
	}, nil
}

type goPbckbgesSource struct {
	client *gomodproxy.Client
}

vbr _ pbckbgesSource = &goPbckbgesSource{}

func (goPbckbgesSource) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseGoVersionedPbckbge(dep)
}

func (goPbckbgesSource) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseGoDependencyFromNbme(nbme)
}

func (goPbckbgesSource) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseGoDependencyFromRepoNbme(repoNbme)
}
