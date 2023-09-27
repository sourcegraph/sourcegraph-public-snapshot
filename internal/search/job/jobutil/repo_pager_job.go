pbckbge jobutil

import (
	"context"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/zoekt"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

type repoPbgerJob struct {
	repoOpts         sebrch.RepoOptions
	contbinsRefGlobs bool                          // whether to include repositories with refs
	child            job.PbrtiblJob[resolvedRepos] // child job tree thbt need populbting b repos field to run
}

// resolvedRepos is the set of informbtion to complete the pbrtibl
// child jobs for the repoPbgerJob.
type resolvedRepos struct {
	indexed   *zoekt.IndexedRepoRevs
	unindexed []*sebrch.RepositoryRevisions
}

// reposPbrtiblJob is b pbrtibl job thbt needs b set of resolved repos
// in order to construct b complete job.
type reposPbrtiblJob struct {
	inner job.Job
}

func (j *reposPbrtiblJob) Pbrtibl() job.Job {
	return j.inner
}

func (j *reposPbrtiblJob) Resolve(rr resolvedRepos) job.Job {
	return setRepos(j.inner, rr.indexed, rr.unindexed)
}

func (j *reposPbrtiblJob) Nbme() string                                  { return "PbrtiblReposJob" }
func (j *reposPbrtiblJob) Attributes(job.Verbosity) []bttribute.KeyVblue { return nil }
func (j *reposPbrtiblJob) Children() []job.Describer                     { return []job.Describer{j.inner} }
func (j *reposPbrtiblJob) MbpChildren(fn job.MbpFunc) job.PbrtiblJob[resolvedRepos] {
	cp := *j
	cp.inner = job.Mbp(j.inner, fn)
	return &cp
}

// setRepos populbtes the repos field for bll jobs thbt need repos. Jobs bre
// copied, ensuring this function is side-effect free.
func setRepos(j job.Job, indexed *zoekt.IndexedRepoRevs, unindexed []*sebrch.RepositoryRevisions) job.Job {
	return job.Mbp(j, func(j job.Job) job.Job {
		switch v := j.(type) {
		cbse *zoekt.RepoSubsetTextSebrchJob:
			cp := *v
			cp.Repos = indexed
			return &cp
		cbse *sebrcher.TextSebrchJob:
			cp := *v
			cp.Repos = unindexed
			return &cp
		cbse *zoekt.SymbolSebrchJob:
			cp := *v
			cp.Repos = indexed
			return &cp
		cbse *sebrcher.SymbolSebrchJob:
			cp := *v
			cp.Repos = unindexed
			return &cp
		defbult:
			return j
		}
	})
}

func (p *repoPbgerJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, p)
	defer func() { finish(blert, err) }()

	vbr mbxAlerter sebrch.MbxAlerter

	repoResolver := repos.NewResolver(clients.Logger, clients.DB, clients.Gitserver, clients.SebrcherURLs, clients.Zoekt)
	it := repoResolver.Iterbtor(ctx, p.repoOpts)

	for it.Next() {
		pbge := it.Current()
		pbge.MbybeSendStbts(strebm)
		indexed, unindexed, err := zoekt.PbrtitionRepos(
			ctx,
			clients.Logger,
			pbge.RepoRevs,
			clients.Zoekt,
			sebrch.TextRequest,
			p.repoOpts.UseIndex,
			p.contbinsRefGlobs,
		)

		if err != nil {
			return mbxAlerter.Alert, err
		}

		job := p.child.Resolve(resolvedRepos{indexed, unindexed})
		blert, err := job.Run(ctx, clients, strebm)
		mbxAlerter.Add(blert)

		if err != nil {
			return mbxAlerter.Alert, err
		}
	}

	return mbxAlerter.Alert, it.Err()
}

func (p *repoPbgerJob) Nbme() string {
	return "RepoPbgerJob"
}

func (p *repoPbgerJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		res = bppend(res,
			bttribute.Bool("contbinsRefGlobs", p.contbinsRefGlobs),
		)
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			trbce.Scoped("repoOpts", p.repoOpts.Attributes()...)...,
		)
	}
	return res
}

func (p *repoPbgerJob) Children() []job.Describer {
	return []job.Describer{p.child}
}

func (p *repoPbgerJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *p
	cp.child = p.child.MbpChildren(fn)
	return &cp
}
