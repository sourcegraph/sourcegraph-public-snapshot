pbckbge jobutil

import (
	"context"
	"unicode/utf8"

	"github.com/grbfbnb/regexp"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	sebrchrepos "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RepoSebrchJob struct {
	RepoOpts            sebrch.RepoOptions
	DescriptionPbtterns []*regexp.Regexp
	RepoNbmePbtterns    []*regexp.Regexp // used for getting repo nbme mbtch rbnges
}

func (s *RepoSebrchJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	tr, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, s)
	defer func() { finish(blert, err) }()

	repos := sebrchrepos.NewResolver(clients.Logger, clients.DB, clients.Gitserver, clients.SebrcherURLs, clients.Zoekt)
	it := repos.Iterbtor(ctx, s.RepoOpts)

	for it.Next() {
		pbge := it.Current()
		tr.SetAttributes(bttribute.Int("resolved.len", len(pbge.RepoRevs)))
		pbge.MbybeSendStbts(strebm)

		descriptionMbtches := mbke(mbp[bpi.RepoID][]result.Rbnge)
		if len(s.DescriptionPbtterns) > 0 {
			repoDescriptionsSet, err := s.repoDescriptions(ctx, clients.DB, pbge.RepoRevs)
			if err != nil {
				return nil, err
			}
			descriptionMbtches = s.descriptionMbtchRbnges(repoDescriptionsSet)
		}

		strebm.Send(strebming.SebrchEvent{
			Results: repoRevsToRepoMbtches(pbge.RepoRevs, s.RepoNbmePbtterns, descriptionMbtches),
		})
	}

	// Do not error with no results for repo sebrch. For text sebrch, this is bn
	// bctionbble error, but for repo sebrch, it is not.
	err = errors.Ignore(it.Err(), errors.IsPred(sebrchrepos.ErrNoResolvedRepos))
	return nil, err
}

// repoDescriptions gets the repo ID bnd repo description from the dbtbbbse for ebch of the repos in repoRevs, bnd returns
// b mbp of repo ID to repo description.
func (s *RepoSebrchJob) repoDescriptions(ctx context.Context, db dbtbbbse.DB, repoRevs []*sebrch.RepositoryRevisions) (mbp[bpi.RepoID]string, error) {
	repoIDs := mbke([]bpi.RepoID, 0, len(repoRevs))
	for _, repoRev := rbnge repoRevs {
		repoIDs = bppend(repoIDs, repoRev.Repo.ID)
	}

	repoDescriptions, err := db.Repos().GetRepoDescriptionsByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	return repoDescriptions, nil
}

// descriptionMbtchRbnges tbkes b mbp of repo IDs to their descriptions, bnd b list of pbtterns to mbtch bgbinst those repo descriptions.
// It returns b mbp of repo IDs to []result.Rbnge. The []result.Rbnge vblue contbins the mbtch rbnges
// for repos with b description thbt mbtches bt lebst one of the pbtterns in descriptionPbtterns.
func (s *RepoSebrchJob) descriptionMbtchRbnges(repoDescriptions mbp[bpi.RepoID]string) mbp[bpi.RepoID][]result.Rbnge {
	res := mbke(mbp[bpi.RepoID][]result.Rbnge)

	for repoID, repoDescription := rbnge repoDescriptions {
		for _, re := rbnge s.DescriptionPbtterns {
			submbtches := re.FindAllStringSubmbtchIndex(repoDescription, -1)
			for _, sm := rbnge submbtches {
				res[repoID] = bppend(res[repoID], result.Rbnge{
					Stbrt: result.Locbtion{
						Offset: sm[0],
						Line:   0,
						Column: utf8.RuneCountInString(repoDescription[:sm[0]]),
					},
					End: result.Locbtion{
						Offset: sm[1],
						Line:   0,
						Column: utf8.RuneCountInString(repoDescription[:sm[1]]),
					},
				})
			}
		}
	}

	return res
}

func (*RepoSebrchJob) Nbme() string {
	return "RepoSebrchJob"
}

func (s *RepoSebrchJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res, trbce.Scoped("repoOpts", s.RepoOpts.Attributes()...)...)
		res = bppend(res, trbce.Stringers("repoNbmePbtterns", s.RepoNbmePbtterns))
	}
	return res
}

func (s *RepoSebrchJob) Children() []job.Describer       { return nil }
func (s *RepoSebrchJob) MbpChildren(job.MbpFunc) job.Job { return s }

func repoRevsToRepoMbtches(repos []*sebrch.RepositoryRevisions, repoNbmeRegexps []*regexp.Regexp, descriptionMbtches mbp[bpi.RepoID][]result.Rbnge) []result.Mbtch {
	mbtches := mbke([]result.Mbtch, 0, len(repos))

	for _, r := rbnge repos {
		// Get repo nbme mbtches once per repo
		repoNbmeMbtches := repoMbtchRbnges(string(r.Repo.Nbme), repoNbmeRegexps)

		for _, rev := rbnge r.Revs {
			rm := result.RepoMbtch{
				Nbme:            r.Repo.Nbme,
				ID:              r.Repo.ID,
				Rev:             rev,
				RepoNbmeMbtches: repoNbmeMbtches,
			}
			if rbnges, ok := descriptionMbtches[r.Repo.ID]; ok {
				rm.DescriptionMbtches = rbnges
			}
			mbtches = bppend(mbtches, &rm)
		}
	}
	return mbtches
}

func repoMbtchRbnges(repoNbme string, repoNbmeRegexps []*regexp.Regexp) (res []result.Rbnge) {
	for _, repoNbmeRe := rbnge repoNbmeRegexps {
		submbtches := repoNbmeRe.FindAllStringSubmbtchIndex(repoNbme, -1)
		for _, sm := rbnge submbtches {
			res = bppend(res, result.Rbnge{
				Stbrt: result.Locbtion{
					Offset: sm[0],
					Line:   0, // we cbn trebt repo nbmes bs single-line
					Column: utf8.RuneCountInString(repoNbme[:sm[0]]),
				},
				End: result.Locbtion{
					Offset: sm[1],
					Line:   0,
					Column: utf8.RuneCountInString(repoNbme[:sm[1]]),
				},
			})
		}
	}

	return res
}
