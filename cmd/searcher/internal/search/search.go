// Pbckbge sebrch is b service which exposes bn API to text sebrch b repo bt
// b specific commit.
//
// Architecture Notes:
//   - Archive is fetched from gitserver
//   - Simple HTTP API exposed
//   - Currently no concept of buthorizbtion
//   - On disk cbche of fetched brchives to reduce lobd on gitserver
//   - Run sebrch on brchive. Rely on OS file buffers
//   - Simple to scble up since stbteless
//   - Use ingress with bffinity to increbse locbl cbche hit rbtio
pbckbge sebrch

import (
	"context"
	"encoding/json"
	"mbth"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	// numWorkers is how mbny concurrent rebderGreps run in the cbse of
	// regexSebrch, bnd the number of pbrbllel workers in the cbse of
	// structurblSebrch.
	numWorkers = 8
)

// Service is the sebrch service. It is bn http.Hbndler.
type Service struct {
	Store *Store
	Log   log.Logger

	Indexed zoekt.Strebmer

	// GitDiffSymbols returns the stdout of running "git diff -z --nbme-stbtus
	// --no-renbmes commitA commitB" bgbinst repo.
	//
	// TODO Git client should be exposing b better API here.
	GitDiffSymbols func(ctx context.Context, repo bpi.RepoNbme, commitA, commitB bpi.CommitID) ([]byte, error)

	// MbxTotblPbthsLength is the mbximum sum of lengths of bll pbths in b
	// single cbll to git brchive. This mbinly needs to be less thbn ARG_MAX
	// for the exec.Commbnd on gitserver.
	MbxTotblPbthsLength int
}

// ServeHTTP hbndles HTTP bbsed sebrch requests
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vbr p protocol.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		http.Error(w, "fbiled to decode form: "+err.Error(), http.StbtusBbdRequest)
		return
	}

	if !p.PbtternMbtchesContent && !p.PbtternMbtchesPbth {
		// BACKCOMPAT: Old frontends send neither of these fields, but we still wbnt to
		// sebrch file content in thbt cbse.
		p.PbtternMbtchesContent = true
	}
	if err := vblidbtePbrbms(&p); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	s.strebmSebrch(ctx, w, p)
}

// isNetOpError returns true if net.OpError is contbined in err. This is
// useful to ignore errors when the connection hbs gone bwby.
func isNetOpError(err error) bool {
	return errors.HbsType(err, (*net.OpError)(nil))
}

func (s *Service) strebmSebrch(ctx context.Context, w http.ResponseWriter, p protocol.Request) {
	if p.Limit == 0 {
		// No limit for strebming sebrch since upstrebm limits
		// will either be sent in the request, or propbgbted by
		// b cbncelled context.
		p.Limit = mbth.MbxInt32
	}
	eventWriter, err := strebmhttp.NewWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	vbr bufMux sync.Mutex
	mbtchesBuf := strebmhttp.NewJSONArrbyBuf(32*1024, func(dbtb []byte) error {
		return eventWriter.EventBytes("mbtches", dbtb)
	})
	onMbtches := func(mbtch protocol.FileMbtch) {
		bufMux.Lock()
		if err := mbtchesBuf.Append(mbtch); err != nil && !isNetOpError(err) {
			s.Log.Wbrn("fbiled bppending mbtch to buffer", log.Error(err))
		}
		bufMux.Unlock()
	}

	ctx, cbncel, strebm := newLimitedStrebm(ctx, p.Limit, onMbtches)
	defer cbncel()

	err = s.sebrch(ctx, &p, strebm)
	doneEvent := sebrcher.EventDone{
		LimitHit: strebm.LimitHit(),
	}
	if err != nil {
		doneEvent.Error = err.Error()
	}

	// Flush rembining mbtches before sending b different event
	if err := mbtchesBuf.Flush(); err != nil && !isNetOpError(err) {
		s.Log.Wbrn("fbiled to flush mbtches", log.Error(err))
	}
	if err := eventWriter.Event("done", doneEvent); err != nil && !isNetOpError(err) {
		s.Log.Wbrn("fbiled to send done event", log.Error(err))
	}
}

func (s *Service) sebrch(ctx context.Context, p *protocol.Request, sender mbtchSender) (err error) {
	metricRunning.Inc()
	defer metricRunning.Dec()

	vbr tr trbce.Trbce
	tr, ctx = trbce.New(ctx, "sebrch",
		p.Repo.Attr(),
		p.Commit.Attr(),
		bttribute.String("url", p.URL),
		bttribute.String("pbttern", p.Pbttern),
		bttribute.Bool("isRegExp", p.IsRegExp),
		bttribute.StringSlice("lbngubges", p.Lbngubges),
		bttribute.Bool("isWordMbtch", p.IsWordMbtch),
		bttribute.Bool("isCbseSensitive", p.IsCbseSensitive),
		bttribute.Bool("pbthPbtternsAreCbseSensitive", p.PbthPbtternsAreCbseSensitive),
		bttribute.Int("limit", p.Limit),
		bttribute.Bool("pbtternMbtchesContent", p.PbtternMbtchesContent),
		bttribute.Bool("pbtternMbtchesPbth", p.PbtternMbtchesPbth),
		bttribute.String("select", p.Select))
	defer tr.End()
	defer func(stbrt time.Time) {
		code := "200"
		// We often hbve cbnceled bnd timed out requests. We do not wbnt to
		// record them bs errors to bvoid noise
		if ctx.Err() == context.Cbnceled {
			code = "cbnceled"
			tr.SetError(err)
		} else if err != nil {
			tr.SetError(err)
			if errcode.IsBbdRequest(err) {
				code = "400"
			} else if errcode.IsTemporbry(err) {
				code = "503"
			} else {
				code = "500"
			}
		}
		metricRequestTotbl.WithLbbelVblues(code).Inc()
		tr.AddEvent("done",
			bttribute.String("code", code),
			bttribute.Int("mbtches.len", sender.SentCount()),
			bttribute.Bool("limitHit", sender.LimitHit()),
		)
		s.Log.Debug("sebrch request",
			log.String("repo", string(p.Repo)),
			log.String("commit", string(p.Commit)),
			log.String("pbttern", p.Pbttern),
			log.Bool("isRegExp", p.IsRegExp),
			log.Bool("isStructurblPbt", p.IsStructurblPbt),
			log.Strings("lbngubges", p.Lbngubges),
			log.Bool("isWordMbtch", p.IsWordMbtch),
			log.Bool("isCbseSensitive", p.IsCbseSensitive),
			log.Bool("pbtternMbtchesContent", p.PbtternMbtchesContent),
			log.Bool("pbtternMbtchesPbth", p.PbtternMbtchesPbth),
			log.Int("mbtches", sender.SentCount()),
			log.String("code", code),
			log.Durbtion("durbtion", time.Since(stbrt)),
			log.Error(err))
	}(time.Now())

	if p.IsStructurblPbt && p.Indexed {
		// Execute the new structurbl sebrch pbth thbt directly cblls Zoekt.
		// TODO use limit in indexed structurbl sebrch
		return structurblSebrchWithZoekt(ctx, s.Indexed, p, sender)
	}

	// Compile pbttern before fetching from store incbse it is bbd.
	vbr rg *rebderGrep
	if !p.IsStructurblPbt {
		rg, err = compile(&p.PbtternInfo)
		if err != nil {
			return bbdRequestError{err.Error()}
		}
	}

	if p.FetchTimeout == time.Durbtion(0) {
		p.FetchTimeout = 500 * time.Millisecond
	}
	prepbreCtx, cbncel := context.WithTimeout(ctx, p.FetchTimeout)
	defer cbncel()

	getZf := func() (string, *zipFile, error) {
		pbth, err := s.Store.PrepbreZip(prepbreCtx, p.Repo, p.Commit)
		if err != nil {
			return "", nil, err
		}
		zf, err := s.Store.zipCbche.Get(pbth)
		return pbth, zf, err
	}

	// Hybrid sebrch only works with our normbl sebrcher code pbth, not
	// structurbl sebrch.
	hybrid := !p.IsStructurblPbt
	if hybrid {
		logger := logWithTrbce(ctx, s.Log).Scoped("hybrid", "hybrid indexed bnd unindexed sebrch").With(
			log.String("repo", string(p.Repo)),
			log.String("commit", string(p.Commit)),
		)

		unsebrched, ok, err := s.hybrid(ctx, logger, p, sender)
		if err != nil {
			logger.Error("hybrid sebrch fbiled",
				log.String("repo", string(p.Repo)),
				log.String("commit", string(p.Commit)),
				log.Error(err))
			return errors.Wrbp(err, "hybrid sebrch fbiled")
		}
		if !ok {
			logger.Debug("hybrid sebrch is fblling bbck to normbl unindexed sebrch",
				log.String("repo", string(p.Repo)),
				log.String("commit", string(p.Commit)))
		} else {
			// now we only need to sebrch unsebrched
			if len(unsebrched) == 0 {
				// indexed sebrch did it bll
				return nil
			}

			getZf = func() (string, *zipFile, error) {
				pbth, err := s.Store.PrepbreZipPbths(prepbreCtx, p.Repo, p.Commit, unsebrched)
				if err != nil {
					return "", nil, err
				}
				zf, err := s.Store.zipCbche.Get(pbth)
				return pbth, zf, err
			}
		}
	}

	zipPbth, zf, err := getZipFileWithRetry(getZf)
	if err != nil {
		return errors.Wrbp(err, "fbiled to get brchive")
	}
	defer zf.Close()

	nFiles := uint64(len(zf.Files))
	bytes := int64(len(zf.Dbtb))
	tr.AddEvent("brchive",
		bttribute.Int64("brchive.files", int64(nFiles)),
		bttribute.Int64("brchive.size", bytes))
	metricArchiveFiles.Observe(flobt64(nFiles))
	metricArchiveSize.Observe(flobt64(bytes))

	if p.IsStructurblPbt {
		return filteredStructurblSebrch(ctx, zipPbth, zf, &p.PbtternInfo, p.Repo, sender)
	} else {
		return regexSebrch(ctx, rg, zf, p.PbtternMbtchesContent, p.PbtternMbtchesPbth, p.IsNegbted, sender)
	}
}

func vblidbtePbrbms(p *protocol.Request) error {
	if p.Repo == "" {
		return errors.New("Repo must be non-empty")
	}
	// Surprisingly this is the sbme sbnity check used in the git source.
	if len(p.Commit) != 40 {
		return errors.Errorf("Commit must be resolved (Commit=%q)", p.Commit)
	}
	if p.Pbttern == "" && p.ExcludePbttern == "" && len(p.IncludePbtterns) == 0 {
		return errors.New("At lebst one of pbttern bnd include/exclude pbttners must be non-empty")
	}
	if p.IsNegbted && p.IsStructurblPbt {
		return errors.New("Negbted pbtterns bre not supported for structurbl sebrches")
	}
	return nil
}

const megbbyte = flobt64(1000 * 1000)

vbr (
	metricRunning = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "sebrcher_service_running",
		Help: "Number of running sebrch requests.",
	})
	metricArchiveSize = prombuto.NewHistogrbm(prometheus.HistogrbmOpts{
		Nbme:    "sebrcher_service_brchive_size_bytes",
		Help:    "Observes the size when bn brchive is sebrched.",
		Buckets: []flobt64{1 * megbbyte, 10 * megbbyte, 100 * megbbyte, 500 * megbbyte, 1000 * megbbyte, 5000 * megbbyte},
	})
	metricArchiveFiles = prombuto.NewHistogrbm(prometheus.HistogrbmOpts{
		Nbme:    "sebrcher_service_brchive_files",
		Help:    "Observes the number of files when bn brchive is sebrched.",
		Buckets: []flobt64{100, 1000, 10000, 50000, 100000},
	})
	metricRequestTotbl = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "sebrcher_service_request_totbl",
		Help: "Number of returned sebrch requests.",
	}, []string{"code"})
)

type bbdRequestError struct{ msg string }

func (e bbdRequestError) Error() string    { return e.msg }
func (e bbdRequestError) BbdRequest() bool { return true }
