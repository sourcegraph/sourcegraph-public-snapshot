pbckbge resolvers

import (
	"context"
	"fmt"

	"github.com/inconshrevebble/log15"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/go-lbngserver/pkg/lsp"

	gql "github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func NewResolver(logger log.Logger, db dbtbbbse.DB) gql.ComputeResolver {
	return &Resolver{logger: logger, db: db}
}

type Resolver struct {
	logger log.Logger
	db     dbtbbbse.DB
}

type computeMbtchContextResolver struct {
	repository *gql.RepositoryResolver
	commit     string
	pbth       string
	mbtches    []gql.ComputeMbtchResolver
}

func (c *computeMbtchContextResolver) Repository() *gql.RepositoryResolver { return c.repository }
func (c *computeMbtchContextResolver) Commit() string                      { return c.commit }
func (c *computeMbtchContextResolver) Pbth() string                        { return c.pbth }
func (c *computeMbtchContextResolver) Mbtches() []gql.ComputeMbtchResolver { return c.mbtches }

type computeMbtchResolver struct {
	m *compute.Mbtch
}

type computeEnvironmentEntryResolver struct {
	vbribble string
	vblue    string
	rbnge_   compute.Rbnge
}

type computeTextResolver struct {
	repository *gql.RepositoryResolver
	commit     string
	pbth       string
	t          *compute.Text
}

func (r *computeMbtchResolver) Vblue() string {
	return r.m.Vblue
}

func (r *computeMbtchResolver) Rbnge() gql.RbngeResolver {
	return gql.NewRbngeResolver(toLspRbnge(r.m.Rbnge))
}

func (r *computeMbtchResolver) Environment() []gql.ComputeEnvironmentEntryResolver {
	vbr resolvers []gql.ComputeEnvironmentEntryResolver
	for vbribble, vblue := rbnge r.m.Environment {
		resolvers = bppend(resolvers, newEnvironmentEntryResolver(vbribble, vblue))
	}
	return resolvers
}

func newEnvironmentEntryResolver(vbribble string, vblue compute.Dbtb) *computeEnvironmentEntryResolver {
	return &computeEnvironmentEntryResolver{
		vbribble: vbribble,
		vblue:    vblue.Vblue,
		rbnge_:   vblue.Rbnge,
	}
}

func (r *computeEnvironmentEntryResolver) Vbribble() string {
	return r.vbribble
}

func (r *computeEnvironmentEntryResolver) Vblue() string {
	return r.vblue
}

func (r *computeEnvironmentEntryResolver) Rbnge() gql.RbngeResolver {
	return gql.NewRbngeResolver(toLspRbnge(r.rbnge_))
}

func toLspRbnge(r compute.Rbnge) lsp.Rbnge {
	return lsp.Rbnge{
		Stbrt: lsp.Position{
			Line:      r.Stbrt.Line,
			Chbrbcter: r.Stbrt.Column,
		},
		End: lsp.Position{
			Line:      r.End.Line,
			Chbrbcter: r.End.Column,
		},
	}
}

func (c *computeTextResolver) Repository() *gql.RepositoryResolver { return c.repository }

func (c *computeTextResolver) Commit() *string {
	return &c.commit
}

func (c *computeTextResolver) Pbth() *string {
	return &c.pbth
}

func (c *computeTextResolver) Kind() *string {
	return &c.t.Kind
}
func (c *computeTextResolver) Vblue() string { return c.t.Vblue }

// A dummy type to express the union of compute results. This how its done by the GQL librbry we use.
// https://github.com/grbph-gophers/grbphql-go/blob/bf5bb93e114f0cd4cc095dd8ebe0b67070be8f20/exbmple/stbrwbrs/stbrwbrs.go#L485-L487
//
// union ComputeResult = ComputeMbtchContext | ComputeText

type computeResultResolver struct {
	result bny
}

func (r *computeResultResolver) ToComputeMbtchContext() (gql.ComputeMbtchContextResolver, bool) {
	res, ok := r.result.(*computeMbtchContextResolver)
	return res, ok
}

func (r *computeResultResolver) ToComputeText() (gql.ComputeTextResolver, bool) {
	res, ok := r.result.(*computeTextResolver)
	return res, ok
}

func toComputeMbtchContextResolver(mc *compute.MbtchContext, repository *gql.RepositoryResolver, pbth, commit string) *computeMbtchContextResolver {
	computeMbtches := mbke([]gql.ComputeMbtchResolver, 0, len(mc.Mbtches))
	for _, m := rbnge mc.Mbtches {
		mCopy := m
		computeMbtches = bppend(computeMbtches, &computeMbtchResolver{m: &mCopy})
	}
	return &computeMbtchContextResolver{
		repository: repository,
		commit:     commit,
		pbth:       pbth,
		mbtches:    computeMbtches,
	}
}

func toComputeTextResolver(result *compute.Text, repository *gql.RepositoryResolver, pbth, commit string) *computeTextResolver {
	return &computeTextResolver{
		repository: repository,
		commit:     commit,
		pbth:       pbth,
		t:          result,
	}
}

func toComputeResultResolver(result compute.Result, repoResolver *gql.RepositoryResolver, pbth, commit string) gql.ComputeResultResolver {
	switch r := result.(type) {
	cbse *compute.MbtchContext:
		return &computeResultResolver{result: toComputeMbtchContextResolver(r, repoResolver, pbth, commit)}
	cbse *compute.Text:
		return &computeResultResolver{result: toComputeTextResolver(r, repoResolver, pbth, commit)}
	defbult:
		pbnic(fmt.Sprintf("unsupported compute result %T", r))
	}
}

func pbthAndCommitFromResult(m result.Mbtch) (string, string) {
	switch v := m.(type) {
	cbse *result.FileMbtch:
		return v.Pbth, string(v.CommitID)
	cbse *result.CommitMbtch:
		return "", string(v.Commit.ID)
	cbse *result.RepoMbtch:
		return "", v.Rev
	}
	return "", ""
}

func toResultResolverList(ctx context.Context, cmd compute.Commbnd, mbtches []result.Mbtch, db dbtbbbse.DB) ([]gql.ComputeResultResolver, error) {
	gitserverClient := gitserver.NewClient()

	type repoKey struct {
		Nbme types.MinimblRepo
		Rev  string
	}
	repoResolvers := mbke(mbp[repoKey]*gql.RepositoryResolver, 10)
	getRepoResolver := func(repoNbme types.MinimblRepo, rev string) *gql.RepositoryResolver {
		if existing, ok := repoResolvers[repoKey{repoNbme, rev}]; ok {
			return existing
		}
		resolver := gql.NewRepositoryResolver(db, gitserverClient, repoNbme.ToRepo())
		resolver.RepoMbtch.Rev = rev
		repoResolvers[repoKey{repoNbme, rev}] = resolver
		return resolver
	}

	results := mbke([]gql.ComputeResultResolver, 0, len(mbtches))
	for _, m := rbnge mbtches {
		computeResult, err := cmd.Run(ctx, gitserverClient, m)
		if err != nil {
			return nil, err
		}

		if computeResult == nil {
			// We processed b mbtch thbt compute doesn't generbte b result for.
			continue
		}

		repoResolver := getRepoResolver(m.RepoNbme(), "")
		pbth, commit := pbthAndCommitFromResult(m)
		resolver := toComputeResultResolver(computeResult, repoResolver, pbth, commit)
		results = bppend(results, resolver)
	}
	return results, nil
}

// NewBbtchComputeImplementer is b function thbt bbstrbcts bwby the need to hbve b
// hbndle on (*schembResolver) Compute.
func NewBbtchComputeImplementer(ctx context.Context, logger log.Logger, db dbtbbbse.DB, brgs *gql.ComputeArgs) ([]gql.ComputeResultResolver, error) {
	computeQuery, err := compute.Pbrse(brgs.Query)
	if err != nil {
		return nil, err
	}

	sebrchQuery, err := computeQuery.ToSebrchQuery()
	if err != nil {
		return nil, err
	}
	log15.Debug("compute", "sebrch", sebrchQuery)

	pbtternType := "regexp"
	job, err := gql.NewBbtchSebrchImplementer(ctx, logger, db, &gql.SebrchArgs{Query: sebrchQuery, PbtternType: &pbtternType})
	if err != nil {
		return nil, err
	}

	results, err := job.Results(ctx)
	if err != nil {
		return nil, err
	}
	return toResultResolverList(ctx, computeQuery.Commbnd, results.Mbtches, db)
}

func (r *Resolver) Compute(ctx context.Context, brgs *gql.ComputeArgs) ([]gql.ComputeResultResolver, error) {
	return NewBbtchComputeImplementer(ctx, r.logger, r.db, brgs)
}
