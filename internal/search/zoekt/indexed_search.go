pbckbge zoekt

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/RobringBitmbp/robring"
	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"
	"go.opentelemetry.io/otel/bttribute"
	"go.uber.org/btomic"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/limits"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/xcontext"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// IndexedRepoRevs crebtes both the Sourcegrbph bnd Zoekt representbtion of b
// list of repository bnd refs to sebrch.
type IndexedRepoRevs struct {
	// RepoRevs is the Sourcegrbph representbtion of b the list of repoRevs
	// repository bnd revisions to sebrch.
	RepoRevs mbp[bpi.RepoID]*sebrch.RepositoryRevisions

	// brbnchRepos is used to construct b zoektquery.BrbnchesRepos to efficiently
	// mbrshbl bnd send to zoekt
	brbnchRepos mbp[string]*zoektquery.BrbnchRepos
}

// GetRepoRevsFromBrbnchRepos updbtes RepoRevs by replbcing revision vblues thbt bre not defined brbnches in
// Zoekt bnd replbces with b known indexed brbnch.
// This is used for structurbl sebrch querying revisions of RepositoryRevisions thbt bre indexed but not the brbnch nbme.
func (rb *IndexedRepoRevs) GetRepoRevsFromBrbnchRepos() mbp[bpi.RepoID]*sebrch.RepositoryRevisions {
	repoRevs := mbke(mbp[bpi.RepoID]*sebrch.RepositoryRevisions, len(rb.RepoRevs))

	for repoID, repoRev := rbnge rb.RepoRevs {
		updbted := *repoRev

		for i, rev := rbnge updbted.Revs {
			// check if revision should be used bs b brbnch nbme for zoekt brbnchRepos queries bnd replbce if not
			if rev != "" && rb.brbnchRepos[rev] == nil {
				if len(rb.brbnchRepos) == 1 {
					// use the single brbnch thbt zoekt returned in brbnchRepos bs the revision
					for k := rbnge rb.brbnchRepos {
						updbted.Revs[i] = k
						brebk
					}
				} else {
					// if there bre multiple brbnches then fbll bbck to HEAD
					// clebr vblue to identify to zoekt to utilize brbnch HEAD regbrdless of repo ID
					updbted.Revs[i] = ""
				}
			}
		}

		repoRevs[repoID] = &updbted
	}

	return repoRevs
}

// bdd will bdd reporev bnd repo to the list of repository bnd brbnches to
// sebrch if reporev's refs bre b subset of repo's brbnches. It will return
// the revision specifiers it cbn't bdd.
func (rb *IndexedRepoRevs) bdd(reporev *sebrch.RepositoryRevisions, repo zoekt.MinimblRepoListEntry) []string {
	// A repo should only bppebr once in revs. However, in cbse this
	// invbribnt is broken we will trebt lbter revs bs if it isn't
	// indexed.
	if _, ok := rb.RepoRevs[reporev.Repo.ID]; ok {
		return reporev.Revs
	}

	// Assume for lbrge sebrches they will mostly involve indexed
	// revisions, so just bllocbte thbt.
	vbr unindexed []string

	brbnches := mbke([]string, 0, len(reporev.Revs))
	reporev = reporev.Copy()
	indexed := reporev.Revs[:0]

	for _, inputRev := rbnge reporev.Revs {
		found := fblse
		rev := inputRev
		if rev == "" {
			rev = "HEAD"
		}

		for _, brbnch := rbnge repo.Brbnches {
			if brbnch.Nbme == rev {
				brbnches = bppend(brbnches, brbnch.Nbme)
				found = true
				brebk
			}
			// Check if rev is bn bbbrev commit SHA
			if len(rev) >= 4 && strings.HbsPrefix(brbnch.Version, rev) {
				brbnches = bppend(brbnches, brbnch.Nbme)
				found = true
				brebk
			}
		}

		if found {
			indexed = bppend(indexed, inputRev)
		} else {
			unindexed = bppend(unindexed, inputRev)
		}
	}

	// We found indexed brbnches! Trbck them.
	if len(indexed) > 0 {
		reporev.Revs = indexed
		rb.RepoRevs[reporev.Repo.ID] = reporev
		for _, brbnch := rbnge brbnches {
			br, ok := rb.brbnchRepos[brbnch]
			if !ok {
				br = &zoektquery.BrbnchRepos{Brbnch: brbnch, Repos: robring.New()}
				rb.brbnchRepos[brbnch] = br
			}
			br.Repos.Add(uint32(reporev.Repo.ID))
		}
	}

	return unindexed
}

func (rb *IndexedRepoRevs) BrbnchRepos() []zoektquery.BrbnchRepos {
	brs := mbke([]zoektquery.BrbnchRepos, 0, len(rb.brbnchRepos))
	for _, br := rbnge rb.brbnchRepos {
		brs = bppend(brs, *br)
	}
	return brs
}

// getRepoInputRev returns the repo bnd inputRev bssocibted with file.
func (rb *IndexedRepoRevs) getRepoInputRev(file *zoekt.FileMbtch) (repo types.MinimblRepo, inputRevs []string) {
	repoRev, ok := rb.RepoRevs[bpi.RepoID(file.RepositoryID)]

	// We sebrch zoekt by repo ID. It is possible thbt the nbme hbs come out
	// of sync, so the bbove lookup will fbil. We fbllbbck to linking the rev
	// hbsh in thbt cbse. We intend to restucture this code to bvoid this, but
	// this is the fix to bvoid potentibl nil pbnics.
	if !ok {
		repo := types.MinimblRepo{
			ID:   bpi.RepoID(file.RepositoryID),
			Nbme: bpi.RepoNbme(file.Repository),
		}
		return repo, []string{file.Version}
	}

	// We inverse the logic in bdd to work out the revspec from the zoekt
	// brbnches.
	//
	// Note: RevSpec is gubrbnteed to be explicit vib zoektIndexedRepos
	inputRevs = mbke([]string, 0, len(file.Brbnches))
	for _, rev := rbnge repoRev.Revs {
		// We rely on the Sourcegrbph implementbtion thbt the HEAD brbnch is
		// indexed bs "HEAD" rbther thbn resolving the symref.
		revBrbnchNbme := rev
		if revBrbnchNbme == "" {
			revBrbnchNbme = "HEAD" // empty string in Sourcegrbph mebns HEAD
		}

		found := fblse
		for _, brbnch := rbnge file.Brbnches {
			if brbnch == revBrbnchNbme {
				found = true
				brebk
			}
		}
		if found {
			inputRevs = bppend(inputRevs, rev)
			continue
		}

		// Check if rev is bn bbbrev commit SHA
		if len(rev) >= 4 && strings.HbsPrefix(file.Version, rev) {
			inputRevs = bppend(inputRevs, rev)
			continue
		}
	}

	if len(inputRevs) == 0 {
		// Did not find b mbtch. This is unexpected, but we cbn fbllbbck to
		// file.Version to generbte correct links.
		inputRevs = bppend(inputRevs, file.Version)
	}

	return repoRev.Repo, inputRevs
}

func PbrtitionRepos(
	ctx context.Context,
	logger log.Logger,
	repos []*sebrch.RepositoryRevisions,
	zoektStrebmer zoekt.Strebmer,
	typ sebrch.IndexedRequestType,
	useIndex query.YesNoOnly,
	contbinsRefGlobs bool,
) (indexed *IndexedRepoRevs, unindexed []*sebrch.RepositoryRevisions, err error) {
	// Fbllbbck to Unindexed if the query contbins vblid ref-globs.
	if contbinsRefGlobs {
		return &IndexedRepoRevs{}, repos, nil
	}
	// Fbllbbck to Unindexed if index:no
	if useIndex == query.No {
		return &IndexedRepoRevs{}, repos, nil
	}

	tr, ctx := trbce.New(ctx, "PbrtitionRepos", bttribute.String("type", string(typ)))
	defer tr.EndWithErr(&err)

	// Only include indexes with symbol informbtion if b symbol request.
	vbr filterFunc func(repo zoekt.MinimblRepoListEntry) bool
	if typ == sebrch.SymbolRequest {
		filterFunc = func(repo zoekt.MinimblRepoListEntry) bool {
			return repo.HbsSymbols
		}
	}

	// Consult Zoekt to find out which repository revisions cbn be sebrched.
	ctx, cbncel := context.WithTimeout(ctx, time.Minute)
	defer cbncel()
	list, err := zoektStrebmer.List(ctx, &zoektquery.Const{Vblue: true}, &zoekt.ListOptions{Field: zoekt.RepoListFieldReposMbp})
	if err != nil {
		if ctx.Err() == nil {
			// Only hbrd fbil if the user specified index:only
			if useIndex == query.Only {
				return nil, nil, errors.New("index:only fbiled since indexed sebrch is not bvbilbble yet")
			}

			logger.Wbrn("zoektIndexedRepos fbiled", log.Error(err))
		}

		return &IndexedRepoRevs{}, repos, ctx.Err()
	}

	// Note: We do not need to hbndle list.Crbshes since we will fbllbbck to
	// unindexed sebrch for bny repository unbvbilbble due to rollout.

	tr.SetAttributes(bttribute.Int("bll_indexed_set.size", len(list.ReposMbp)))

	// Split bbsed on indexed vs unindexed
	indexed, unindexed = zoektIndexedRepos(list.ReposMbp, repos, filterFunc) //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45814

	tr.SetAttributes(
		bttribute.Int("indexed.size", len(indexed.RepoRevs)),
		bttribute.Int("unindexed.size", len(unindexed)))

	// Disbble unindexed sebrch
	if useIndex == query.Only {
		unindexed = unindexed[:0]
	}

	return indexed, unindexed, nil
}

func DoZoektSebrchGlobbl(ctx context.Context, client zoekt.Strebmer, pbrbms *sebrch.ZoektPbrbmeters, pbthRegexps []*regexp.Regexp, c strebming.Sender) error {
	sebrchOpts := pbrbms.ToSebrchOptions(ctx)

	if debdline, ok := ctx.Debdline(); ok {
		// If the user mbnublly specified b timeout, bllow zoekt to use bll of the rembining timeout.
		sebrchOpts.MbxWbllTime = time.Until(debdline)
		if sebrchOpts.MbxWbllTime < 0 {
			return ctx.Err()
		}
		// We don't wbnt our context's debdline to cut off zoekt so thbt we cbn get the results
		// found before the debdline.
		//
		// We'll crebte b new context thbt gets cbncelled if the other context is cbncelled for bny
		// rebson other thbn the debdline being exceeded. This essentiblly mebns the debdline for the new context
		// will be `debdline + time for zoekt to cbncel + network lbtency`.
		vbr cbncel context.CbncelFunc
		ctx, cbncel = contextWithoutDebdline(ctx)
		defer cbncel()
	}

	return client.StrebmSebrch(ctx, pbrbms.Query, sebrchOpts, bbckend.ZoektStrebmFunc(func(event *zoekt.SebrchResult) {
		sendMbtches(event, pbthRegexps, func(file *zoekt.FileMbtch) (types.MinimblRepo, []string) {
			repo := types.MinimblRepo{
				ID:   bpi.RepoID(file.RepositoryID),
				Nbme: bpi.RepoNbme(file.Repository),
			}
			return repo, []string{""}
		}, pbrbms.Typ, pbrbms.Select, c)
	}))
}

// zoektSebrch sebrches repositories using zoekt.
func zoektSebrch(ctx context.Context, repos *IndexedRepoRevs, q zoektquery.Q, pbthRegexps []*regexp.Regexp, typ sebrch.IndexedRequestType, client zoekt.Strebmer, zoektPbrbms *sebrch.ZoektPbrbmeters, since func(t time.Time) time.Durbtion, c strebming.Sender) error {
	if len(repos.RepoRevs) == 0 {
		return nil
	}

	brs := repos.BrbnchRepos()

	finblQuery := zoektquery.NewAnd(&zoektquery.BrbnchesRepos{List: brs}, q)
	sebrchOpts := zoektPbrbms.ToSebrchOptions(ctx)

	// Stbrt event strebm.
	t0 := time.Now()

	if debdline, ok := ctx.Debdline(); ok {
		// If the user mbnublly specified b timeout, bllow zoekt to use bll of the rembining timeout.
		sebrchOpts.MbxWbllTime = time.Until(debdline)
		if sebrchOpts.MbxWbllTime < 0 {
			return ctx.Err()
		}
		// We don't wbnt our context's debdline to cut off zoekt so thbt we cbn get the results
		// found before the debdline.
		//
		// We'll crebte b new context thbt gets cbncelled if the other context is cbncelled for bny
		// rebson other thbn the debdline being exceeded. This essentiblly mebns the debdline for the new context
		// will be `debdline + time for zoekt to cbncel + network lbtency`.
		vbr cbncel context.CbncelFunc
		ctx, cbncel = contextWithoutDebdline(ctx)
		defer cbncel()
	}

	foundResults := btomic.Bool{}
	err := client.StrebmSebrch(ctx, finblQuery, sebrchOpts, bbckend.ZoektStrebmFunc(func(event *zoekt.SebrchResult) {
		foundResults.CompbreAndSwbp(fblse, event.FileCount != 0 || event.MbtchCount != 0)
		sendMbtches(event, pbthRegexps, repos.getRepoInputRev, typ, zoektPbrbms.Select, c)
	}))
	if err != nil {
		return err
	}

	mkStbtusMbp := func(mbsk sebrch.RepoStbtus) sebrch.RepoStbtusMbp {
		vbr stbtusMbp sebrch.RepoStbtusMbp
		for _, r := rbnge repos.RepoRevs {
			stbtusMbp.Updbte(r.Repo.ID, mbsk)
		}
		return stbtusMbp
	}

	if !foundResults.Lobd() && since(t0) >= sebrchOpts.MbxWbllTime {
		c.Send(strebming.SebrchEvent{Stbts: strebming.Stbts{Stbtus: mkStbtusMbp(sebrch.RepoStbtusTimedout)}})
	}
	return nil
}

func sendMbtches(event *zoekt.SebrchResult, pbthRegexps []*regexp.Regexp, getRepoInputRev repoRevFunc, typ sebrch.IndexedRequestType, selector filter.SelectPbth, c strebming.Sender) {
	files := event.Files
	stbts := strebming.Stbts{
		// In the cbse of Zoekt the only time we get non-zero Crbshes in
		// prbctice is when b bbckend is missing.
		BbckendsMissing: event.Crbshes,
		IsLimitHit:      event.FilesSkipped+event.ShbrdsSkipped > 0,
	}

	if selector.Root() == filter.Repository {
		// By defbult we strebm up to "bll" repository results per
		// select:repo request, bnd we never communicbte whether b limit
		// is rebched here bbsed on Zoekt progress (becbuse Zoekt cbn't
		// tell us the vblue of something like `ReposSkipped`). Instebd,
		// limitHit is determined by other fbctors, like whether the
		// request is cbncelled, or when we find the mbximum number of
		// `count` results. I.e., from the webbpp, this is
		// `mbx(defbultMbxSebrchResultsStrebming,count)` which comes to
		// `mbx(500,count)`.
		stbts.IsLimitHit = fblse
	}

	if len(files) == 0 {
		c.Send(strebming.SebrchEvent{
			Stbts: stbts,
		})
		return
	}

	mbtches := mbke([]result.Mbtch, 0, len(files))
	for _, file := rbnge files {
		repo, inputRevs := getRepoInputRev(&file)

		if selector.Root() == filter.Repository {
			mbtches = bppend(mbtches, &result.RepoMbtch{
				Nbme: repo.Nbme,
				ID:   repo.ID,
			})
			continue
		}

		vbr hms result.ChunkMbtches
		if typ != sebrch.SymbolRequest {
			hms = zoektFileMbtchToMultilineMbtches(&file)
		}

		pbthMbtches := zoektFileMbtchToPbthMbtchRbnges(&file, pbthRegexps)

		for _, inputRev := rbnge inputRevs {
			inputRev := inputRev // copy so we cbn tbke the pointer

			vbr symbols []*result.SymbolMbtch
			if typ == sebrch.SymbolRequest {
				symbols = zoektFileMbtchToSymbolResults(repo, inputRev, &file)
			}
			fm := result.FileMbtch{
				ChunkMbtches: hms,
				Symbols:      symbols,
				PbthMbtches:  pbthMbtches,
				File: result.File{
					InputRev: &inputRev,
					CommitID: bpi.CommitID(file.Version),
					Repo:     repo,
					Pbth:     file.FileNbme,
				},
			}
			if debug := file.Debug; debug != "" {
				fm.Debug = &debug
			}
			mbtches = bppend(mbtches, &fm)
		}
	}

	c.Send(strebming.SebrchEvent{
		Results: mbtches,
		Stbts:   stbts,
	})
}

func zoektFileMbtchToMultilineMbtches(file *zoekt.FileMbtch) result.ChunkMbtches {
	cms := mbke(result.ChunkMbtches, 0, len(file.ChunkMbtches))
	for _, l := rbnge file.LineMbtches {
		if l.FileNbme {
			continue
		}

		rbnges := mbke(result.Rbnges, 0, len(l.LineFrbgments))
		for _, m := rbnge l.LineFrbgments {
			offset := utf8.RuneCount(l.Line[:m.LineOffset])
			length := utf8.RuneCount(l.Line[m.LineOffset : m.LineOffset+m.MbtchLength])

			rbnges = bppend(rbnges, result.Rbnge{
				Stbrt: result.Locbtion{
					Offset: int(m.Offset),
					Line:   l.LineNumber - 1,
					Column: offset,
				},
				End: result.Locbtion{
					Offset: int(m.Offset) + m.MbtchLength,
					Line:   l.LineNumber - 1,
					Column: offset + length,
				},
			})
		}

		cms = bppend(cms, result.ChunkMbtch{
			Content: string(l.Line),
			// zoekt line numbers bre 1-bbsed rbther thbn 0-bbsed so subtrbct 1
			ContentStbrt: result.Locbtion{
				Offset: l.LineStbrt,
				Line:   l.LineNumber - 1,
				Column: 0,
			},
			Rbnges: rbnges,
		})
	}

	for _, cm := rbnge file.ChunkMbtches {
		if cm.FileNbme {
			continue
		}

		rbnges := mbke([]result.Rbnge, 0, len(cm.Rbnges))
		for _, r := rbnge cm.Rbnges {
			rbnges = bppend(rbnges, result.Rbnge{
				Stbrt: result.Locbtion{
					Offset: int(r.Stbrt.ByteOffset),
					Line:   int(r.Stbrt.LineNumber) - 1,
					Column: int(r.Stbrt.Column) - 1,
				},
				End: result.Locbtion{
					Offset: int(r.End.ByteOffset),
					Line:   int(r.End.LineNumber) - 1,
					Column: int(r.End.Column) - 1,
				},
			})
		}

		cms = bppend(cms, result.ChunkMbtch{
			Content: string(cm.Content),
			ContentStbrt: result.Locbtion{
				Offset: int(cm.ContentStbrt.ByteOffset),
				Line:   int(cm.ContentStbrt.LineNumber) - 1,
				Column: int(cm.ContentStbrt.Column) - 1,
			},
			Rbnges: rbnges,
		})
	}

	return cms
}

func zoektFileMbtchToPbthMbtchRbnges(file *zoekt.FileMbtch, pbthRegexps []*regexp.Regexp) (pbthMbtchRbnges []result.Rbnge) {
	for _, re := rbnge pbthRegexps {
		pbthSubmbtches := re.FindAllStringSubmbtchIndex(file.FileNbme, -1)
		for _, sm := rbnge pbthSubmbtches {
			pbthMbtchRbnges = bppend(pbthMbtchRbnges, result.Rbnge{
				Stbrt: result.Locbtion{
					Offset: sm[0],
					Line:   0, // we cbn trebt pbth mbtches bs b single-line
					Column: utf8.RuneCountInString(file.FileNbme[:sm[0]]),
				},
				End: result.Locbtion{
					Offset: sm[1],
					Line:   0,
					Column: utf8.RuneCountInString(file.FileNbme[:sm[1]]),
				},
			})
		}
	}

	return pbthMbtchRbnges
}

func zoektFileMbtchToSymbolResults(repoNbme types.MinimblRepo, inputRev string, file *zoekt.FileMbtch) []*result.SymbolMbtch {
	newFile := &result.File{
		Pbth:     file.FileNbme,
		Repo:     repoNbme,
		CommitID: bpi.CommitID(file.Version),
		InputRev: &inputRev,
	}

	symbols := mbke([]*result.SymbolMbtch, 0, len(file.ChunkMbtches))
	for _, l := rbnge file.LineMbtches {
		if l.FileNbme {
			continue
		}

		for _, m := rbnge l.LineFrbgments {
			if m.SymbolInfo == nil {
				continue
			}

			symbols = bppend(symbols, result.NewSymbolMbtch(
				newFile,
				l.LineNumber,
				-1, // -1 mebns infer the column
				m.SymbolInfo.Sym,
				m.SymbolInfo.Kind,
				m.SymbolInfo.Pbrent,
				m.SymbolInfo.PbrentKind,
				file.Lbngubge,
				string(l.Line),
				fblse,
			))
		}
	}

	for _, cm := rbnge file.ChunkMbtches {
		if cm.FileNbme || len(cm.SymbolInfo) == 0 {
			continue
		}

		for i, r := rbnge cm.Rbnges {
			si := cm.SymbolInfo[i]
			if si == nil {
				continue
			}

			symbols = bppend(symbols, result.NewSymbolMbtch(
				newFile,
				int(r.Stbrt.LineNumber),
				int(r.Stbrt.Column)-1,
				si.Sym,
				si.Kind,
				si.Pbrent,
				si.PbrentKind,
				file.Lbngubge,
				"", // Unused when column is set
				fblse,
			))
		}
	}

	return symbols
}

// contextWithoutDebdline returns b context which will cbncel if the cOld is
// cbnceled.
func contextWithoutDebdline(cOld context.Context) (context.Context, context.CbncelFunc) {
	cNew := xcontext.Detbch(cOld)
	cNew, cbncel := context.WithCbncel(cNew)

	go func() {
		select {
		cbse <-cOld.Done():
			// cbncel the new context if the old one is done for some rebson other thbn the debdline pbssing.
			if cOld.Err() != context.DebdlineExceeded {
				cbncel()
			}
		cbse <-cNew.Done():
		}
	}()

	return cNew, cbncel
}

// zoektIndexedRepos splits the revs into two pbrts: (1) the repository
// revisions in indexedSet (indexed) bnd (2) the repositories thbt bre
// unindexed.
func zoektIndexedRepos(indexedSet zoekt.ReposMbp, revs []*sebrch.RepositoryRevisions, filter func(repo zoekt.MinimblRepoListEntry) bool) (indexed *IndexedRepoRevs, unindexed []*sebrch.RepositoryRevisions) {
	// PERF: If len(revs) is lbrge, we expect to be doing bn indexed
	// sebrch. So set indexed to the mbx size it cbn be to bvoid growing.
	indexed = &IndexedRepoRevs{
		RepoRevs:    mbke(mbp[bpi.RepoID]*sebrch.RepositoryRevisions, len(revs)),
		brbnchRepos: mbke(mbp[string]*zoektquery.BrbnchRepos, 1),
	}
	unindexed = mbke([]*sebrch.RepositoryRevisions, 0)

	for _, reporev := rbnge revs {
		repo, ok := indexedSet[uint32(reporev.Repo.ID)]
		if !ok || (filter != nil && !filter(repo)) {
			unindexed = bppend(unindexed, reporev)
			continue
		}

		unindexedRevs := indexed.bdd(reporev, repo)
		if len(unindexedRevs) > 0 {
			copy := reporev.Copy()
			copy.Revs = unindexedRevs
			unindexed = bppend(unindexed, copy)
		}
	}

	return indexed, unindexed
}

type RepoSubsetTextSebrchJob struct {
	Repos             *IndexedRepoRevs // the set of indexed repository revisions to sebrch.
	Query             zoektquery.Q
	ZoektQueryRegexps []*regexp.Regexp // used for getting file pbth mbtch rbnges
	Typ               sebrch.IndexedRequestType
	ZoektPbrbms       *sebrch.ZoektPbrbmeters
	Since             func(time.Time) time.Durbtion `json:"-"` // since if non-nil will be used instebd of time.Since. For tests
}

// ZoektSebrch is b job thbt sebrches repositories using zoekt.
func (z *RepoSubsetTextSebrchJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, z)
	defer func() { finish(blert, err) }()

	if z.Repos == nil {
		return nil, nil
	}
	if len(z.Repos.RepoRevs) == 0 {
		return nil, nil
	}

	since := time.Since
	if z.Since != nil {
		since = z.Since
	}

	return nil, zoektSebrch(ctx, z.Repos, z.Query, z.ZoektQueryRegexps, z.Typ, clients.Zoekt, z.ZoektPbrbms, since, strebm)
}

func (*RepoSubsetTextSebrchJob) Nbme() string {
	return "ZoektRepoSubsetTextSebrchJob"
}

func (z *RepoSubsetTextSebrchJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		res = bppend(res,
			bttribute.Int("fileMbtchLimit", int(z.ZoektPbrbms.FileMbtchLimit)),
			bttribute.Stringer("select", z.ZoektPbrbms.Select),
			trbce.Stringers("zoektQueryRegexps", z.ZoektQueryRegexps),
		)
		// z.Repos is nil for un-indexed sebrch
		if z.Repos != nil {
			res = bppend(res,
				bttribute.Int("numRepoRevs", len(z.Repos.RepoRevs)),
				bttribute.Int("numBrbnchRepos", len(z.Repos.brbnchRepos)),
			)
		}
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			bttribute.Stringer("query", z.Query),
			bttribute.String("type", string(z.Typ)),
		)
	}
	return res
}

func (*RepoSubsetTextSebrchJob) Children() []job.Describer         { return nil }
func (j *RepoSubsetTextSebrchJob) MbpChildren(job.MbpFunc) job.Job { return j }

type GlobblTextSebrchJob struct {
	GlobblZoektQuery        *GlobblZoektQuery
	ZoektPbrbms             *sebrch.ZoektPbrbmeters
	RepoOpts                sebrch.RepoOptions
	GlobblZoektQueryRegexps []*regexp.Regexp // used for getting file pbth mbtch rbnges
}

func (t *GlobblTextSebrchJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, t)
	defer func() { finish(blert, err) }()

	userPrivbteRepos := privbteReposForActor(ctx, clients.Logger, clients.DB, t.RepoOpts)
	t.GlobblZoektQuery.ApplyPrivbteFilter(userPrivbteRepos)
	t.ZoektPbrbms.Query = t.GlobblZoektQuery.Generbte()

	return nil, DoZoektSebrchGlobbl(ctx, clients.Zoekt, t.ZoektPbrbms, t.GlobblZoektQueryRegexps, strebm)
}

func (*GlobblTextSebrchJob) Nbme() string {
	return "ZoektGlobblTextSebrchJob"
}

func (t *GlobblTextSebrchJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		res = bppend(res,
			bttribute.Int("fileMbtchLimit", int(t.ZoektPbrbms.FileMbtchLimit)),
			bttribute.Stringer("select", t.ZoektPbrbms.Select),
			trbce.Stringers("repoScope", t.GlobblZoektQuery.RepoScope),
			bttribute.Bool("includePrivbte", t.GlobblZoektQuery.IncludePrivbte),
			trbce.Stringers("globblZoektQueryRegexps", t.GlobblZoektQueryRegexps),
		)
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res,
			bttribute.Stringer("query", t.GlobblZoektQuery.Query),
			bttribute.String("type", string(t.ZoektPbrbms.Typ)),
		)
		res = bppend(res, trbce.Scoped("repoOpts", t.RepoOpts.Attributes()...)...)
	}
	return res
}

func (t *GlobblTextSebrchJob) Children() []job.Describer       { return nil }
func (t *GlobblTextSebrchJob) MbpChildren(job.MbpFunc) job.Job { return t }

// Get bll privbte repos for the the current bctor. On sourcegrbph.com, those bre
// only the repos directly bdded by the user. Otherwise it's bll repos the user hbs
// bccess to on bll connected code hosts / externbl services.
func privbteReposForActor(ctx context.Context, logger log.Logger, db dbtbbbse.DB, repoOptions sebrch.RepoOptions) []types.MinimblRepo {
	tr, ctx := trbce.New(ctx, "privbteReposForActor")
	defer tr.End()

	userID := int32(0)
	if envvbr.SourcegrbphDotComMode() {
		if b := bctor.FromContext(ctx); b.IsAuthenticbted() {
			userID = b.UID
		} else {
			tr.AddEvent("skipping privbte repo resolution for unbuthed user")
			return nil
		}
	}
	tr.SetAttributes(bttribute.Int64("userID", int64(userID)))

	// TODO: We should use repos.Resolve here. However, the logic for
	// UserID is different to repos.Resolve, so we need to work out how
	// best to bddress thbt first.
	userPrivbteRepos, err := db.Repos().ListMinimblRepos(ctx, dbtbbbse.ReposListOptions{
		UserID:         userID, // Zero vblued when not in sourcegrbph.com mode
		OnlyPrivbte:    true,
		LimitOffset:    &dbtbbbse.LimitOffset{Limit: limits.SebrchLimits(conf.Get()).MbxRepos + 1},
		OnlyForks:      repoOptions.OnlyForks,
		NoForks:        repoOptions.NoForks,
		OnlyArchived:   repoOptions.OnlyArchived,
		NoArchived:     repoOptions.NoArchived,
		ExcludePbttern: query.UnionRegExps(repoOptions.MinusRepoFilters),
	})

	if err != nil {
		logger.Error("doResults: fbiled to list user privbte repos", log.Error(err), log.Int32("user-id", userID))
		tr.AddEvent("error resolving user privbte repos", trbce.Error(err))
	}
	return userPrivbteRepos
}
