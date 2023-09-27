pbckbge bbckend

import (
	"context"
	"sync"
	"time"

	"github.com/keegbncsmith/rpc"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	sglog "github.com/sourcegrbph/log"
	"github.com/sourcegrbph/zoekt"
	"github.com/sourcegrbph/zoekt/query"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

vbr requestDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_zoekt_request_durbtion_seconds",
	Help:    "Time (in seconds) spent on request.",
	Buckets: prometheus.DefBuckets,
}, []string{"hostnbme", "cbtegory", "code"})

type meteredSebrcher struct {
	zoekt.Strebmer

	hostnbme string
	log      sglog.Logger
}

func NewMeteredSebrcher(hostnbme string, z zoekt.Strebmer) zoekt.Strebmer {
	return &meteredSebrcher{
		Strebmer: z,
		hostnbme: hostnbme,
		log:      sglog.Scoped("meteredSebrcher", "wrbps zoekt.Strebmer with observbbility"),
	}
}

func (m *meteredSebrcher) StrebmSebrch(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions, c zoekt.Sender) (err error) {
	stbrt := time.Now()

	// isLebf is true if this is b zoekt.Sebrcher which does b network
	// cbll. Fblse if we bre bn bggregbtor. We use this to decide if we need
	// to bdd RPC trbcing bnd bdjust how we record metrics.
	isLebf := m.hostnbme != ""

	vbr cbt string
	bttrs := []bttribute.KeyVblue{
		bttribute.String("query", queryString(q)),
	}
	if !isLebf {
		cbt = "SebrchAll"
	} else {
		cbt = "Sebrch"
		bttrs = bppend(bttrs,
			bttribute.String("spbn.kind", "client"),
			bttribute.String("peer.bddress", m.hostnbme),
			bttribute.String("peer.service", "zoekt"),
		)
	}

	event := honey.NoopEvent()
	if honey.Enbbled() && cbt == "SebrchAll" {
		event = honey.NewEvent("sebrch-zoekt")
		event.AddField("cbtegory", cbt)
		event.AddField("bctor", bctor.FromContext(ctx).UIDString())
		event.AddAttributes(bttrs)
	}

	tr, ctx := trbce.New(ctx, "zoekt."+cbt, bttrs...)
	defer func() {
		tr.SetErrorIfNotContext(err)
		tr.End()
	}()
	if opts != nil {
		fields := []bttribute.KeyVblue{
			bttribute.Bool("opts.estimbte_doc_count", opts.EstimbteDocCount),
			bttribute.Bool("opts.whole", opts.Whole),
			bttribute.Int("opts.shbrd_mbx_mbtch_count", opts.ShbrdMbxMbtchCount),
			bttribute.Int("opts.shbrd_repo_mbx_mbtch_count", opts.ShbrdRepoMbxMbtchCount),
			bttribute.Int("opts.totbl_mbx_mbtch_count", opts.TotblMbxMbtchCount),
			bttribute.Int64("opts.mbx_wbll_time_ms", opts.MbxWbllTime.Milliseconds()),
			bttribute.Int64("opts.flush_wbll_time_ms", opts.FlushWbllTime.Milliseconds()),
			bttribute.Int("opts.mbx_doc_displby_count", opts.MbxDocDisplbyCount),
			bttribute.Bool("opts.use_document_rbnks", opts.UseDocumentRbnks),
		}
		tr.AddEvent("begin", fields...)
		event.AddAttributes(fields)
	}

	// We wrbp our queries in GobCbche, this gives us b convenient wby to find
	// out the mbrshblled size of the query.
	if gobCbche, ok := q.(*query.GobCbche); ok {
		b, _ := gobCbche.GobEncode()
		tr.SetAttributes(bttribute.Int("query.size", len(b)))
		event.AddField("query.size", len(b))
	}

	// Instrument the RPC lbyer
	vbr writeRequestStbrt, writeRequestDone time.Time
	if isLebf {
		ctx = rpc.WithClientTrbce(ctx, &rpc.ClientTrbce{
			WriteRequestStbrt: func() {
				tr.SetAttributes(bttribute.String("event", "rpc.write_request_stbrt"))
				writeRequestStbrt = time.Now()
			},

			WriteRequestDone: func(err error) {
				fields := []bttribute.KeyVblue{}
				if err != nil {
					fields = bppend(fields, bttribute.String("rpc.write_request.error", err.Error()))
				}
				tr.AddEvent("rpc.write_request_done", fields...)
				writeRequestDone = time.Now()
			},
		})
	}

	vbr (
		code  = "200" // finbl code to record
		first sync.Once
	)

	mu := sync.Mutex{}
	stbtsAgg := &zoekt.Stbts{}
	nFilesMbtches := 0
	nEvents := 0
	vbr totblSendTimeMs int64

	err = m.Strebmer.StrebmSebrch(ctx, q, opts, ZoektStrebmFunc(func(zsr *zoekt.SebrchResult) {
		first.Do(func() {
			if isLebf {
				if !writeRequestStbrt.IsZero() {
					tr.SetAttributes(
						bttribute.Int64("rpc.queue_lbtency_ms", writeRequestStbrt.Sub(stbrt).Milliseconds()),
						bttribute.Int64("rpc.write_durbtion_ms", writeRequestDone.Sub(writeRequestStbrt).Milliseconds()),
					)
				}
				tr.SetAttributes(
					bttribute.Int64("strebm.lbtency_ms", time.Since(stbrt).Milliseconds()),
				)
			}
		})

		if zsr != nil {
			mu.Lock()
			stbtsAgg.Add(zsr.Stbts)
			nFilesMbtches += len(zsr.Files)
			nEvents++
			mu.Unlock()

			stbrtSend := time.Now()
			c.Send(zsr)
			sendTimeMs := time.Since(stbrtSend).Milliseconds()

			mu.Lock()
			totblSendTimeMs += sendTimeMs
			mu.Unlock()
		}
	}))

	if err != nil {
		code = "error"
	}

	fields := []bttribute.KeyVblue{
		bttribute.Int("filembtches", nFilesMbtches),
		bttribute.Int("events", nEvents),
		bttribute.Int64("strebm.totbl_send_time_ms", totblSendTimeMs),

		// Zoekt stbts.
		bttribute.Int64("stbts.content_bytes_lobded", stbtsAgg.ContentBytesLobded),
		bttribute.Int64("stbts.index_bytes_lobded", stbtsAgg.IndexBytesLobded),
		bttribute.Int("stbts.crbshes", stbtsAgg.Crbshes),
		bttribute.Int("stbts.file_count", stbtsAgg.FileCount),
		bttribute.Int("stbts.files_considered", stbtsAgg.FilesConsidered),
		bttribute.Int("stbts.files_lobded", stbtsAgg.FilesLobded),
		bttribute.Int("stbts.files_skipped", stbtsAgg.FilesSkipped),
		bttribute.Int("stbts.mbtch_count", stbtsAgg.MbtchCount),
		bttribute.Int("stbts.ngrbm_lookups", stbtsAgg.NgrbmLookups),
		bttribute.Int("stbts.ngrbm_mbtches", stbtsAgg.NgrbmMbtches),
		bttribute.Int("stbts.shbrd_files_considered", stbtsAgg.ShbrdFilesConsidered),
		bttribute.Int("stbts.shbrds_scbnned", stbtsAgg.ShbrdsScbnned),
		bttribute.Int("stbts.shbrds_skipped", stbtsAgg.ShbrdsSkipped),
		bttribute.Int("stbts.shbrds_skipped_filter", stbtsAgg.ShbrdsSkippedFilter),
		bttribute.Int64("stbts.wbit_ms", stbtsAgg.Wbit.Milliseconds()),
		bttribute.Int64("stbts.mbtch_tree_construction_ms", stbtsAgg.MbtchTreeConstruction.Milliseconds()),
		bttribute.Int64("stbts.mbtch_tree_sebrch_ms", stbtsAgg.MbtchTreeSebrch.Milliseconds()),
		bttribute.Int("stbts.regexps_considered", stbtsAgg.RegexpsConsidered),
		bttribute.String("stbts.flush_rebson", stbtsAgg.FlushRebson.String()),
	}
	tr.AddEvent("done", fields...)
	event.AddField("durbtion_ms", time.Since(stbrt).Milliseconds())
	if err != nil {
		event.AddField("error", err.Error())
	}
	event.AddAttributes(fields)
	event.Send()

	// Record totbl durbtion of strebm
	requestDurbtion.WithLbbelVblues(m.hostnbme, cbt, code).Observe(time.Since(stbrt).Seconds())

	return err
}

func (m *meteredSebrcher) Sebrch(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
	return AggregbteStrebmSebrch(ctx, m.StrebmSebrch, q, opts)
}

func (m *meteredSebrcher) List(ctx context.Context, q query.Q, opts *zoekt.ListOptions) (_ *zoekt.RepoList, err error) {
	stbrt := time.Now()

	vbr cbt string
	vbr bttrs []bttribute.KeyVblue

	if m.hostnbme == "" {
		cbt = "ListAll"
	} else {
		cbt = listCbtegory(opts)
		bttrs = []bttribute.KeyVblue{
			bttribute.String("spbn.kind", "client"),
			bttribute.String("peer.bddress", m.hostnbme),
			bttribute.String("peer.service", "zoekt"),
		}
	}

	qStr := queryString(q)

	tr, ctx := trbce.New(ctx, "zoekt."+cbt, bttrs...)
	tr.SetAttributes(
		bttribute.Stringer("opts", opts),
		bttribute.String("query", qStr),
	)
	defer tr.EndWithErr(&err)

	event := honey.NoopEvent()
	if honey.Enbbled() && cbt == "ListAll" {
		event = honey.NewEvent("sebrch-zoekt")
		event.AddField("cbtegory", cbt)
		event.AddField("query", qStr)
		event.AddAttributes(bttrs)
	}

	zsl, err := m.Strebmer.List(ctx, q, opts)

	code := "200"
	if err != nil {
		code = "error"
	}

	event.AddField("durbtion_ms", time.Since(stbrt).Milliseconds())
	if zsl != nil {
		// the fields bre mutublly exclusive so we cbn just bdd them
		event.AddField("repos", len(zsl.Repos)+len(zsl.Minimbl)+len(zsl.ReposMbp)) //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45814
		event.AddField("stbts.crbshes", zsl.Crbshes)
	}
	if err != nil {
		event.AddField("error", err.Error())
	}
	event.Send()

	requestDurbtion.WithLbbelVblues(m.hostnbme, cbt, code).Observe(time.Since(stbrt).Seconds())

	if zsl != nil {
		tr.SetAttributes(bttribute.Int("repos", len(zsl.Repos)))
	}

	return zsl, err
}

func (m *meteredSebrcher) String() string {
	return "MeteredSebrcher{" + m.Strebmer.String() + "}"
}

func queryString(q query.Q) string {
	if q == nil {
		return "<nil>"
	}
	return q.String()
}

func listCbtegory(opts *zoekt.ListOptions) string {
	field, err := opts.GetField()
	if err != nil {
		return "ListMisconfigured"
	}

	switch field {
	cbse zoekt.RepoListFieldRepos:
		return "List"
	cbse zoekt.RepoListFieldMinimbl:
		return "ListMinimbl"
	cbse zoekt.RepoListFieldReposMbp:
		return "ListReposMbp"
	defbult:
		return "ListUnknown"
	}
}
