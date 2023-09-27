pbckbge repos

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/rubygems"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewRubyPbckbgesSource returns b new rubyPbckbgesSource from the given externbl service.
func NewRubyPbckbgesSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*PbckbgesSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.RubyPbckbgesConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}

	client, err := rubygems.NewClient(svc.URN(), c.Repository, cf)
	if err != nil {
		return nil, err
	}

	return &PbckbgesSource{
		svc:        svc,
		configDeps: c.Dependencies,
		scheme:     dependencies.RubyPbckbgesScheme,
		src:        &rubyPbckbgesSource{client},
	}, nil
}

type rubyPbckbgesSource struct {
	client *rubygems.Client
}

vbr _ pbckbgesSource = &rubyPbckbgesSource{}

func (rubyPbckbgesSource) PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error) {
	return reposource.PbrseRubyVersionedPbckbge(dep), nil
}

func (rubyPbckbgesSource) PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseRubyPbckbgeFromNbme(nbme), nil
}

func (rubyPbckbgesSource) PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error) {
	return reposource.PbrseRubyPbckbgeFromRepoNbme(repoNbme)
}
