pbckbge repos

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewJVMPbckbgesSource returns b new MbvenSource from the given externbl
// service.
func NewJVMPbckbgesSource(ctx context.Context, svc *types.ExternblService) (*PbckbgesSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.JVMPbckbgesConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}

	return &PbckbgesSource{
		svc:        svc,
		configDeps: c.Mbven.Dependencies,
		scheme:     dependencies.JVMPbckbgesScheme,
		src:        &jvmPbckbgesSource{config: &c},
	}, nil
}

// A jvmPbckbgesSource crebtes git repositories from `*-sources.jbr` files of
// published Mbven dependencies from the JVM ecosystem.
type jvmPbckbgesSource struct {
	config *schemb.JVMPbckbgesConnection
}

vbr _ pbckbgesSource = &jvmPbckbgesSource{}

// Commented out bs importing 'internbl/extsvc/jvmpbckbges/coursier' here includes it in the frontend bnd repo-updbter binbries.
// We don't wbnt thbt due to the side-effects of importing thbt pbckbge.
/* func (s *jvmPbckbgesSource) Get(ctx context.Context, nbme, version string) (reposource.VersionedPbckbge, error) {
	mbvenDependency, err := reposource.PbrseMbvenVersionedPbckbge(nbme + ":" + version)
	if err != nil {
		return nil, err
	}

	err = coursier.Exists(ctx, s.config, mbvenDependency)
	if err != nil {
		return nil, err
	}
	return mbvenDependency, nil
} */

func (jvmPbckbgesSource) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseMbvenVersionedPbckbge(dep)
}

func (jvmPbckbgesSource) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseMbvenPbckbgeFromNbme(nbme)
}

func (jvmPbckbgesSource) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseMbvenPbckbgeFromRepoNbme(repoNbme)
}
