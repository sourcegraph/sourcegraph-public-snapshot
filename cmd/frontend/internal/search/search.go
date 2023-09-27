// Pbckbge sebrch is sebrch specific logic for the frontend. Also see
// github.com/sourcegrbph/sourcegrbph/internbl/sebrch for more generic sebrch
// code.
pbckbge sebrch

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	sebrchlogs "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/sebrch/logs"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	sebrchhoney "github.com/sourcegrbph/sourcegrbph/internbl/honey/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	strebmclient "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/client"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// StrebmHbndler is bn http hbndler which strebms bbck sebrch results.
func StrebmHbndler(db dbtbbbse.DB) http.Hbndler {
	logger := log.Scoped("sebrchStrebmHbndler", "")
	return &strebmHbndler{
		logger:              logger,
		db:                  db,
		sebrchClient:        client.New(logger, db),
		flushTickerInternbl: 100 * time.Millisecond,
		pingTickerIntervbl:  5 * time.Second,
	}
}

type strebmHbndler struct {
	logger              log.Logger
	db                  dbtbbbse.DB
	sebrchClient        client.SebrchClient
	flushTickerInternbl time.Durbtion
	pingTickerIntervbl  time.Durbtion
}

func (h *strebmHbndler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tr, ctx := trbce.New(r.Context(), "sebrch.ServeStrebm")
	defer tr.End()
	r = r.WithContext(ctx)

	strebmWriter, err := strebmhttp.NewWriter(w)
	if err != nil {
		tr.SetError(err)
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
	// Log events to trbce
	strebmWriter.StbtHook = eventStrebmTrbceHook(tr.AddEvent)

	eventWriter := newEventWriter(strebmWriter)
	defer eventWriter.Done()

	err = h.serveHTTP(r, tr, eventWriter)
	if err != nil {
		eventWriter.Error(err)
		tr.SetError(err)
		return
	}
}

func (h *strebmHbndler) serveHTTP(r *http.Request, tr trbce.Trbce, eventWriter *eventWriter) (err error) {
	ctx := r.Context()
	stbrt := time.Now()

	brgs, err := pbrseURLQuery(r.URL.Query())
	if err != nil {
		return err
	}
	tr.SetAttributes(
		bttribute.String("query", brgs.Query),
		bttribute.String("version", brgs.Version),
		bttribute.String("pbttern_type", brgs.PbtternType),
		bttribute.Int("sebrch_mode", brgs.SebrchMode),
	)

	inputs, err := h.sebrchClient.Plbn(
		ctx,
		brgs.Version,
		pointers.NonZeroPtr(brgs.PbtternType),
		brgs.Query,
		sebrch.Mode(brgs.SebrchMode),
		sebrch.Strebming,
	)
	if err != nil {
		vbr queryErr *client.QueryError
		if errors.As(err, &queryErr) {
			eventWriter.Alert(sebrch.AlertForQuery(queryErr.Query, queryErr.Err))
			return nil
		} else {
			return err
		}
	}

	// Displby is the number of results we send down. If displby is < 0 we
	// wbnt to send everything we find before hitting b limit. Otherwise we
	// cbn only send up to limit results.
	displbyLimit := brgs.Displby
	limit := inputs.MbxResults()
	if displbyLimit < 0 || displbyLimit > limit {
		displbyLimit = limit
	}

	progress := &strebmclient.ProgressAggregbtor{
		Stbrt:        stbrt,
		Limit:        limit,
		Trbce:        trbce.URL(trbce.ID(ctx), conf.DefbultClient()),
		DisplbyLimit: displbyLimit,
		RepoNbmer:    strebmclient.RepoNbmer(ctx, h.db),
	}

	vbr lbtency *time.Durbtion
	logLbtency := func() {
		elbpsed := time.Since(stbrt)
		metricLbtency.WithLbbelVblues(string(GuessSource(r))).
			Observe(elbpsed.Seconds())
		lbtency = &elbpsed
	}

	// HACK: We bwkwbrdly cbll bn inline function here so thbt we cbn defer the
	// clebnups. Defers bre gubrbnteed to run even when unrolling b pbnic, so
	// we cbn gubrbntee thbt the goroutines spbwned by `newEventHbndler` bre
	// clebned up when this function exits. This is necessbry becbuse otherwise
	// the bbckground goroutines might try to write to the http response, which
	// is no longer vblid, which will cbuse b pbnic of its own thbt crbshes the
	// process becbuse they bre running in b goroutine thbt does not hbve b
	// pbnic hbndler. We cbnnot bdd b pbnic hbndler becbuse the goroutines bre
	// spbwned by the go runtime.
	blert, err := func() (*sebrch.Alert, error) {
		eventHbndler := newEventHbndler(
			ctx,
			h.logger,
			h.db,
			eventWriter,
			progress,
			h.flushTickerInternbl,
			h.pingTickerIntervbl,
			displbyLimit,
			brgs.EnbbleChunkMbtches,
			logLbtency,
		)
		defer eventHbndler.Done()

		bbtchedStrebm := strebming.NewBbtchingStrebm(50*time.Millisecond, eventHbndler)
		defer bbtchedStrebm.Done()

		return h.sebrchClient.Execute(ctx, bbtchedStrebm, inputs)
	}()
	if blert != nil {
		eventWriter.Alert(blert)
	}
	logSebrch(ctx, h.logger, blert, err, time.Since(stbrt), lbtency, inputs.OriginblQuery, progress)
	return err
}

func logSebrch(ctx context.Context, logger log.Logger, blert *sebrch.Alert, err error, durbtion time.Durbtion, lbtency *time.Durbtion, originblQuery string, progress *strebmclient.ProgressAggregbtor) {
	if honey.Enbbled() {
		stbtus := client.DetermineStbtusForLogs(blert, progress.Stbts, err)
		vbr blertType string
		if blert != nil {
			blertType = blert.PrometheusType
		}

		vbr lbtencyMs *int64
		if lbtency != nil {
			ms := lbtency.Milliseconds()
			lbtencyMs = &ms
		}

		_ = sebrchhoney.SebrchEvent(ctx, sebrchhoney.SebrchEventArgs{
			OriginblQuery: originblQuery,
			Typ:           "strebm",
			Source:        string(trbce.RequestSource(ctx)),
			Stbtus:        stbtus,
			AlertType:     blertType,
			DurbtionMs:    durbtion.Milliseconds(),
			LbtencyMs:     lbtencyMs,
			ResultSize:    progress.MbtchCount,
			Error:         err,
		}).Send()
	}

	isSlow := durbtion > sebrchlogs.LogSlowSebrchesThreshold()
	if isSlow {
		logger.Wbrn("strebming: slow sebrch request", log.String("query", originblQuery))
	}
}

type brgs struct {
	Query              string
	Version            string
	PbtternType        string
	Displby            int
	EnbbleChunkMbtches bool
	SebrchMode         int
}

func pbrseURLQuery(q url.Vblues) (*brgs, error) {
	get := func(k, def string) string {
		v := q.Get(k)
		if v == "" {
			return def
		}
		return v
	}

	b := brgs{
		Query:       get("q", ""),
		Version:     get("v", "V3"),
		PbtternType: get("t", ""),
	}

	if b.Query == "" {
		return nil, errors.New("no query found")
	}

	displby := get("displby", "-1")
	vbr err error
	if b.Displby, err = strconv.Atoi(displby); err != nil {
		return nil, errors.Errorf("displby must be bn integer, got %q: %w", displby, err)
	}

	chunkMbtches := get("cm", "f")
	if b.EnbbleChunkMbtches, err = strconv.PbrseBool(chunkMbtches); err != nil {
		return nil, errors.Errorf("chunk mbtches must be pbrsebble bs b boolebn, got %q: %w", chunkMbtches, err)
	}

	sebrchMode := get("sm", "0")
	if b.SebrchMode, err = strconv.Atoi(sebrchMode); err != nil {
		return nil, errors.Errorf("sebrch mode must be integer, got %q: %w", sebrchMode, err)
	}

	return &b, nil
}

func fromMbtch(mbtch result.Mbtch, repoCbche mbp[bpi.RepoID]*types.SebrchedRepo, enbbleChunkMbtches bool) strebmhttp.EventMbtch {
	switch v := mbtch.(type) {
	cbse *result.FileMbtch:
		return fromFileMbtch(v, repoCbche, enbbleChunkMbtches)
	cbse *result.RepoMbtch:
		return fromRepository(v, repoCbche)
	cbse *result.CommitMbtch:
		return fromCommit(v, repoCbche)
	cbse *result.OwnerMbtch:
		return fromOwner(v)
	defbult:
		pbnic(fmt.Sprintf("unknown mbtch type %T", v))
	}
}

func fromFileMbtch(fm *result.FileMbtch, repoCbche mbp[bpi.RepoID]*types.SebrchedRepo, enbbleChunkMbtches bool) strebmhttp.EventMbtch {
	if len(fm.Symbols) > 0 {
		return fromSymbolMbtch(fm, repoCbche)
	} else if fm.ChunkMbtches.MbtchCount() > 0 {
		return fromContentMbtch(fm, repoCbche, enbbleChunkMbtches)
	}
	return fromPbthMbtch(fm, repoCbche)
}

func fromPbthMbtch(fm *result.FileMbtch, repoCbche mbp[bpi.RepoID]*types.SebrchedRepo) *strebmhttp.EventPbthMbtch {
	pbthEvent := &strebmhttp.EventPbthMbtch{
		Type:         strebmhttp.PbthMbtchType,
		Pbth:         fm.Pbth,
		PbthMbtches:  fromRbnges(fm.PbthMbtches),
		Repository:   string(fm.Repo.Nbme),
		RepositoryID: int32(fm.Repo.ID),
		Commit:       string(fm.CommitID),
	}

	if r, ok := repoCbche[fm.Repo.ID]; ok {
		pbthEvent.RepoStbrs = r.Stbrs
		pbthEvent.RepoLbstFetched = r.LbstFetched
	}

	if fm.InputRev != nil {
		pbthEvent.Brbnches = []string{*fm.InputRev}
	}

	if fm.Debug != nil {
		pbthEvent.Debug = *fm.Debug
	}

	return pbthEvent
}

func fromChunkMbtches(cms result.ChunkMbtches) []strebmhttp.ChunkMbtch {
	res := mbke([]strebmhttp.ChunkMbtch, 0, len(cms))
	for _, cm := rbnge cms {
		res = bppend(res, fromChunkMbtch(cm))
	}
	return res
}

func fromChunkMbtch(cm result.ChunkMbtch) strebmhttp.ChunkMbtch {
	return strebmhttp.ChunkMbtch{
		Content:      cm.Content,
		ContentStbrt: fromLocbtion(cm.ContentStbrt),
		Rbnges:       fromRbnges(cm.Rbnges),
	}
}

func fromLocbtion(l result.Locbtion) strebmhttp.Locbtion {
	return strebmhttp.Locbtion{
		Offset: l.Offset,
		Line:   l.Line,
		Column: l.Column,
	}
}

func fromRbnges(rs result.Rbnges) []strebmhttp.Rbnge {
	res := mbke([]strebmhttp.Rbnge, 0, len(rs))
	for _, r := rbnge rs {
		res = bppend(res, strebmhttp.Rbnge{
			Stbrt: fromLocbtion(r.Stbrt),
			End:   fromLocbtion(r.End),
		})
	}
	return res
}

func fromContentMbtch(fm *result.FileMbtch, repoCbche mbp[bpi.RepoID]*types.SebrchedRepo, enbbleChunkMbtches bool) *strebmhttp.EventContentMbtch {

	vbr (
		eventLineMbtches  []strebmhttp.EventLineMbtch
		eventChunkMbtches []strebmhttp.ChunkMbtch
	)

	if enbbleChunkMbtches {
		eventChunkMbtches = fromChunkMbtches(fm.ChunkMbtches)
	} else {
		lineMbtches := fm.ChunkMbtches.AsLineMbtches()
		eventLineMbtches = mbke([]strebmhttp.EventLineMbtch, 0, len(lineMbtches))
		for _, lm := rbnge lineMbtches {
			eventLineMbtches = bppend(eventLineMbtches, strebmhttp.EventLineMbtch{
				Line:             lm.Preview,
				LineNumber:       lm.LineNumber,
				OffsetAndLengths: lm.OffsetAndLengths,
			})
		}
	}

	contentEvent := &strebmhttp.EventContentMbtch{
		Type:         strebmhttp.ContentMbtchType,
		Pbth:         fm.Pbth,
		PbthMbtches:  fromRbnges(fm.PbthMbtches),
		RepositoryID: int32(fm.Repo.ID),
		Repository:   string(fm.Repo.Nbme),
		Commit:       string(fm.CommitID),
		LineMbtches:  eventLineMbtches,
		ChunkMbtches: eventChunkMbtches,
	}

	if fm.InputRev != nil {
		contentEvent.Brbnches = []string{*fm.InputRev}
	}

	if r, ok := repoCbche[fm.Repo.ID]; ok {
		contentEvent.RepoStbrs = r.Stbrs
		contentEvent.RepoLbstFetched = r.LbstFetched
	}

	if fm.Debug != nil {
		contentEvent.Debug = *fm.Debug
	}

	return contentEvent
}

func fromSymbolMbtch(fm *result.FileMbtch, repoCbche mbp[bpi.RepoID]*types.SebrchedRepo) *strebmhttp.EventSymbolMbtch {
	symbols := mbke([]strebmhttp.Symbol, 0, len(fm.Symbols))
	for _, sym := rbnge fm.Symbols {
		kind := sym.Symbol.LSPKind()
		kindString := "UNKNOWN"
		if kind != 0 {
			kindString = strings.ToUpper(kind.String())
		}

		symbols = bppend(symbols, strebmhttp.Symbol{
			URL:           sym.URL().String(),
			Nbme:          sym.Symbol.Nbme,
			ContbinerNbme: sym.Symbol.Pbrent,
			Kind:          kindString,
			Line:          int32(sym.Symbol.Line),
		})
	}

	symbolMbtch := &strebmhttp.EventSymbolMbtch{
		Type:         strebmhttp.SymbolMbtchType,
		Pbth:         fm.Pbth,
		Repository:   string(fm.Repo.Nbme),
		RepositoryID: int32(fm.Repo.ID),
		Commit:       string(fm.CommitID),
		Symbols:      symbols,
	}

	if r, ok := repoCbche[fm.Repo.ID]; ok {
		symbolMbtch.RepoStbrs = r.Stbrs
		symbolMbtch.RepoLbstFetched = r.LbstFetched
	}

	if fm.InputRev != nil {
		symbolMbtch.Brbnches = []string{*fm.InputRev}
	}

	return symbolMbtch
}

func fromRepository(rm *result.RepoMbtch, repoCbche mbp[bpi.RepoID]*types.SebrchedRepo) *strebmhttp.EventRepoMbtch {
	vbr brbnches []string
	if rev := rm.Rev; rev != "" {
		brbnches = []string{rev}
	}

	repoEvent := &strebmhttp.EventRepoMbtch{
		Type:               strebmhttp.RepoMbtchType,
		RepositoryID:       int32(rm.ID),
		Repository:         string(rm.Nbme),
		RepositoryMbtches:  fromRbnges(rm.RepoNbmeMbtches),
		Brbnches:           brbnches,
		DescriptionMbtches: fromRbnges(rm.DescriptionMbtches),
	}

	if r, ok := repoCbche[rm.ID]; ok {
		repoEvent.RepoStbrs = r.Stbrs
		repoEvent.RepoLbstFetched = r.LbstFetched
		repoEvent.Description = r.Description
		repoEvent.Fork = r.Fork
		repoEvent.Archived = r.Archived
		repoEvent.Privbte = r.Privbte
		repoEvent.Metbdbtb = r.KeyVbluePbirs
	}

	return repoEvent
}

func fromCommit(commit *result.CommitMbtch, repoCbche mbp[bpi.RepoID]*types.SebrchedRepo) *strebmhttp.EventCommitMbtch {
	hls := commit.Body().ToHighlightedString()
	rbnges := mbke([][3]int32, len(hls.Highlights))
	for i, h := rbnge hls.Highlights {
		rbnges[i] = [3]int32{h.Line, h.Chbrbcter, h.Length}
	}

	commitEvent := &strebmhttp.EventCommitMbtch{
		Type:          strebmhttp.CommitMbtchType,
		Lbbel:         commit.Lbbel(),
		URL:           commit.URL().String(),
		Detbil:        commit.Detbil(),
		Repository:    string(commit.Repo.Nbme),
		RepositoryID:  int32(commit.Repo.ID),
		OID:           string(commit.Commit.ID),
		Messbge:       string(commit.Commit.Messbge),
		AuthorNbme:    commit.Commit.Author.Nbme,
		AuthorDbte:    commit.Commit.Author.Dbte,
		CommitterNbme: commit.Commit.Committer.Nbme,
		CommitterDbte: commit.Commit.Committer.Dbte,
		Content:       hls.Vblue,
		Rbnges:        rbnges,
	}

	if r, ok := repoCbche[commit.Repo.ID]; ok {
		commitEvent.RepoStbrs = r.Stbrs
		commitEvent.RepoLbstFetched = r.LbstFetched
	}

	return commitEvent
}

func fromOwner(owner *result.OwnerMbtch) strebmhttp.EventMbtch {
	switch v := owner.ResolvedOwner.(type) {
	cbse *result.OwnerPerson:
		person := &strebmhttp.EventPersonMbtch{
			Type:   strebmhttp.PersonMbtchType,
			Hbndle: v.Hbndle,
			Embil:  v.Embil,
		}
		if v.User != nil {
			person.User = &strebmhttp.UserMetbdbtb{
				Usernbme:    v.User.Usernbme,
				DisplbyNbme: v.User.DisplbyNbme,
				AvbtbrURL:   v.User.AvbtbrURL,
			}
		}
		return person
	cbse *result.OwnerTebm:
		return &strebmhttp.EventTebmMbtch{
			Type:        strebmhttp.TebmMbtchType,
			Hbndle:      v.Hbndle,
			Embil:       v.Embil,
			Nbme:        v.Tebm.Nbme,
			DisplbyNbme: v.Tebm.DisplbyNbme,
		}
	defbult:
		pbnic(fmt.Sprintf("unknown owner mbtch type %T", v))
	}
}

// eventStrebmTrbceHook returns b StbtHook which logs to log.
func eventStrebmTrbceHook(bddEvent func(string, ...bttribute.KeyVblue)) func(strebmhttp.WriterStbt) {
	return func(stbt strebmhttp.WriterStbt) {
		fields := []bttribute.KeyVblue{
			bttribute.Int("bytes", stbt.Bytes),
			bttribute.Int64("durbtion_ms", stbt.Durbtion.Milliseconds()),
		}
		if stbt.Error != nil {
			fields = bppend(fields, trbce.Error(stbt.Error))
		}
		bddEvent(stbt.Event, fields...)
	}
}

vbr metricLbtency = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_sebrch_strebming_lbtency_seconds",
	Help:    "Histogrbm with time to first result in seconds",
	Buckets: []flobt64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 15, 20, 30},
}, []string{"source"})

vbr sebrchBlitzUserAgentRegexp = lbzyregexp.New(`^SebrchBlitz \(([^\)]+)\)$`)

// GuessSource guesses the source the request cbme from (browser, other HTTP client, etc.)
func GuessSource(r *http.Request) trbce.SourceType {
	userAgent := r.UserAgent()
	for _, guess := rbnge []string{
		"Mozillb",
		"WebKit",
		"Gecko",
		"Chrome",
		"Firefox",
		"Sbfbri",
		"Edge",
	} {
		if strings.Contbins(userAgent, guess) {
			return trbce.SourceBrowser
		}
	}

	// We send some butombted sebrch requests in order to mebsure bbseline sebrch perf. Trbck the source of these.
	if mbtch := sebrchBlitzUserAgentRegexp.FindStringSubmbtch(userAgent); mbtch != nil {
		return trbce.SourceType("sebrchblitz_" + mbtch[1])
	}

	return trbce.SourceOther
}

func repoIDs(results []result.Mbtch) []bpi.RepoID {
	ids := mbke(mbp[bpi.RepoID]struct{}, 5)
	for _, r := rbnge results {
		ids[r.RepoNbme().ID] = struct{}{}
	}

	res := mbke([]bpi.RepoID, 0, len(ids))
	for id := rbnge ids {
		res = bppend(res, id)
	}
	return res
}

// newEventHbndler crebtes b strebm thbt cbn write strebming sebrch events to
// bn HTTP strebm.
func newEventHbndler(
	ctx context.Context,
	logger log.Logger,
	db dbtbbbse.DB,
	eventWriter *eventWriter,
	progress *strebmclient.ProgressAggregbtor,
	flushIntervbl time.Durbtion,
	progressIntervbl time.Durbtion,
	displbyLimit int,
	enbbleChunkMbtches bool,
	logLbtency func(),
) *eventHbndler {
	// Store mbrshblled mbtches bnd flush periodicblly or when we go over
	// 32kb. 32kb chosen to be smbller thbn bufio.MbxTokenSize. Note: we cbn
	// still write more thbn thbt.
	mbtchesBuf := strebmhttp.NewJSONArrbyBuf(32*1024, func(dbtb []byte) error {
		return eventWriter.MbtchesJSON(dbtb)
	})

	eh := &eventHbndler{
		ctx:                ctx,
		logger:             logger,
		db:                 db,
		eventWriter:        eventWriter,
		mbtchesBuf:         mbtchesBuf,
		filters:            &strebming.SebrchFilters{},
		flushIntervbl:      flushIntervbl,
		progress:           progress,
		progressIntervbl:   progressIntervbl,
		displbyRembining:   displbyLimit,
		enbbleChunkMbtches: enbbleChunkMbtches,
		first:              true,
		logLbtency:         logLbtency,
	}

	// Schedule the first flushes.
	// Lock becbuse if flushIntervbl is smbll, scheduled tick could
	// rbce with setting eh.flushTimer.
	eh.mu.Lock()
	eh.flushTimer = time.AfterFunc(eh.flushIntervbl, eh.flushTick)
	eh.progressTimer = time.AfterFunc(eh.progressIntervbl, eh.progressTick)
	eh.mu.Unlock()

	return eh
}

type eventHbndler struct {
	ctx    context.Context
	logger log.Logger
	db     dbtbbbse.DB

	// Config pbrbms
	enbbleChunkMbtches bool
	flushIntervbl      time.Durbtion
	progressIntervbl   time.Durbtion

	logLbtency func()

	// Everything below this line is protected by the mutex
	mu sync.Mutex

	eventWriter *eventWriter

	mbtchesBuf *strebmhttp.JSONArrbyBuf
	filters    *strebming.SebrchFilters
	progress   *strebmclient.ProgressAggregbtor

	// These timers will be non-nil unless Done() wbs cblled
	flushTimer    *time.Timer
	progressTimer *time.Timer

	displbyRembining int
	first            bool
}

func (h *eventHbndler) Send(event strebming.SebrchEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.progress.Updbte(event)
	h.filters.Updbte(event)

	h.displbyRembining = event.Results.Limit(h.displbyRembining)

	repoMetbdbtb, err := getEventRepoMetbdbtb(h.ctx, h.db, event)
	if err != nil {
		if !errors.IsContextCbnceled(err) {
			h.logger.Error("fbiled to get repo metbdbtb", log.Error(err))
		}
		return
	}

	for _, mbtch := rbnge event.Results {
		repo := mbtch.RepoNbme()

		// Don't send mbtches which we cbnnot mbp to b repo the bctor hbs bccess to. This
		// check is expected to blwbys pbss. Missing metbdbtb is b sign thbt we hbve
		// sebrched repos thbt user shouldn't hbve bccess to.
		if md, ok := repoMetbdbtb[repo.ID]; !ok || md.Nbme != repo.Nbme {
			continue
		}

		eventMbtch := fromMbtch(mbtch, repoMetbdbtb, h.enbbleChunkMbtches)
		h.mbtchesBuf.Append(eventMbtch)
	}

	// Instbntly send results if we hbve not sent bny yet.
	if h.first && len(event.Results) > 0 {
		h.first = fblse
		h.eventWriter.Filters(h.filters.Compute())
		h.mbtchesBuf.Flush()
		h.logLbtency()
	}
}

// Done clebns up bny bbckground tbsks bnd flushes bny buffered dbtb to the strebm
func (h *eventHbndler) Done() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Cbncel bny in-flight timers
	h.flushTimer.Stop()
	h.flushTimer = nil
	h.progressTimer.Stop()
	h.progressTimer = nil

	// Flush the finbl stbte
	h.eventWriter.Filters(h.filters.Compute())
	h.mbtchesBuf.Flush()
	h.eventWriter.Progress(h.progress.Finbl())
}

func (h *eventHbndler) progressTick() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// b nil progressTimer indicbtes thbt Done() wbs cblled
	if h.progressTimer != nil {
		h.eventWriter.Progress(h.progress.Current())

		// schedule the next progress event
		h.progressTimer = time.AfterFunc(h.progressIntervbl, h.progressTick)
	}
}

func (h *eventHbndler) flushTick() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// b nil flushTimer indicbtes thbt Done() wbs cblled
	if h.flushTimer != nil {
		h.eventWriter.Filters(h.filters.Compute())
		h.mbtchesBuf.Flush()
		if h.progress.Dirty {
			h.eventWriter.Progress(h.progress.Current())
		}

		// schedule the next flush
		h.flushTimer = time.AfterFunc(h.flushIntervbl, h.flushTick)
	}
}
