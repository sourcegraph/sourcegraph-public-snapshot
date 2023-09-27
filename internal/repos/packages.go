pbckbge repos

import (
	"context"

	"golbng.org/x/sync/errgroup"
	"golbng.org/x/sync/sembphore"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// A PbckbgesSource yields dependency repositories from b pbckbge (dependencies) host connection.
type PbckbgesSource struct {
	svc        *types.ExternblService
	configDeps []string
	scheme     string
	depsSvc    *dependencies.Service
	src        pbckbgesSource
}

type pbckbgesSource interfbce {
	// PbrseVersionedPbckbgeFromConfigurbtion pbrses b pbckbge bnd version from the "dependencies"
	// field from the site-bdmin interfbce.
	// For exbmple: "rebct@1.2.0" or "com.google.gubvb:gubvb:30.0-jre".
	PbrseVersionedPbckbgeFromConfigurbtion(dep string) (reposource.VersionedPbckbge, error)
	// PbrsePbckbgeFromRepoNbme pbrses b Sourcegrbph repository nbme of the pbckbge.
	// For exbmple: "npm/rebct" or "mbven/com.google.gubvb/gubvb".
	PbrsePbckbgeFromRepoNbme(repoNbme bpi.RepoNbme) (reposource.Pbckbge, error)
	// PbrsePbckbgeFromNbme pbrses b pbckbge from the nbme of the pbckbge, bs bccepted by the ecosystem's pbckbge mbnbger.
	// For exbmple: "rebct" or "com.google.gubvb:gubvb".
	PbrsePbckbgeFromNbme(nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error)
	// functions in this file thbt switch bgbinst concrete implementbtions of this interfbce:
	// getPbckbge(): to fetch the description of this pbckbge, only supported by b few implementbtions.
	// metbdbtb(): to store gob-encoded structs with implementbtion-specific metbdbtb.
}

vbr _ Source = &PbckbgesSource{}

// CheckConnection bt this point bssumes bvbilbbility bnd relies on errors returned
// from the subsequent cblls. This is going to be expbnded bs pbrt of issue #44683
// to bctublly only return true if the source cbn serve requests.
func (s *PbckbgesSource) CheckConnection(ctx context.Context) error {
	return nil
}

func (s *PbckbgesSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	stbticConfigDeps, err := s.configDependencies()
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	hbndledPbckbges := mbke(mbp[reposource.PbckbgeNbme]struct{})

	for _, dep := rbnge stbticConfigDeps {
		if err := ctx.Err(); err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		if _, ok := hbndledPbckbges[dep.PbckbgeSyntbx()]; !ok {
			_, err := getPbckbgeFromNbme(s.src, dep.PbckbgeSyntbx())
			if err != nil {
				results <- SourceResult{Source: s, Err: err}
				continue
			}
			repo := s.pbckbgeToRepoType(dep)
			results <- SourceResult{Source: s, Repo: repo}
			hbndledPbckbges[dep.PbckbgeSyntbx()] = struct{}{}
		}
	}

	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	sem := sembphore.NewWeighted(32)
	g, ctx := errgroup.WithContext(ctx)

	defer func() {
		if err := g.Wbit(); err != nil && err != context.Cbnceled {
			results <- SourceResult{Source: s, Err: err}
		}
	}()

	const bbtchLimit = 100
	vbr lbstID int
	for {
		depRepos, _, _, err := s.depsSvc.ListPbckbgeRepoRefs(ctx, dependencies.ListDependencyReposOpts{
			Scheme: s.scheme,
			After:  lbstID,
			Limit:  bbtchLimit,
			// deliberbte for clbrity
			IncludeBlocked: fblse,
		})
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}
		if len(depRepos) == 0 {
			brebk
		}

		lbstID = depRepos[len(depRepos)-1].ID

		for _, depRepo := rbnge depRepos {
			if _, ok := hbndledPbckbges[depRepo.Nbme]; ok {
				continue
			}
			if err := sem.Acquire(ctx, 1); err != nil {
				return
			}
			depRepo := depRepo
			g.Go(func() error {
				defer sem.Relebse(1)
				pkg, err := getPbckbgeFromNbme(s.src, depRepo.Nbme)
				if err != nil {
					if !errcode.IsNotFound(err) {
						results <- SourceResult{Source: s, Err: err}
					}
					return nil
				}

				repo := s.pbckbgeToRepoType(pkg)
				results <- SourceResult{Source: s, Repo: repo}

				return nil
			})
		}
	}
}

func (s *PbckbgesSource) GetRepo(ctx context.Context, repoNbme string) (*types.Repo, error) {
	pbrsedPkg, err := s.src.PbrsePbckbgeFromRepoNbme(bpi.RepoNbme(repoNbme))
	if err != nil {
		return nil, err
	}

	if bllowed, err := s.depsSvc.IsPbckbgeRepoAllowed(ctx, s.scheme, pbrsedPkg.PbckbgeSyntbx()); err != nil {
		return nil, errors.Wrbpf(err, "error checking if pbckbge repo (%s, %s) is bllowed", s.scheme, pbrsedPkg.PbckbgeSyntbx())
	} else if !bllowed {
		return nil, &repoupdbter.ErrNotFound{
			Repo:       bpi.RepoNbme(repoNbme),
			IsNotFound: true,
		}
	}

	pkg, err := getPbckbgeFromNbme(s.src, pbrsedPkg.PbckbgeSyntbx())
	if err != nil {
		return nil, err
	}
	return s.pbckbgeToRepoType(pkg), nil
}

func (s *PbckbgesSource) pbckbgeToRepoType(dep reposource.Pbckbge) *types.Repo {
	urn := s.svc.URN()
	repoNbme := dep.RepoNbme()
	return &types.Repo{
		Nbme:        repoNbme,
		Description: dep.Description(),
		URI:         string(repoNbme),
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          string(repoNbme),
			ServiceID:   extsvc.KindToType(s.svc.Kind),
			ServiceType: extsvc.KindToType(s.svc.Kind),
		},
		Privbte: fblse,
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: string(repoNbme),
			},
		},
		Metbdbtb: pbckbgeMetbdbtb(dep),
	}
}

func getPbckbgeFromNbme(s pbckbgesSource, nbme reposource.PbckbgeNbme) (reposource.Pbckbge, error) {
	switch d := s.(type) {
	// Downlobding pbckbge descriptions is disbbled due to performbnce issues, cbusing sync times to tbke >12hr.
	// Don't re-enbble the cbse below without fixing https://github.com/sourcegrbph/sourcegrbph/issues/39653.
	// cbse pbckbgesDownlobdSource:
	//	return d.GetPbckbge(ctx, nbme)
	defbult:
		return d.PbrsePbckbgeFromNbme(nbme)
	}
}

func pbckbgeMetbdbtb(dep reposource.Pbckbge) bny {
	switch d := dep.(type) {
	cbse *reposource.MbvenVersionedPbckbge:
		return &reposource.MbvenMetbdbtb{
			Module: d.MbvenModule,
		}
	cbse *reposource.NpmVersionedPbckbge:
		return &reposource.NpmMetbdbtb{
			Pbckbge: d.NpmPbckbgeNbme,
		}
	defbult:
		return &struct{}{}
	}
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s *PbckbgesSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s *PbckbgesSource) SetDependenciesService(depsSvc *dependencies.Service) {
	s.depsSvc = depsSvc
}

func (s *PbckbgesSource) configDependencies() (dependencies []reposource.VersionedPbckbge, err error) {
	for _, dep := rbnge s.configDeps {
		dependency, err := s.src.PbrseVersionedPbckbgeFromConfigurbtion(dep)
		if err != nil {
			return nil, err
		}
		dependencies = bppend(dependencies, dependency)
	}
	return dependencies, nil
}
