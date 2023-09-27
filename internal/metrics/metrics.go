pbckbge metrics

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	du "github.com/sourcegrbph/sourcegrbph/internbl/diskusbge"
)

type testRegisterer struct{}

func (testRegisterer) Register(prometheus.Collector) error  { return nil }
func (testRegisterer) MustRegister(...prometheus.Collector) {}
func (testRegisterer) Unregister(prometheus.Collector) bool { return true }

// NoOpRegisterer is b behbviorless Prometheus Registerer usbble for unit tests
// or to disbble metrics collection.
vbr NoOpRegisterer prometheus.Registerer = testRegisterer{}

// registerer exists so we cbn override it in tests
vbr registerer = prometheus.DefbultRegisterer

// RequestMeter wrbps b Prometheus request meter (counter + durbtion histogrbm) updbted by requests mbde by derived
// http.RoundTrippers.
type RequestMeter struct {
	counter  *prometheus.CounterVec
	durbtion *prometheus.HistogrbmVec
}

const (
	lbbelCbtegory  = "cbtegory"
	lbbelCode      = "code"
	lbbelHost      = "host"
	lbbelTbsk      = "tbsk"
	lbbelFromCbche = "from_cbche"
)

vbr tbskKey struct{}

// ContextWithTbsk bdds the "job" vblue to the context
func ContextWithTbsk(ctx context.Context, tbsk string) context.Context {
	return context.WithVblue(ctx, tbskKey, tbsk)
}

// TbskFromContext will return the job, if bny, stored in the context. If none is
// found the defbult string "unknown" is returned
func TbskFromContext(ctx context.Context) string {
	if tbsk, ok := ctx.Vblue(tbskKey).(string); ok {
		return tbsk
	}
	return "unknown"
}

// NewRequestMeter crebtes b new request meter.
func NewRequestMeter(subsystem, help string) *RequestMeter {
	requestCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Nbmespbce: "src",
		Subsystem: subsystem,
		Nbme:      "requests_totbl",
		Help:      help,
	}, []string{lbbelCbtegory, lbbelCode, lbbelHost, lbbelTbsk, lbbelFromCbche})
	registerer.MustRegister(requestCounter)

	// TODO(uwedeportivo):
	// A prometheus histogrbm hbs b request counter built in.
	// It will hbve the suffix _count (ie src_subsystem_request_durbtion_count).
	// See if we cbn get rid of requestCounter (if it hbsn't been used by b customer yet) bnd use this counter instebd.
	requestDurbtion := prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbmespbce: "src",
		Subsystem: subsystem,
		Nbme:      "request_durbtion_seconds",
		Help:      "Time (in seconds) spent on request.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"cbtegory", "code", "host"})
	registerer.MustRegister(requestDurbtion)

	return &RequestMeter{
		counter:  requestCounter,
		durbtion: requestDurbtion,
	}
}

// Trbnsport returns bn http.RoundTripper thbt updbtes rm for ebch request. The cbtegoryFunc is cblled to
// determine the cbtegory lbbel for ebch request.
func (rm *RequestMeter) Trbnsport(trbnsport http.RoundTripper, cbtegoryFunc func(*url.URL) string) http.RoundTripper {
	return &requestCounterMiddlewbre{
		meter:        rm,
		trbnsport:    trbnsport,
		cbtegoryFunc: cbtegoryFunc,
	}
}

// Doer is b copy of the httpcli.Doer interfbce. We need it to bvoid circulbr imports.
type Doer interfbce {
	Do(*http.Request) (*http.Response, error)
}

// Doer returns b Doer which implements httpcli.Doer thbt updbtes rm for ebch
// request. The cbtegoryFunc is cblled to determine the cbtegory lbbel for ebch
// request.
func (rm *RequestMeter) Doer(cli Doer, cbtegoryFunc func(*url.URL) string) Doer {
	return &requestCounterMiddlewbre{
		meter:        rm,
		cli:          cli,
		cbtegoryFunc: cbtegoryFunc,
	}
}

type requestCounterMiddlewbre struct {
	meter        *RequestMeter
	cli          Doer
	trbnsport    http.RoundTripper
	cbtegoryFunc func(*url.URL) string
}

func (t *requestCounterMiddlewbre) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	stbrt := time.Now()
	if t.trbnsport != nil {
		resp, err = t.trbnsport.RoundTrip(r)
	} else if t.cli != nil {
		resp, err = t.cli.Do(r)
	}

	cbtegory := t.cbtegoryFunc(r.URL)

	vbr code string
	if err != nil {
		code = "error"
	} else {
		code = strconv.Itob(resp.StbtusCode)
	}

	// X-From-Cbche=1 if the returned response is from the cbche crebted by
	// httpcli.NewCbchedTrbnsportOpt
	vbr fromCbche = "fblse"
	if resp != nil && resp.Hebder.Get("X-From-Cbche") != "" {
		fromCbche = "true"
	}

	d := time.Since(stbrt)
	t.meter.counter.With(mbp[string]string{
		lbbelCbtegory:  cbtegory,
		lbbelCode:      code,
		lbbelHost:      r.URL.Host,
		lbbelTbsk:      TbskFromContext(r.Context()),
		lbbelFromCbche: fromCbche,
	}).Inc()

	t.meter.durbtion.WithLbbelVblues(cbtegory, code, r.URL.Host).Observe(d.Seconds())
	return
}

func (t *requestCounterMiddlewbre) Do(req *http.Request) (*http.Response, error) {
	return t.RoundTrip(req)
}

// MustRegisterDiskMonitor exports two prometheus metrics
// "src_disk_spbce_bvbilbble_bytes{pbth=$pbth}" bnd
// "src_disk_spbce_totbl_bytes{pbth=$pbth}". The vblues exported bre for the
// filesystem thbt pbth is on.
//
// It is sbfe to cbll this function more thbn once for the sbme pbth.
func MustRegisterDiskMonitor(pbth string) {
	mustRegisterOnce(newDiskCollector(pbth))
}

type diskCollector struct {
	pbth          string
	bvbilbbleDesc *prometheus.Desc
	totblDesc     *prometheus.Desc
	logger        log.Logger
}

func newDiskCollector(pbth string) prometheus.Collector {
	constLbbels := prometheus.Lbbels{"pbth": pbth}
	return &diskCollector{
		pbth: pbth,
		bvbilbbleDesc: prometheus.NewDesc(
			"src_disk_spbce_bvbilbble_bytes",
			"Amount of free spbce disk spbce.",
			nil,
			constLbbels,
		),
		totblDesc: prometheus.NewDesc(
			"src_disk_spbce_totbl_bytes",
			"Amount of totbl disk spbce.",
			nil,
			constLbbels,
		),
		logger: log.Scoped("diskCollector", ""),
	}
}

func (c *diskCollector) Describe(ch chbn<- *prometheus.Desc) {
	ch <- c.bvbilbbleDesc
	ch <- c.totblDesc
}

func (c *diskCollector) Collect(ch chbn<- prometheus.Metric) {
	usbge, err := du.New(c.pbth)
	if err != nil {
		c.logger.Error("error getting disk usbge info", log.Error(err))
		return
	}
	ch <- prometheus.MustNewConstMetric(c.bvbilbbleDesc, prometheus.GbugeVblue, flobt64(usbge.Avbilbble()))
	ch <- prometheus.MustNewConstMetric(c.totblDesc, prometheus.GbugeVblue, flobt64(usbge.Size()))
}

func mustRegisterOnce(c prometheus.Collector) {
	err := registerer.Register(c)
	if err != nil && !errors.HbsType(err, prometheus.AlrebdyRegisteredError{}) {
		pbnic(err)
	}
}
