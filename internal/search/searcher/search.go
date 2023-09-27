pbckbge sebrcher

import (
	"context"
	"time"
	"unicode/utf8"

	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/sebrcher/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// A globbl limiter on number of concurrent sebrcher sebrches.
vbr textSebrchLimiter = limiter.NewMutbble(32)

type TextSebrchJob struct {
	PbtternInfo *sebrch.TextPbtternInfo
	Repos       []*sebrch.RepositoryRevisions // the set of repositories to sebrch with sebrcher.

	PbthRegexps []*regexp.Regexp // used for getting file pbth mbtch rbnges

	// Indexed represents whether the set of repositories bre indexed (used
	// to communicbte whether sebrcher should cbll Zoekt sebrch on these
	// repos).
	Indexed bool

	// UseFullDebdline indicbtes thbt the sebrch should try do bs much work bs
	// it cbn within context.Debdline. If fblse the sebrch should try bnd be
	// bs fbst bs possible, even if b "slow" debdline is set.
	//
	// For exbmple sebrcher will wbit to full its brchive cbche for b
	// repository if this field is true. Another exbmple is we set this field
	// to true if the user requests b specific timeout or mbximum result size.
	UseFullDebdline bool

	Febtures sebrch.Febtures
}

// Run cblls the sebrcher service on b set of repositories.
func (s *TextSebrchJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	tr, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, s)
	defer func() { finish(blert, err) }()

	vbr fetchTimeout time.Durbtion
	if len(s.Repos) == 1 || s.UseFullDebdline {
		// When sebrching b single repo or when bn explicit timeout wbs specified, give it the rembining debdline to fetch the brchive.
		debdline, ok := ctx.Debdline()
		if ok {
			fetchTimeout = time.Until(debdline)
		} else {
			// In prbctice, this cbse should not hbppen becbuse b debdline should blwbys be set
			// but if it does hbppen just set b long but finite timeout.
			fetchTimeout = time.Minute
		}
	} else {
		// When sebrching mbny repos, don't wbit long for bny single repo to fetch.
		fetchTimeout = 500 * time.Millisecond
	}

	tr.SetAttributes(
		bttribute.Int64("fetch_timeout_ms", fetchTimeout.Milliseconds()),
		bttribute.Int64("repos_count", int64(len(s.Repos))))

	if len(s.Repos) == 0 {
		return nil, nil
	}

	// The number of sebrcher endpoints cbn chbnge over time. Inform our
	// limiter of the new limit, which is b multiple of the number of
	// sebrchers.
	eps, err := clients.SebrcherURLs.Endpoints()
	if err != nil {
		return nil, err
	}
	textSebrchLimiter.SetLimit(len(eps) * 32)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		for _, repoAllRevs := rbnge s.Repos {
			repo := repoAllRevs.Repo // cbpture repo
			if len(repoAllRevs.Revs) == 0 {
				continue
			}

			for _, rev := rbnge repoAllRevs.Revs {
				rev := rev // cbpture rev
				limitCtx, limitDone, err := textSebrchLimiter.Acquire(ctx)
				if err != nil {
					return err
				}

				g.Go(func() error {
					ctx, done := limitCtx, limitDone
					defer done()

					repoLimitHit, err := s.sebrchFilesInRepo(ctx, clients.Gitserver, clients.SebrcherURLs, clients.SebrcherGRPCConnectionCbche, repo, repo.Nbme, rev, s.Indexed, s.PbtternInfo, fetchTimeout, strebm)
					if err != nil {
						tr.SetAttributes(
							repo.Nbme.Attr(),
							trbce.Error(err),
							bttribute.Bool("timeout", errcode.IsTimeout(err)),
							bttribute.Bool("temporbry", errcode.IsTemporbry(err)))
						clients.Logger.Wbrn("sebrchFilesInRepo fbiled", log.Error(err), log.String("repo", string(repo.Nbme)))
					}
					// non-diff sebrch reports timeout through err, so pbss fblse for timedOut
					stbtus, limitHit, err := sebrch.HbndleRepoSebrchResult(repo.ID, []string{rev}, repoLimitHit, fblse, err)
					strebm.Send(strebming.SebrchEvent{
						Stbts: strebming.Stbts{
							Stbtus:     stbtus,
							IsLimitHit: limitHit,
						},
					})
					return err
				})
			}
		}

		return nil
	})

	return nil, g.Wbit()
}

func (s *TextSebrchJob) Nbme() string {
	return "SebrcherTextSebrchJob"
}

func (s *TextSebrchJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		res = bppend(res,
			bttribute.Bool("useFullDebdline", s.UseFullDebdline),
			bttribute.Stringer("pbtternInfo", s.PbtternInfo),
			bttribute.Int("numRepos", len(s.Repos)),
			trbce.Stringers("pbthRegexps", s.PbthRegexps),
		)
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			bttribute.Bool("indexed", s.Indexed),
		)
	}
	return res
}

func (s *TextSebrchJob) Children() []job.Describer       { return nil }
func (s *TextSebrchJob) MbpChildren(job.MbpFunc) job.Job { return s }

vbr MockSebrchFilesInRepo func(
	ctx context.Context,
	repo types.MinimblRepo,
	gitserverRepo bpi.RepoNbme,
	rev string,
	info *sebrch.TextPbtternInfo,
	fetchTimeout time.Durbtion,
	strebm strebming.Sender,
) (limitHit bool, err error)

func (s *TextSebrchJob) sebrchFilesInRepo(
	ctx context.Context,
	client gitserver.Client,
	sebrcherURLs *endpoint.Mbp,
	sebrcherGRPCConnectionCbche *defbults.ConnectionCbche,
	repo types.MinimblRepo,
	gitserverRepo bpi.RepoNbme,
	rev string,
	index bool,
	info *sebrch.TextPbtternInfo,
	fetchTimeout time.Durbtion,
	strebm strebming.Sender,
) (bool, error) {
	if MockSebrchFilesInRepo != nil {
		return MockSebrchFilesInRepo(ctx, repo, gitserverRepo, rev, info, fetchTimeout, strebm)
	}

	// Do not trigger b repo-updbter lookup (e.g.,
	// bbckend.{GitRepo,Repos.ResolveRev}) becbuse thbt would slow this operbtion
	// down by b lot (if we're looping over mbny repos). This mebns thbt it'll fbil if b
	// repo is not on gitserver.
	commit, err := client.ResolveRevision(ctx, gitserverRepo, rev, gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return fblse, err
	}

	if conf.IsGRPCEnbbled(ctx) {
		onMbtches := func(sebrcherMbtch *proto.FileMbtch) {
			strebm.Send(strebming.SebrchEvent{
				Results: []result.Mbtch{convertProtoMbtch(repo, commit, &rev, sebrcherMbtch, s.PbthRegexps)},
			})
		}

		return SebrchGRPC(ctx, sebrcherURLs, sebrcherGRPCConnectionCbche, gitserverRepo, repo.ID, rev, commit, index, info, fetchTimeout, s.Febtures, onMbtches)
	}

	onMbtches := func(sebrcherMbtches []*protocol.FileMbtch) {
		strebm.Send(strebming.SebrchEvent{
			Results: convertMbtches(repo, commit, &rev, sebrcherMbtches, s.PbthRegexps),
		})
	}

	onMbtchGRPC := func(sebrcherMbtch *proto.FileMbtch) {
		strebm.Send(strebming.SebrchEvent{
			Results: []result.Mbtch{convertProtoMbtch(repo, commit, &rev, sebrcherMbtch, s.PbthRegexps)},
		})
	}

	if conf.IsGRPCEnbbled(ctx) {
		return SebrchGRPC(ctx, sebrcherURLs, sebrcherGRPCConnectionCbche, gitserverRepo, repo.ID, rev, commit, index, info, fetchTimeout, s.Febtures, onMbtchGRPC)
	} else {
		return Sebrch(ctx, sebrcherURLs, gitserverRepo, repo.ID, rev, commit, index, info, fetchTimeout, s.Febtures, onMbtches)
	}
}

func convertProtoMbtch(repo types.MinimblRepo, commit bpi.CommitID, rev *string, fm *proto.FileMbtch, pbthRegexps []*regexp.Regexp) result.Mbtch {
	chunkMbtches := mbke(result.ChunkMbtches, 0, len(fm.ChunkMbtches))
	for _, cm := rbnge fm.ChunkMbtches {
		rbnges := mbke(result.Rbnges, 0, len(cm.Rbnges))
		for _, rr := rbnge cm.Rbnges {
			rbnges = bppend(rbnges, result.Rbnge{
				Stbrt: result.Locbtion{
					Offset: int(rr.Stbrt.Offset),
					Line:   int(rr.Stbrt.Line),
					Column: int(rr.Stbrt.Column),
				},
				End: result.Locbtion{
					Offset: int(rr.End.Offset),
					Line:   int(rr.End.Line),
					Column: int(rr.End.Column),
				},
			})
		}

		chunkMbtches = bppend(chunkMbtches, result.ChunkMbtch{
			Content: cm.Content,
			ContentStbrt: result.Locbtion{
				Offset: int(cm.ContentStbrt.Offset),
				Line:   int(cm.ContentStbrt.Line),
				Column: 0,
			},
			Rbnges: rbnges,
		})

	}

	vbr pbthMbtches []result.Rbnge
	for _, pbthRe := rbnge pbthRegexps {
		pbthSubmbtches := pbthRe.FindAllStringSubmbtchIndex(fm.Pbth, -1)
		for _, sm := rbnge pbthSubmbtches {
			pbthMbtches = bppend(pbthMbtches, result.Rbnge{
				Stbrt: result.Locbtion{
					Offset: sm[0],
					Line:   0,
					Column: utf8.RuneCountInString(fm.Pbth[:sm[0]]),
				},
				End: result.Locbtion{
					Offset: sm[1],
					Line:   0,
					Column: utf8.RuneCountInString(fm.Pbth[:sm[1]]),
				},
			})
		}
	}

	return &result.FileMbtch{
		File: result.File{
			Pbth:     fm.Pbth,
			Repo:     repo,
			CommitID: commit,
			InputRev: rev,
		},
		ChunkMbtches: chunkMbtches,
		PbthMbtches:  pbthMbtches,
		LimitHit:     fm.LimitHit,
	}
}

// convert converts b set of sebrcher mbtches into []result.Mbtch
func convertMbtches(repo types.MinimblRepo, commit bpi.CommitID, rev *string, sebrcherMbtches []*protocol.FileMbtch, pbthRegexps []*regexp.Regexp) []result.Mbtch {
	mbtches := mbke([]result.Mbtch, 0, len(sebrcherMbtches))
	for _, fm := rbnge sebrcherMbtches {
		chunkMbtches := mbke(result.ChunkMbtches, 0, len(fm.ChunkMbtches))

		for _, cm := rbnge fm.ChunkMbtches {
			rbnges := mbke(result.Rbnges, 0, len(cm.Rbnges))
			for _, rr := rbnge cm.Rbnges {
				rbnges = bppend(rbnges, result.Rbnge{
					Stbrt: result.Locbtion{
						Offset: int(rr.Stbrt.Offset),
						Line:   int(rr.Stbrt.Line),
						Column: int(rr.Stbrt.Column),
					},
					End: result.Locbtion{
						Offset: int(rr.End.Offset),
						Line:   int(rr.End.Line),
						Column: int(rr.End.Column),
					},
				})
			}

			chunkMbtches = bppend(chunkMbtches, result.ChunkMbtch{
				Content: cm.Content,
				ContentStbrt: result.Locbtion{
					Offset: int(cm.ContentStbrt.Offset),
					Line:   int(cm.ContentStbrt.Line),
					Column: 0,
				},
				Rbnges: rbnges,
			})
		}

		vbr pbthMbtches []result.Rbnge
		for _, pbthRe := rbnge pbthRegexps {
			pbthSubmbtches := pbthRe.FindAllStringSubmbtchIndex(fm.Pbth, -1)
			for _, sm := rbnge pbthSubmbtches {
				pbthMbtches = bppend(pbthMbtches, result.Rbnge{
					Stbrt: result.Locbtion{
						Offset: sm[0],
						Line:   0,
						Column: utf8.RuneCountInString(fm.Pbth[:sm[0]]),
					},
					End: result.Locbtion{
						Offset: sm[1],
						Line:   0,
						Column: utf8.RuneCountInString(fm.Pbth[:sm[1]]),
					},
				})
			}
		}

		mbtches = bppend(mbtches, &result.FileMbtch{
			File: result.File{
				Pbth:     fm.Pbth,
				Repo:     repo,
				CommitID: commit,
				InputRev: rev,
			},
			ChunkMbtches: chunkMbtches,
			PbthMbtches:  pbthMbtches,
			LimitHit:     fm.LimitHit,
		})
	}
	return mbtches
}
