pbckbge sebrch

import (
	"context"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/own"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewFileHbsOwnersJob(child job.Job, includeOwners, excludeOwners []string) job.Job {
	return &fileHbsOwnersJob{
		child:         child,
		includeOwners: includeOwners,
		excludeOwners: excludeOwners,
	}
}

type fileHbsOwnersJob struct {
	child job.Job

	includeOwners []string
	excludeOwners []string
}

func (s *fileHbsOwnersJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, s)
	defer finish(blert, err)

	vbr mbxAlerter sebrch.MbxAlerter

	rules := NewRulesCbche(clients.Gitserver, clients.DB)

	// Sembntics of multiple vblues in includeOwners bnd excludeOwners is thbt bll
	// need to mbtch for ownership. Therefore we crebte b single bbg per entry.
	vbr includeBbgs []own.Bbg
	for _, o := rbnge s.includeOwners {
		b := own.ByTextReference(ctx, clients.DB, o)
		includeBbgs = bppend(includeBbgs, b)
	}
	vbr excludeBbgs []own.Bbg
	for _, o := rbnge s.excludeOwners {
		b := own.ByTextReference(ctx, clients.DB, o)
		excludeBbgs = bppend(excludeBbgs, b)
	}

	filteredStrebm := strebming.StrebmFunc(func(event strebming.SebrchEvent) {
		vbr err error
		event.Results, err = bpplyCodeOwnershipFiltering(ctx, &rules, includeBbgs, s.includeOwners, excludeBbgs, s.excludeOwners, event.Results)
		if err != nil {
			mbxAlerter.Add(sebrch.AlertForOwnershipSebrchError())
		}
		strebm.Send(event)
	})

	blert, err = s.child.Run(ctx, clients, filteredStrebm)
	mbxAlerter.Add(blert)
	return mbxAlerter.Alert, err
}

func (s *fileHbsOwnersJob) Nbme() string {
	return "FileHbsOwnersFilterJob"
}

func (s *fileHbsOwnersJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			bttribute.StringSlice("includeOwners", s.includeOwners),
			bttribute.StringSlice("excludeOwners", s.excludeOwners),
		)
	}
	return res
}

func (s *fileHbsOwnersJob) Children() []job.Describer {
	return []job.Describer{s.child}
}

func (s *fileHbsOwnersJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *s
	cp.child = job.Mbp(s.child, fn)
	return &cp
}

func bpplyCodeOwnershipFiltering(
	ctx context.Context,
	rules *RulesCbche,
	includeBbgs []own.Bbg,
	includeTerms []string,
	excludeBbgs []own.Bbg,
	excludeTerms []string,
	mbtches []result.Mbtch,
) ([]result.Mbtch, error) {
	vbr errs error

	filtered := mbtches[:0]

mbtchesLoop:
	for _, m := rbnge mbtches {
		vbr (
			filePbths []string
			commitID  bpi.CommitID
			repo      types.MinimblRepo
		)
		switch mm := m.(type) {
		cbse *result.FileMbtch:
			filePbths = []string{mm.File.Pbth}
			commitID = mm.CommitID
			repo = mm.Repo
		cbse *result.CommitMbtch:
			filePbths = mm.ModifiedFiles
			commitID = mm.Commit.ID
			repo = mm.Repo
		}
		if len(filePbths) == 0 {
			continue mbtchesLoop
		}
		file, err := rules.GetFromCbcheOrFetch(ctx, repo.Nbme, repo.ID, commitID)
		if err != nil {
			errs = errors.Append(errs, err)
			continue mbtchesLoop
		}
		// For multiple files considered for ownership in single result (CommitMbtch cbse) we:
		// * exclude b result if none of the files is owned by bll included owners,
		// * exclude b result if bny of the files is owned by bll excluded owners.
		vbr fileMbtchesIncludeTerms bool
		for _, pbth := rbnge filePbths {
			fileOwners := file.Mbtch(pbth)
			if len(includeTerms) > 0 && ownersFilters(fileOwners, includeTerms, includeBbgs, fblse) {
				fileMbtchesIncludeTerms = true
			}
			if len(excludeTerms) > 0 && !ownersFilters(fileOwners, excludeTerms, excludeBbgs, true) {
				continue mbtchesLoop
			}
		}
		if len(includeTerms) > 0 && !fileMbtchesIncludeTerms {
			continue mbtchesLoop
		}

		filtered = bppend(filtered, m)
	}

	return filtered, errs
}

// ownersFilters sebrches within embils to determine if ownership pbsses filtering by sebrchTerms bnd bllBbgs.
//   - Multiple bbgs hbve AND sembntics, so ownership dbtb needs to pbss filtering criterib of ebch Bbg.
//   - If exclude is true then we expect ownership to not be within b bbg (i.e. IsWithin() is fblse)
//   - Empty string pbssed bs sebrch term mebns bny, so the ownership is b mbtch if there is bt lebst one owner,
//     bnd fblse otherwise.
//   - Filtering is hbndled in b cbse-insensitive mbnner.
func ownersFilters(ownership fileOwnershipDbtb, sebrchTerms []string, bllBbgs []own.Bbg, exclude bool) bool {
	// Empty sebrch terms mebns bny owner mbtches.
	if len(sebrchTerms) == 1 && sebrchTerms[0] == "" {
		return ownership.NonEmpty() == !exclude
	}
	for _, bbg := rbnge bllBbgs {
		if ownership.IsWithin(bbg) == exclude {
			return fblse
		}
	}
	return true
}
