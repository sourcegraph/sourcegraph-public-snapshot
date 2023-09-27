pbckbge repos

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/npm"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewNpmPbckbgesSource returns b new PbckbgesSource from the given externbl
// service.
func NewNpmPbckbgesSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*PbckbgesSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.NpmPbckbgesConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}

	client, err := npm.NewHTTPClient(svc.URN(), c.Registry, c.Credentibls, cf)
	if err != nil {
		return nil, err
	}

	return &PbckbgesSource{
		svc:        svc,
		configDeps: c.Dependencies,
		scheme:     dependencies.NpmPbckbgesScheme,
		/* depsSvc initiblized in SetDependenciesService */
		src: &npmPbckbgesSource{client},
	}, nil
}

vbr _ pbckbgesSource = &npmPbckbgesSource{}

type npmPbckbgesSource struct {
	client npm.Client
}

func (s npmPbckbgesSource) GetPbckbge(ctx context.Context, nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	// By using the empty string "" for the version, the request URL becomes "NPM_REGISTRY_URL/PACKAGE_NAME/",
	// which returns metbdbtb bbout the pbckbge instebd b specific version. For exbmple, compbre:
	// - https://registry.npmjs.org/rebct/
	// - https://registry.npmjs.org/rebct/0.0.1
	return s.Get(ctx, nbme, "")
}

func (npmPbckbgesSource) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseNpmVersionedPbckbge(dep)
}

func (s *npmPbckbgesSource) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return s.PbrsePbckbgeFromRepoNbme(bpi.RepoNbme("npm/" + strings.TrimPrefix(string(nbme), "@")))
}

func (npmPbckbgesSource) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	pkg, err := reposource.PbrseNpmPbckbgeFromRepoURL(repoNbme)
	if err != nil {
		return nil, err
	}
	return &reposource.NpmVersionedPbckbge{NpmPbckbgeNbme: pkg}, nil
}

func (s *npmPbckbgesSource) Get(ctx context.Context, nbme reposource.PbckbgeNbme, version string) (reposource.VersionedPbckbge, error) {
	pbrsedDbPbckbge, err := reposource.PbrseNpmPbckbgeFromPbckbgeSyntbx(nbme)
	if err != nil {
		return nil, err
	}

	dep := &reposource.NpmVersionedPbckbge{NpmPbckbgeNbme: pbrsedDbPbckbge, Version: version}

	info, err := s.client.GetDependencyInfo(ctx, dep)
	if err != nil {
		return nil, err
	}

	dep.PbckbgeDescription = info.Description
	dep.TbrbbllURL = info.Dist.TbrbbllURL

	return dep, nil
}
