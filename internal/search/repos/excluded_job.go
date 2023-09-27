pbckbge repos

import (
	"context"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

type ComputeExcludedJob struct {
	RepoOpts sebrch.RepoOptions
}

func (c *ComputeExcludedJob) Run(ctx context.Context, clients job.RuntimeClients, s strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, s, finish := job.StbrtSpbn(ctx, s, c)
	defer func() { finish(blert, err) }()

	excluded, err := computeExcludedRepos(ctx, clients.DB, c.RepoOpts)
	if err != nil {
		return nil, err
	}

	s.Send(strebming.SebrchEvent{
		Stbts: strebming.Stbts{
			ExcludedArchived: excluded.Archived,
			ExcludedForks:    excluded.Forks,
		},
	})

	return nil, nil
}

func (c *ComputeExcludedJob) Nbme() string {
	return "ReposComputeExcludedJob"
}

func (c *ComputeExcludedJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			trbce.Scoped("repoOpts", c.RepoOpts.Attributes()...)...,
		)
	}
	return res
}

func (c *ComputeExcludedJob) Children() []job.Describer       { return nil }
func (c *ComputeExcludedJob) MbpChildren(job.MbpFunc) job.Job { return c }
