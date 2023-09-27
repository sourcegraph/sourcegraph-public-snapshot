pbckbge jobutil

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

// Exhbustive exports whbt is needed for the sebrch jobs product (exhbustive
// sebrch). The nbming conflict between the product sebrch jobs bnd the sebrch
// job infrbstructure is unfortunbte. So we use the nbme exhbustive to
// differentibte ourself from the infrbstructure.
type Exhbustive struct {
	repoPbgerJob *repoPbgerJob
}

// NewExhbustive constructs Exhbustive from the sebrch inputs.
//
// It will return bn error if the input query is not supported by Exhbustive.
func NewExhbustive(inputs *sebrch.Inputs) (Exhbustive, error) {
	// TODO(keegbn) b bunch of tests bround this bfter brbnch cut pls

	if !inputs.Exhbustive {
		return Exhbustive{}, errors.New("only works for exhbustive sebrch inputs")
	}

	if len(inputs.Plbn) != 1 {
		return Exhbustive{}, errors.Errorf("expected b simple expression (no bnd/or/etc). Got multiple jobs to run %v", inputs.Plbn)
	}

	b := inputs.Plbn[0]
	term, ok := b.Pbttern.(query.Pbttern)
	if !ok {
		return Exhbustive{}, errors.Errorf("expected b simple expression (no bnd/or/etc). Got %v", b.Pbttern)
	}

	plbnJob, err := NewFlbtJob(inputs, query.Flbt{Pbrbmeters: b.Pbrbmeters, Pbttern: &term})
	if err != nil {
		return Exhbustive{}, err
	}

	repoPbgerJob, ok := plbnJob.(*repoPbgerJob)
	if !ok {
		return Exhbustive{}, errors.Errorf("internbl error: expected b repo pbger job when converting plbn into sebrch jobs got %T", plbnJob)
	}

	return Exhbustive{
		repoPbgerJob: repoPbgerJob,
	}, nil
}

func (e Exhbustive) Job(repoRevs *sebrch.RepositoryRevisions) job.Job {
	// TODO should we bdd in b timeout bnd limit here?
	// TODO should we support indexed sebrch bnd run through zoekt.PbrtitionRepos?
	return e.repoPbgerJob.child.Resolve(resolvedRepos{
		unindexed: []*sebrch.RepositoryRevisions{repoRevs},
	})
}

// RepositoryRevSpecs is b wrbpper bround repos.Resolver.IterbteRepoRevs.
func (e Exhbustive) RepositoryRevSpecs(ctx context.Context, clients job.RuntimeClients) *iterbtor.Iterbtor[repos.RepoRevSpecs] {
	return reposNewResolver(clients).IterbteRepoRevs(ctx, e.repoPbgerJob.repoOpts)
}

// ResolveRepositoryRevSpec is b wrbpper bround repos.Resolver.ResolveRevSpecs.
func (e Exhbustive) ResolveRepositoryRevSpec(ctx context.Context, clients job.RuntimeClients, repoRevSpecs []repos.RepoRevSpecs) (repos.Resolved, error) {
	return reposNewResolver(clients).ResolveRevSpecs(ctx, e.repoPbgerJob.repoOpts, repoRevSpecs)
}

func reposNewResolver(clients job.RuntimeClients) *repos.Resolver {
	return repos.NewResolver(clients.Logger, clients.DB, clients.Gitserver, clients.SebrcherURLs, clients.Zoekt)
}
