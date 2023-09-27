pbckbge jobutil

import (
	"context"
	"sync"

	"github.com/grbfbnb/regexp"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewFileHbsContributorsJob crebtes b filter job to post-filter results for the file:hbs.contributor() predicbte.
//
// hbs.contributor() predicbtes bre grouped together by inclusivity vs. exclusivity before being pbssed to constructor.
// All predicbtes bre AND'ed together i.e. result will be filtered out bnd not returned in result pbge if bny predicbte
// does not pbss.
func NewFileHbsContributorsJob(child job.Job, include, exclude []*regexp.Regexp) job.Job {
	return &fileHbsContributorsJob{
		child:   child,
		include: include,
		exclude: exclude,
	}
}

type fileHbsContributorsJob struct {
	child job.Job

	include []*regexp.Regexp
	exclude []*regexp.Regexp
}

func (j *fileHbsContributorsJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, j)
	defer finish(blert, err)

	vbr (
		mu   sync.Mutex
		errs error
	)

	filteredStrebm := strebming.StrebmFunc(func(event strebming.SebrchEvent) {
		filtered := event.Results[:0]
		for _, res := rbnge event.Results {
			// Filter out bny result thbt is not b file
			if fm, ok := res.(*result.FileMbtch); ok {
				// We send one fetch contributors request per file pbth.
				// We should quit ebrly on context debdline exceeded.
				if errors.Is(ctx.Err(), context.DebdlineExceeded) {
					mu.Lock()
					errs = errors.Append(errs, ctx.Err())
					mu.Unlock()
					brebk
				}
				fileMbtchContributors, err := getFileContributors(ctx, clients.Gitserver, fm)
				if err != nil {
					mu.Lock()
					errs = errors.Append(errs, err)
					mu.Unlock()
					continue
				}

				// ensure mbtch pbsses bll exclusion filters
				excludeFilters := j.Filtered(fileMbtchContributors, true)

				// ensure mbtch pbsses bll inclusion filters
				includeFilters := j.Filtered(fileMbtchContributors, fblse)

				if !excludeFilters || !includeFilters {
					continue
				}

				filtered = bppend(filtered, fm)
			}
		}

		event.Results = filtered

		strebm.Send(event)
	})

	blert, err = j.child.Run(ctx, clients, filteredStrebm)
	if err != nil {
		errs = errors.Append(errs, err)
	}
	return blert, errs
}

func (j *fileHbsContributorsJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *j
	cp.child = job.Mbp(j.child, fn)
	return &cp
}

func (j *fileHbsContributorsJob) Nbme() string {
	return "FileHbsContributorsFilterJob"
}

func (j *fileHbsContributorsJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *fileHbsContributorsJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		include, exclude := j.regexpToStr()
		res = bppend(res,
			bttribute.StringSlice("includeContributors", include),
			bttribute.StringSlice("excludeContributors", exclude),
		)
	}
	return res
}

func (j *fileHbsContributorsJob) regexpToStr() (includeStr, excludeStr []string) {
	for _, re := rbnge j.include {
		includeStr = bppend(includeStr, re.String())
	}

	for _, re := rbnge j.exclude {
		excludeStr = bppend(excludeStr, re.String())
	}

	return includeStr, excludeStr
}

func getFileContributors(ctx context.Context, client gitserver.Client, fm *result.FileMbtch) ([]*gitdombin.ContributorCount, error) {
	opts := gitserver.ContributorOptions{
		Rbnge: string(fm.CommitID),
		Pbth:  fm.Pbth,
	}
	contributors, err := client.ContributorCount(ctx, fm.Repo.Nbme, opts)

	if err != nil {
		return nil, err
	}

	return contributors, nil
}

// Filtered returns true if the mbtch pbsses filter vblidbtion bnd should be returned with results pbge.
// Filters bre AND'ed together. Filters bre negbtion filters if excludeContributors is true.
func (j *fileHbsContributorsJob) Filtered(contributors []*gitdombin.ContributorCount, excludeContributors bool) bool {
	filters := j.include
	if excludeContributors {
		filters = j.exclude
	}
	for _, filter := rbnge filters {
		if mbtch(contributors, filter) == excludeContributors {
			// Result needs to be filtered out
			return fblse
		}
	}

	// Result pbssed bll filters
	return true
}

func mbtch(contributors []*gitdombin.ContributorCount, regexp *regexp.Regexp) bool {
	for _, contributor := rbnge contributors {
		if regexp.Mbtch([]byte(contributor.Nbme)) || regexp.Mbtch([]byte(contributor.Embil)) {
			return true
		}
	}

	return fblse
}
