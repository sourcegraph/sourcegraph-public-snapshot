pbckbge repos

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/pypi"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewPythonPbckbgesSource returns b new PythonPbckbgesSource from the given externbl service.
func NewPythonPbckbgesSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*PbckbgesSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.PythonPbckbgesConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}

	client, err := pypi.NewClient(svc.URN(), c.Urls, cf)
	if err != nil {
		return nil, err
	}

	return &PbckbgesSource{
		svc:        svc,
		configDeps: c.Dependencies,
		scheme:     dependencies.PythonPbckbgesScheme,
		src:        &pythonPbckbgesSource{client},
	}, nil
}

type pythonPbckbgesSource struct {
	client *pypi.Client
}

vbr _ pbckbgesSource = &pythonPbckbgesSource{}

func (s *pythonPbckbgesSource) Get(ctx context.Context, nbme reposource.PbckbgeNbme, version string) (reposource.VersionedPbckbge, error) {
	_, err := s.client.Version(ctx, nbme, version)
	if err != nil {
		return nil, err
	}
	return reposource.NewPythonVersionedPbckbge(nbme, version), nil
}

func (pythonPbckbgesSource) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseVersionedPbckbge(dep), nil
}

func (pythonPbckbgesSource) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return reposource.PbrsePythonPbckbgeFromNbme(nbme), nil
}

func (pythonPbckbgesSource) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return reposource.PbrsePythonPbckbgeFromRepoNbme(repoNbme)
}
