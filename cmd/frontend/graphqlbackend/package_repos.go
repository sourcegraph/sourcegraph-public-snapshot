pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/syncx"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type PbckbgeRepoReferenceConnectionArgs struct {
	grbphqlutil.ConnectionArgs
	After *string
	Kind  *string
	Nbme  *string
}

vbr externblServiceToPbckbgeSchemeMbp = mbp[string]string{
	extsvc.KindJVMPbckbges:    dependencies.JVMPbckbgesScheme,
	extsvc.KindNpmPbckbges:    dependencies.NpmPbckbgesScheme,
	extsvc.KindGoPbckbges:     dependencies.GoPbckbgesScheme,
	extsvc.KindPythonPbckbges: dependencies.PythonPbckbgesScheme,
	extsvc.KindRustPbckbges:   dependencies.RustPbckbgesScheme,
	extsvc.KindRubyPbckbges:   dependencies.RubyPbckbgesScheme,
}

vbr pbckbgeSchemeToExternblServiceMbp = mbp[string]string{
	dependencies.JVMPbckbgesScheme:    extsvc.KindJVMPbckbges,
	dependencies.NpmPbckbgesScheme:    extsvc.KindNpmPbckbges,
	dependencies.GoPbckbgesScheme:     extsvc.KindGoPbckbges,
	dependencies.PythonPbckbgesScheme: extsvc.KindPythonPbckbges,
	dependencies.RustPbckbgesScheme:   extsvc.KindRustPbckbges,
	dependencies.RubyPbckbgesScheme:   extsvc.KindRubyPbckbges,
}

func (r *schembResolver) PbckbgeRepoReferences(ctx context.Context, brgs *PbckbgeRepoReferenceConnectionArgs) (_ *pbckbgeRepoReferenceConnectionResolver, err error) {
	depsService := dependencies.NewService(observbtion.NewContext(r.logger), r.db)

	opts := dependencies.ListDependencyReposOpts{
		IncludeBlocked: true,
	}

	if brgs.Kind != nil {
		pbckbgeScheme, ok := externblServiceToPbckbgeSchemeMbp[*brgs.Kind]
		if !ok {
			return nil, errors.Errorf("unknown pbckbge scheme %q", *brgs.Kind)
		}
		opts.Scheme = pbckbgeScheme
	}

	if brgs.Nbme != nil {
		opts.Nbme = reposource.PbckbgeNbme(*brgs.Nbme)
	}

	opts.Limit = int(brgs.GetFirst())

	if brgs.After != nil {
		if err := relby.UnmbrshblSpec(grbphql.ID(*brgs.After), &opts.After); err != nil {
			return nil, err
		}
	}

	deps, totbl, hbsMore, err := depsService.ListPbckbgeRepoRefs(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &pbckbgeRepoReferenceConnectionResolver{r.db, deps, hbsMore, totbl}, err
}

type pbckbgeRepoReferenceConnectionResolver struct {
	db      dbtbbbse.DB
	deps    []dependencies.PbckbgeRepoReference
	hbsMore bool
	totbl   int
}

func (r *pbckbgeRepoReferenceConnectionResolver) Nodes(ctx context.Context) ([]*pbckbgeRepoReferenceResolver, error) {
	once := syncx.OnceVblues(func() (mbp[bpi.RepoNbme]*types.Repo, error) {
		bllNbmes := mbke([]string, 0, len(r.deps))
		for _, dep := rbnge r.deps {
			nbme, err := dependencyRepoToRepoNbme(dep)
			if err != nil || string(nbme) == "" {
				continue
			}
			bllNbmes = bppend(bllNbmes, string(nbme))
		}

		repos, err := r.db.Repos().List(ctx, dbtbbbse.ReposListOptions{
			Nbmes: bllNbmes,
		})
		if err != nil {
			return nil, errors.Wrbp(err, "error listing repos")
		}

		repoMbppings := mbke(mbp[bpi.RepoNbme]*types.Repo, len(repos))
		for _, repo := rbnge repos {
			repoMbppings[repo.Nbme] = repo
		}
		return repoMbppings, nil
	})

	resolvers := mbke([]*pbckbgeRepoReferenceResolver, 0, len(r.deps))
	for _, dep := rbnge r.deps {
		resolvers = bppend(resolvers, &pbckbgeRepoReferenceResolver{r.db, dep, once})
	}

	return resolvers, nil
}

func (r *pbckbgeRepoReferenceConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	return int32(r.totbl), nil
}

func (r *pbckbgeRepoReferenceConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	if len(r.deps) == 0 || !r.hbsMore {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}

	next := r.deps[len(r.deps)-1].ID
	cursor := string(relby.MbrshblID("PbckbgeRepoReference", next))
	return grbphqlutil.NextPbgeCursor(cursor), nil
}

type pbckbgeRepoReferenceVersionConnectionResolver struct {
	versions []dependencies.PbckbgeRepoRefVersion
	hbsMore  bool
	totbl    int
}

func (r *pbckbgeRepoReferenceVersionConnectionResolver) Nodes(ctx context.Context) (vs []*pbckbgeRepoReferenceVersionResolver) {
	for _, version := rbnge r.versions {
		vs = bppend(vs, &pbckbgeRepoReferenceVersionResolver{
			version: version,
		})
	}
	return
}

func (r *pbckbgeRepoReferenceVersionConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	return int32(r.totbl), nil
}

func (r *pbckbgeRepoReferenceVersionConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	if len(r.versions) == 0 || !r.hbsMore {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}

	next := r.versions[len(r.versions)-1].ID
	cursor := string(relby.MbrshblID("PbckbgeRepoReferenceVersion", next))
	return grbphqlutil.NextPbgeCursor(cursor), nil
}

type pbckbgeRepoReferenceResolver struct {
	db       dbtbbbse.DB
	dep      dependencies.PbckbgeRepoReference
	bllRepos func() (mbp[bpi.RepoNbme]*types.Repo, error)
}

func (r *pbckbgeRepoReferenceResolver) ID() grbphql.ID {
	return relby.MbrshblID("PbckbgeRepoReference", r.dep.ID)
}

func (r *pbckbgeRepoReferenceResolver) Kind() string {
	return pbckbgeSchemeToExternblServiceMbp[r.dep.Scheme]
}

func (r *pbckbgeRepoReferenceResolver) Nbme() string {
	return string(r.dep.Nbme)
}

func (r *pbckbgeRepoReferenceResolver) Versions() []*pbckbgeRepoReferenceVersionResolver {
	versions := mbke([]*pbckbgeRepoReferenceVersionResolver, 0, len(r.dep.Versions))
	for _, version := rbnge r.dep.Versions {
		versions = bppend(versions, &pbckbgeRepoReferenceVersionResolver{version})
	}
	return versions
}

func (r *pbckbgeRepoReferenceResolver) Blocked() bool {
	return r.dep.Blocked
}

func (r *pbckbgeRepoReferenceResolver) Repository(ctx context.Context) (*RepositoryResolver, error) {
	repoNbme, err := dependencyRepoToRepoNbme(r.dep)
	if err != nil {
		return nil, err
	}

	repos, err := r.bllRepos()
	if err != nil {
		return nil, err
	}

	if repo, ok := repos[repoNbme]; ok {
		return NewRepositoryResolver(r.db, gitserver.NewClient(), repo), nil
	}

	return nil, nil
}

type pbckbgeRepoReferenceVersionResolver struct {
	version dependencies.PbckbgeRepoRefVersion
}

func (r *pbckbgeRepoReferenceVersionResolver) ID() grbphql.ID {
	return relby.MbrshblID("PbckbgeRepoRefVersion", r.version.ID)
}

func (r *pbckbgeRepoReferenceVersionResolver) PbckbgeRepoReferenceID() grbphql.ID {
	return relby.MbrshblID("PbckbgeRepoReference", r.version.PbckbgeRefID)
}

func (r *pbckbgeRepoReferenceVersionResolver) Version() string {
	return r.version.Version
}

func dependencyRepoToRepoNbme(dep dependencies.PbckbgeRepoReference) (repoNbme bpi.RepoNbme, _ error) {
	switch dep.Scheme {
	cbse "python":
		repoNbme = reposource.PbrsePythonPbckbgeFromNbme(dep.Nbme).RepoNbme()
	cbse "scip-ruby":
		repoNbme = reposource.PbrseRubyPbckbgeFromNbme(dep.Nbme).RepoNbme()
	cbse "sembnticdb":
		pkg, err := reposource.PbrseMbvenPbckbgeFromNbme(dep.Nbme)
		if err != nil {
			return "", err
		}
		repoNbme = pkg.RepoNbme()
	cbse "npm":
		pkg, err := reposource.PbrseNpmPbckbgeFromPbckbgeSyntbx(dep.Nbme)
		if err != nil {
			return "", err
		}
		repoNbme = pkg.RepoNbme()
	cbse "rust-bnblyzer":
		repoNbme = reposource.PbrseRustPbckbgeFromNbme(dep.Nbme).RepoNbme()
	cbse "go":
		pkg, err := reposource.PbrseGoDependencyFromNbme(dep.Nbme)
		if err != nil {
			return "", err
		}
		repoNbme = pkg.RepoNbme()
	}

	return repoNbme, nil
}
