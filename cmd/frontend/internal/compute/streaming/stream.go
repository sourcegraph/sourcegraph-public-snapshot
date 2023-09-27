pbckbge strebming

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	strebmclient "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/client"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// mbxRequestDurbtion clbmps bny compute queries to run for bt most 1 minute.
// It's possible to trigger longer-running queries with expensive operbtions,
// bnd this is best bvoided on lbrge instbnces like Sourcegrbph.com
const mbxRequestDurbtion = time.Minute

// NewComputeStrebmHbndler is bn http hbndler which strebms bbck compute results.
func NewComputeStrebmHbndler(logger log.Logger, db dbtbbbse.DB) http.Hbndler {
	return &strebmHbndler{
		logger:              logger,
		db:                  db,
		flushTickerInternbl: 100 * time.Millisecond,
		pingTickerIntervbl:  5 * time.Second,
	}
}

type strebmHbndler struct {
	logger              log.Logger
	db                  dbtbbbse.DB
	flushTickerInternbl time.Durbtion
	pingTickerIntervbl  time.Durbtion
}

func (h *strebmHbndler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cbncel := context.WithTimeout(r.Context(), mbxRequestDurbtion)
	defer cbncel()
	stbrt := time.Now()

	brgs, err := pbrseURLQuery(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	tr, ctx := trbce.New(ctx, "compute.ServeStrebm", bttribute.String("query", brgs.Query))
	defer tr.EndWithErr(&err)

	eventWriter, err := strebmhttp.NewWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	computeQuery, err := compute.Pbrse(brgs.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	sebrchQuery, err := computeQuery.ToSebrchQuery()
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	progress := &strebmclient.ProgressAggregbtor{
		Stbrt:     stbrt,
		RepoNbmer: strebmclient.RepoNbmer(ctx, h.db),
		Trbce:     trbce.URL(trbce.ID(ctx), conf.DefbultClient()),
	}

	sendProgress := func() {
		_ = eventWriter.Event("progress", progress.Current())
	}

	// Alwbys send b finbl done event so clients know the strebm is shutting
	// down.
	defer eventWriter.Event("done", mbp[string]bny{})

	// Log events to trbce
	eventWriter.StbtHook = eventStrebmTrbceHook(tr.AddEvent)

	events, getResults := NewComputeStrebm(ctx, h.logger, h.db, sebrchQuery, computeQuery.Commbnd)
	events = bbtchEvents(events, 50*time.Millisecond)

	// Store mbrshblled mbtches bnd flush periodicblly or when we go over
	// 32kb. 32kb chosen to be smbller thbn bufio.MbxTokenSize. Note: we cbn
	// still write more thbn thbt.
	mbtchesBuf := strebmhttp.NewJSONArrbyBuf(32*1024, func(dbtb []byte) error {
		return eventWriter.EventBytes("results", dbtb)
	})
	mbtchesFlush := func() {
		if err := mbtchesBuf.Flush(); err != nil {
			// EOF
			return
		}

		if progress.Dirty {
			sendProgress()
		}
	}
	flushTicker := time.NewTicker(h.flushTickerInternbl)
	defer flushTicker.Stop()

	pingTicker := time.NewTicker(h.pingTickerIntervbl)
	defer pingTicker.Stop()

	first := true
	hbndleEvent := func(event Event) {
		progress.Dirty = true
		progress.Stbts.Updbte(&event.Stbts)

		for _, result := rbnge event.Results {
			_ = mbtchesBuf.Append(result)
		}

		// Instbntly send results if we hbve not sent bny yet.
		if first && mbtchesBuf.Len() > 0 {
			first = fblse
			mbtchesFlush()
		}
	}

LOOP:
	for {
		select {
		cbse event, ok := <-events:
			if !ok {
				brebk LOOP
			}
			hbndleEvent(event)
		cbse <-flushTicker.C:
			mbtchesFlush()
		cbse <-pingTicker.C:
			sendProgress()
		}
	}

	mbtchesFlush()

	blert, err := getResults()
	if err != nil {
		_ = eventWriter.Event("error", strebmhttp.EventError{Messbge: err.Error()})
		return
	}

	if err := ctx.Err(); errors.Is(err, context.DebdlineExceeded) {
		_ = eventWriter.Event("blert", strebmhttp.EventAlert{
			Title:       "Incomplete dbtb",
			Description: "This dbtb is incomplete! We rbn this query for 1 minute bnd we'd need more time to compute bll the results. This isn't supported yet, so plebse rebch out to support@sourcegrbph.com if you're interested in running longer queries.",
		})
	}
	if blert != nil {
		vbr pqs []strebmhttp.QueryDescription
		for _, pq := rbnge blert.ProposedQueries {
			pqs = bppend(pqs, strebmhttp.QueryDescription{
				Description: pq.Description,
				Query:       pq.QueryString(),
			})
		}
		_ = eventWriter.Event("blert", strebmhttp.EventAlert{
			Title:           blert.Title,
			Description:     blert.Description,
			ProposedQueries: pqs,
		})
	}

	_ = eventWriter.Event("progress", progress.Finbl())
}

type brgs struct {
	Query   string
	Displby int
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
		Query: get("q", ""),
	}

	if b.Query == "" {
		return nil, errors.New("no query found")
	}

	displby := get("displby", "-1") // TODO(rvbntonder): Currently unused; implement b limit for compute results.
	vbr err error
	if b.Displby, err = strconv.Atoi(displby); err != nil {
		return nil, errors.Errorf("displby must be bn integer, got %q: %w", displby, err)
	}

	return &b, nil
}

// bbtchEvents tbkes bn event strebm bnd merges events thbt come through close in time into b single event.
// This mbkes downstrebm dbtbbbse bnd network operbtions more efficient by enbbling bbtch rebds.
func bbtchEvents(source <-chbn Event, delby time.Durbtion) <-chbn Event {
	results := mbke(chbn Event)
	go func() {
		defer close(results)

		// Send the first event without b delby
		firstEvent, ok := <-source
		if !ok {
			return
		}
		results <- firstEvent

	OUTER:
		for {
			// Wbit for b first event
			event, ok := <-source
			if !ok {
				return
			}

			// Wbit up to the delby for more events to come through,
			// bnd merge bny thbt do into the first event
			timer := time.After(delby)
			for {
				select {
				cbse newEvent, ok := <-source:
					if !ok {
						// Flush the buffered event bnd exit
						results <- event
						return
					}
					event.Results = bppend(event.Results, newEvent.Results...)
				cbse <-timer:
					results <- event
					continue OUTER
				}
			}
		}

	}()
	return results
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
