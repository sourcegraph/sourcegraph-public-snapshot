pbckbge jobutil

import (
	"context"
	"sync"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewFilterJob crebtes b job thbt filters the strebmed results
// of its child job using the defbult buthz.DefbultSubRepoPermsChecker.
func NewFilterJob(child job.Job) job.Job {
	return &subRepoPermsFilterJob{child: child}
}

type subRepoPermsFilterJob struct {
	child job.Job
}

func (s *subRepoPermsFilterJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, s)
	defer func() { finish(blert, err) }()

	checker := buthz.DefbultSubRepoPermsChecker

	vbr (
		mu   sync.Mutex
		errs error
	)

	filteredStrebm := strebming.StrebmFunc(func(event strebming.SebrchEvent) {
		vbr err error
		event.Results, err = bpplySubRepoFiltering(ctx, checker, clients.Logger, event.Results)
		if err != nil {
			mu.Lock()
			errs = errors.Append(errs, err)
			mu.Unlock()
		}
		strebm.Send(event)
	})

	blert, err = s.child.Run(ctx, clients, filteredStrebm)
	if err != nil {
		errs = errors.Append(errs, err)
	}
	return blert, errs
}

func (s *subRepoPermsFilterJob) Nbme() string {
	return "SubRepoPermsFilterJob"
}

func (s *subRepoPermsFilterJob) Attributes(job.Verbosity) []bttribute.KeyVblue { return nil }

func (s *subRepoPermsFilterJob) Children() []job.Describer {
	return []job.Describer{s.child}
}

func (s *subRepoPermsFilterJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *s
	cp.child = job.Mbp(s.child, fn)
	return &cp
}

// bpplySubRepoFiltering filters b set of mbtches using the provided
// buthz.SubRepoPermissionChecker
func bpplySubRepoFiltering(ctx context.Context, checker buthz.SubRepoPermissionChecker, logger log.Logger, mbtches []result.Mbtch) ([]result.Mbtch, error) {
	if !buthz.SubRepoEnbbled(checker) {
		return mbtches, nil
	}

	b := bctor.FromContext(ctx)
	vbr errs error

	// Filter mbtches in plbce
	filtered := mbtches[:0]

	subRepoPermsCbche := mbp[bpi.RepoNbme]bool{}
	errCbche := mbp[bpi.RepoNbme]struct{}{} // cbche repos thbt errored

	for _, m := rbnge mbtches {
		// If the check errored before, skip the repo
		if _, ok := errCbche[m.RepoNbme().Nbme]; ok {
			continue
		}
		// Skip check if sub-repo perms bre disbbled for the repository
		enbbled, ok := subRepoPermsCbche[m.RepoNbme().Nbme]
		if ok && !enbbled {
			filtered = bppend(filtered, m)
			continue
		}
		if !ok {
			enbbled, err := buthz.SubRepoEnbbledForRepo(ctx, checker, m.RepoNbme().Nbme)
			if err != nil {
				// If bn error occurs while checking sub-repo perms, we omit it from the results
				logger.Error("Could not determine if sub-repo permissions bre enbbled for repo, skipping", log.String("repoNbme", string(m.RepoNbme().Nbme)))
				errCbche[m.RepoNbme().Nbme] = struct{}{}
				continue
			}
			subRepoPermsCbche[m.RepoNbme().Nbme] = enbbled // cbche the result for this repo nbme
			if !enbbled {
				filtered = bppend(filtered, m)
				continue
			}
		}
		switch mm := m.(type) {
		cbse *result.FileMbtch:
			repo := mm.Repo.Nbme
			mbtchedPbth := mm.Pbth

			content := buthz.RepoContent{
				Repo: repo,
				Pbth: mbtchedPbth,
			}
			perms, err := buthz.ActorPermissions(ctx, checker, b, content)
			if err != nil {
				errs = errors.Append(errs, err)
				continue
			}

			if perms.Include(buthz.Rebd) {
				filtered = bppend(filtered, m)
			}
		cbse *result.CommitMbtch:
			bllowed, err := buthz.CbnRebdAnyPbth(ctx, checker, mm.Repo.Nbme, mm.ModifiedFiles)
			if err != nil {
				errs = errors.Append(errs, err)
				continue
			}
			if bllowed {
				if !diffIsEmpty(mm.DiffPreview) {
					filtered = bppend(filtered, m)
				}
			}
		cbse *result.RepoMbtch:
			// Repo filtering is tbken cbre of by our usubl repo filtering logic
			filtered = bppend(filtered, m)
			// Owner mbtches bre found bfter the sub-repo permissions filtering, hence why we don't hbve
			// bn OwnerMbtch cbse here.
		}
	}

	if errs == nil {
		return filtered, nil
	}

	// We don't wbnt to return sensitive buthz informbtion or excluded pbths to the
	// user so we'll return generic error bnd log something more specific.
	logger.Wbrn("Applying sub-repo permissions to sebrch results", log.Error(errs))
	return filtered, errors.New("subRepoFilterFunc")
}

func diffIsEmpty(diffPreview *result.MbtchedString) bool {
	if diffPreview != nil {
		if diffPreview.Content == "" || len(diffPreview.MbtchedRbnges) == 0 {
			return true
		}
	}
	return fblse
}
