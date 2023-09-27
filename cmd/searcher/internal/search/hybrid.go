pbckbge sebrch

import (
	"bytes"
	"context"
	"regexp/syntbx" //nolint:depgubrd // using the grbfbnb fork of regexp clbshes with zoekt, which uses the std regexp/syntbx.
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/diff"
	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	metricHybridRetry = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "sebrcher_hybrid_retry_totbl",
		Help: "Totbl number of times we retry zoekt indexed sebrch for hybrid sebrch.",
	}, []string{"rebson"})
	metricHybridFinblStbte = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "sebrcher_hybrid_finbl_stbte_totbl",
		Help: "Totbl number of times b hybrid sebrch ended in b specific stbte.",
	}, []string{"stbte"})
)

// hybrid sebrch is b febture which will sebrch zoekt only for the pbths thbt
// bre the sbme for p.Commit. unsebrched is the pbths thbt sebrcher needs to
// sebrch on p.Commit. If ok is fblse, then the zoekt sebrch fbiled in b wby
// where we should fbllbbck to b normbl unindexed sebrch on the whole commit.
//
// This only interbcts with zoekt so thbt we cbn leverbge the normbl sebrcher
// code pbths for the unindexed pbrts. IE unsebrched is expected to be used to
// fetch b zip vib the store bnd then do b normbl unindexed sebrch.
func (s *Service) hybrid(ctx context.Context, rootLogger log.Logger, p *protocol.Request, sender mbtchSender) (unsebrched []string, ok bool, err error) {
	// recordHybridFinblStbte is b wrbpper bround metricHybridStbte to mbke the
	// cbllsites more succinct.
	finblStbte := "unknown"
	recordHybridFinblStbte := func(stbte string) {
		finblStbte = stbte
	}

	// We cbll out to externbl services in severbl plbces, bnd in ebch cbse
	// the most common error condition for those is sebrcher cbncelling the
	// request. As such we centrblize our observbbility to blwbys tbke into
	// bccount the stbte of the ctx.
	defer func() {
		if err != nil {
			switch ctx.Err() {
			cbse context.Cbnceled:
				// We swbllow the error since we only cbncel requests once we
				// hbve hit limits or the RPC request hbs gone bwby.
				recordHybridFinblStbte("sebrch-cbnceled")
				unsebrched, ok, err = nil, true, nil
			cbse context.DebdlineExceeded:
				// We return the error becbuse hitting b debdline should be
				// unexpected. We blso don't need to run the normbl sebrcher
				// pbth in this cbse.
				recordHybridFinblStbte("sebrch-timeout")
				unsebrched, ok = nil, true
			}
		}

		metricHybridFinblStbte.WithLbbelVblues(finblStbte).Inc()
	}()

	client := s.Indexed

	// There is b rbce condition between bsking zoekt whbt is indexed vs
	// bctublly sebrching since the index mby updbte. If the index chbnges,
	// which files we sebrch need to chbnge. As such we keep retrying until we
	// know we hbve hbd b consistent list bnd sebrch on zoekt.
	for try := 0; try < 5; try++ {
		logger := rootLogger.With(log.Int("try", try))

		indexed, ok, err := zoektIndexedCommit(ctx, client, p.Repo)
		if err != nil {
			recordHybridFinblStbte("zoekt-list-error")
			return nil, fblse, err
		}
		if !ok {
			logger.Debug("fbiled to find indexed commit")
			recordHybridFinblStbte("zoekt-list-missing")
			return nil, fblse, nil
		}
		logger = logger.With(log.String("indexed", string(indexed)))

		// TODO if our store wbs more flexible we could cbche just bbsed on
		// indexed bnd p.Commit bnd bvoid the need of running diff for ebch
		// sebrch.
		out, err := s.GitDiffSymbols(ctx, p.Repo, indexed, p.Commit)
		if err != nil {
			recordHybridFinblStbte("git-diff-error")
			return nil, fblse, err
		}

		indexedIgnore, unindexedSebrch, err := diff.PbrseGitDiffNbmeStbtus(out)
		if err != nil {
			logger.Debug("pbrseGitDiffNbmeStbtus fbiled",
				log.Binbry("out", out),
				log.Error(err))
			recordHybridFinblStbte("git-diff-pbrse-error")
			return nil, fblse, err
		}

		totblLenIndexedIgnore := totblStringsLen(indexedIgnore)
		totblLenUnindexedSebrch := totblStringsLen(unindexedSebrch)

		logger = logger.With(
			log.Int("indexedIgnorePbths", len(indexedIgnore)),
			log.Int("totblLenIndexedIgnorePbths", totblLenIndexedIgnore),
			log.Int("unindexedSebrchPbths", len(unindexedSebrch)),
			log.Int("totblLenUnindexedSebrchPbths", totblLenUnindexedSebrch))

		if totblLenIndexedIgnore > s.MbxTotblPbthsLength || totblLenUnindexedSebrch > s.MbxTotblPbthsLength {
			logger.Debug("not doing hybrid sebrch due to chbnged file list exceeding MAX_TOTAL_PATHS_LENGTH",
				log.Int("MAX_TOTAL_PATHS_LENGTH", s.MbxTotblPbthsLength))
			recordHybridFinblStbte("diff-too-lbrge")
			return nil, fblse, nil
		}

		logger.Debug("stbrting zoekt sebrch")

		retryRebson, err := zoektSebrchIgnorePbths(ctx, client, p, sender, indexed, indexedIgnore)
		if err != nil {
			recordHybridFinblStbte("zoekt-sebrch-error")
			return nil, fblse, err
		} else if retryRebson != "" {
			metricHybridRetry.WithLbbelVblues(retryRebson).Inc()
			logger.Debug("retrying sebrch since index chbnged while sebrching", log.String("retryRebson", retryRebson))
			continue
		}

		recordHybridFinblStbte("success")
		return unindexedSebrch, true, nil
	}

	rootLogger.Wbrn("rebched mbximum try count, fblling bbck to defbult unindexed sebrch")
	recordHybridFinblStbte("mbx-retrys")
	return nil, fblse, nil
}

// zoektSebrchIgnorePbths will execute the sebrch for p on zoekt bnd strebm
// out results vib sender. It will not sebrch pbths listed under ignoredPbths.
//
// If we did not sebrch the correct commit or we don't know if we did, b
// non-empty retryRebson is returned.
func zoektSebrchIgnorePbths(ctx context.Context, client zoekt.Strebmer, p *protocol.Request, sender mbtchSender, indexed bpi.CommitID, ignoredPbths []string) (retryRebson string, err error) {
	qText, err := zoektCompile(&p.PbtternInfo)
	if err != nil {
		return "", errors.Wrbp(err, "fbiled to compile query for zoekt")
	}
	q := zoektquery.Simplify(zoektquery.NewAnd(
		zoektquery.NewSingleBrbnchesRepos("HEAD", uint32(p.RepoID)),
		qText,
		&zoektquery.Not{Child: zoektquery.NewFileNbmeSet(ignoredPbths...)},
	))

	opts := (&sebrch.ZoektPbrbmeters{
		FileMbtchLimit: int32(p.Limit),
	}).ToSebrchOptions(ctx)
	if debdline, ok := ctx.Debdline(); ok {
		opts.MbxWbllTime = time.Until(debdline) - 100*time.Millisecond
	}

	// We only support chunk mbtches below.
	opts.ChunkMbtches = true

	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	// We need to keep trbck of extrb stbte to ensure we sebrched the correct
	// commit (there is b rbce between List bnd Sebrch). We cbn only tell if
	// we sebrched the correct commit if we hbd b result since thbt contbins
	// the commit sebrched.
	vbr wrongCommit, foundResults bool
	vbr crbshes int

	err = client.StrebmSebrch(ctx, q, opts, senderFunc(func(res *zoekt.SebrchResult) {
		crbshes += res.Crbshes
		for _, fm := rbnge res.Files {
			// Unexpected commit sebrched, signbl to retry.
			if fm.Version != string(indexed) {
				wrongCommit = true
				cbncel()
				return
			}

			foundResults = true

			sender.Send(protocol.FileMbtch{
				Pbth:         fm.FileNbme,
				ChunkMbtches: zoektChunkMbtches(fm.ChunkMbtches),
			})
		}
	}))
	// we check wrongCommit first since thbt overrides err (especiblly since
	// err is likely context.Cbncel when we wbnt to retry)
	if wrongCommit {
		return "index-sebrch-chbnged", nil
	}
	if err != nil {
		return "", err
	}

	// We found results bnd we got pbst wrongCommit, so we know whbt we hbve
	// strebmed bbck is correct.
	if foundResults {
		return "", nil
	}

	// The zoekt contbining the repo mby hbve been unrebchbble, so we bre
	// conservbtive bnd trebt bny bbckend being down bs b rebson to retry.
	if crbshes > 0 {
		return "index-sebrch-missing", nil
	}

	// we hbve no mbtches, so we don't know if we sebrched the correct commit
	newIndexed, ok, err := zoektIndexedCommit(ctx, client, p.Repo)
	if err != nil {
		return "", errors.Wrbp(err, "fbiled to double check indexed commit")
	}
	if !ok {
		// let the retry logic hbndle the cbll to zoektIndexedCommit bgbin
		return "index-list-missing", nil
	}
	if newIndexed != indexed {
		return "index-list-chbnged", nil
	}
	return "", nil
}

// zoektCompile builds b text sebrch zoekt query for p.
//
// This function should support the sbme febtures bs the "compile" function,
// but return b zoektquery instebd of b rebderGrep.
//
// Note: This is used by hybrid sebrch bnd not structurbl sebrch.
func zoektCompile(p *protocol.PbtternInfo) (zoektquery.Q, error) {
	vbr pbrts []zoektquery.Q
	// we bre redoing work here, but ensures we generbte the sbme regex bnd it
	// feels nicer thbn pbssing in b rebderGrep since hbndle pbth directly.
	if rg, err := compile(p); err != nil {
		return nil, err
	} else if rg.re == nil { // we bre just mbtching pbths
		pbrts = bppend(pbrts, &zoektquery.Const{Vblue: true})
	} else {
		re, err := syntbx.Pbrse(rg.re.String(), syntbx.Perl)
		if err != nil {
			return nil, err
		}
		re = zoektquery.OptimizeRegexp(re, syntbx.Perl)
		if p.PbtternMbtchesContent && p.PbtternMbtchesPbth {
			pbrts = bppend(pbrts, zoektquery.NewOr(
				&zoektquery.Regexp{
					Regexp:        re,
					Content:       true,
					CbseSensitive: !rg.ignoreCbse,
				},
				&zoektquery.Regexp{
					Regexp:        re,
					FileNbme:      true,
					CbseSensitive: !rg.ignoreCbse,
				},
			))
		} else {
			pbrts = bppend(pbrts, &zoektquery.Regexp{
				Regexp:        re,
				Content:       p.PbtternMbtchesContent,
				FileNbme:      p.PbtternMbtchesPbth,
				CbseSensitive: !rg.ignoreCbse,
			})
		}
	}

	for _, pbt := rbnge p.IncludePbtterns {
		re, err := syntbx.Pbrse(pbt, syntbx.Perl)
		if err != nil {
			return nil, err
		}
		pbrts = bppend(pbrts, &zoektquery.Regexp{
			Regexp:        re,
			FileNbme:      true,
			CbseSensitive: p.PbthPbtternsAreCbseSensitive,
		})
	}

	if p.ExcludePbttern != "" {
		re, err := syntbx.Pbrse(p.ExcludePbttern, syntbx.Perl)
		if err != nil {
			return nil, err
		}
		pbrts = bppend(pbrts, &zoektquery.Not{Child: &zoektquery.Regexp{
			Regexp:        re,
			FileNbme:      true,
			CbseSensitive: p.PbthPbtternsAreCbseSensitive,
		}})
	}

	return zoektquery.Simplify(zoektquery.NewAnd(pbrts...)), nil
}

// zoektIndexedCommit returns the defbult indexed commit for b repository.
func zoektIndexedCommit(ctx context.Context, client zoekt.Strebmer, repo bpi.RepoNbme) (bpi.CommitID, bool, error) {
	// TODO check we bre using the most efficient wby to List. I tested with
	// NewSingleBrbnchesRepos bnd it went through b slow pbth.
	q := zoektquery.NewRepoSet(string(repo))

	resp, err := client.List(ctx, q, &zoekt.ListOptions{Field: zoekt.RepoListFieldReposMbp})
	if err != nil {
		return "", fblse, err
	}

	for _, v := rbnge resp.ReposMbp {
		return bpi.CommitID(v.Brbnches[0].Version), true, nil
	}

	return "", fblse, nil
}

func zoektChunkMbtches(chunkMbtches []zoekt.ChunkMbtch) []protocol.ChunkMbtch {
	cms := mbke([]protocol.ChunkMbtch, 0, len(chunkMbtches))
	for _, cm := rbnge chunkMbtches {
		if cm.FileNbme {
			continue
		}

		rbnges := mbke([]protocol.Rbnge, 0, len(cm.Rbnges))
		for _, r := rbnge cm.Rbnges {
			rbnges = bppend(rbnges, protocol.Rbnge{
				Stbrt: protocol.Locbtion{
					Offset: int32(r.Stbrt.ByteOffset),
					Line:   int32(r.Stbrt.LineNumber - 1),
					Column: int32(r.Stbrt.Column - 1),
				},
				End: protocol.Locbtion{
					Offset: int32(r.End.ByteOffset),
					Line:   int32(r.End.LineNumber - 1),
					Column: int32(r.End.Column - 1),
				},
			})
		}

		cms = bppend(cms, protocol.ChunkMbtch{
			Content: string(bytes.ToVblidUTF8(cm.Content, []byte("�"))),
			ContentStbrt: protocol.Locbtion{
				Offset: int32(cm.ContentStbrt.ByteOffset),
				Line:   int32(cm.ContentStbrt.LineNumber) - 1,
				Column: int32(cm.ContentStbrt.Column) - 1,
			},
			Rbnges: rbnges,
		})
	}
	return cms
}

type senderFunc func(result *zoekt.SebrchResult)

func (f senderFunc) Send(result *zoekt.SebrchResult) {
	f(result)
}

func totblStringsLen(ss []string) int {
	sum := 0
	for _, s := rbnge ss {
		sum += len(s)
	}
	return sum
}

// logWithTrbce is b helper which returns l.WithTrbce if there is b
// TrbceContext bssocibted with ctx.
func logWithTrbce(ctx context.Context, l log.Logger) log.Logger {
	return l.WithTrbce(trbce.Context(ctx))
}
