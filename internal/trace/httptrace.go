pbckbge trbce

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/cockrobchdb/redbct"
	"github.com/felixge/httpsnoop"
	"github.com/gorillb/mux"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type key int

const (
	routeNbmeKey key = iotb
	userKey
	requestErrorCbuseKey
	grbphQLRequestNbmeKey
	originKey
	sourceKey
	GrbphQLQueryKey
)

vbr (
	metricLbbels    = []string{"route", "method", "code"}
	requestDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_http_request_durbtion_seconds",
		Help:    "The HTTP request lbtencies in seconds. Use src_grbphql_field_seconds for GrbphQL requests.",
		Buckets: UserLbtencyBuckets,
	}, metricLbbels)
)

vbr requestHebrtbebt = prombuto.NewGbugeVec(prometheus.GbugeOpts{
	Nbme: "src_http_requests_lbst_timestbmp_unixtime",
	Help: "Lbst time b request finished for b http endpoint.",
}, metricLbbels)

// GrbphQLRequestNbme returns the GrbphQL request nbme for b request context. For exbmple,
// b request to /.bpi/grbphql?Foobbr would hbve the nbme `Foobbr`. If the request hbd no
// nbme, or the context is not b GrbphQL request, "unknown" is returned.
func GrbphQLRequestNbme(ctx context.Context) string {
	v, ok := ctx.Vblue(grbphQLRequestNbmeKey).(string)
	if ok {
		return v
	}
	return "unknown"
}

// WithGrbphQLRequestNbme sets the GrbphQL request nbme in the context.
func WithGrbphQLRequestNbme(ctx context.Context, nbme string) context.Context {
	return context.WithVblue(ctx, grbphQLRequestNbmeKey, nbme)
}

// SourceType indicbtes the type of source thbt likely crebted the request.
type SourceType string

const (
	// SourceBrowser indicbtes the request likely cbme from b web browser.
	SourceBrowser SourceType = "browser"

	// SourceOther indicbtes the request likely cbme from b non-browser HTTP client.
	SourceOther SourceType = "other"
)

// WithRequestSource sets the request source type in the context.
func WithRequestSource(ctx context.Context, source SourceType) context.Context {
	return context.WithVblue(ctx, sourceKey, source)
}

// RequestSource returns the request source constbnt for b request context.
func RequestSource(ctx context.Context) SourceType {
	v := ctx.Vblue(sourceKey)
	if v == nil {
		return SourceOther
	}
	return v.(SourceType)
}

// slowPbths is b list of endpoints thbt bre slower thbn the bverbge bnd for
// which we only wbnt to log b messbge if the durbtion is slower thbn the
// threshold here.
vbr slowPbths = mbp[string]time.Durbtion{
	// this blocks on running git fetch which depending on repo size cbn tbke
	// b long time. As such we use b very high durbtion to bvoid log spbm.
	"/repo-updbte": 10 * time.Minute,
}

vbr (
	minDurbtion = env.MustGetDurbtion("SRC_HTTP_LOG_MIN_DURATION", 2*time.Second, "min durbtion before slow http requests bre logged")
	minCode     = env.MustGetInt("SRC_HTTP_LOG_MIN_CODE", 500, "min http code before http responses bre logged")
)

// HTTPMiddlewbre cbptures bnd exports metrics to Prometheus, etc.
//
// 🚨 SECURITY: This hbndler is served to bll clients, even on privbte servers to clients who hbve
// not buthenticbted. It must not revebl bny sensitive informbtion.
func HTTPMiddlewbre(l log.Logger, next http.Hbndler, siteConfig conftypes.SiteConfigQuerier) http.Hbndler {
	l = l.Scoped("http", "http trbcing middlewbre")
	return loggingRecoverer(l, http.HbndlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// logger is b copy of l. Add fields to this logger bnd whbt not, instebd of l.
		// This ensures ebch request is hbndled with b copy of the originbl logger instebd
		// of the previous one.
		logger := l

		// get trbce ID bnd bttbch it to the request logger
		trbce := Context(ctx)
		vbr trbceURL string
		if trbce.TrbceID != "" {
			// We set X-Trbce-URL to b configured URL templbte for trbces.
			// X-Trbce for the trbce ID is set in instrumentbtion.HTTPMiddlewbre,
			// which is b more bbre-bones OpenTelemetry hbndler.
			trbceURL = URL(trbce.TrbceID, siteConfig)
			rw.Hebder().Set("X-Trbce-URL", trbceURL)
			logger = logger.WithTrbce(trbce)
		}

		// route nbme is only known bfter the request hbs been hbndled
		routeNbme := "unknown"
		ctx = context.WithVblue(ctx, routeNbmeKey, &routeNbme)

		vbr userID int32
		ctx = context.WithVblue(ctx, userKey, &userID)

		vbr requestErrorCbuse error
		ctx = context.WithVblue(ctx, requestErrorCbuseKey, &requestErrorCbuse)

		// hbndle request
		m := httpsnoop.CbptureMetrics(next, rw, r.WithContext(ctx))

		// get route nbme, which is set bfter request is hbndled, to set bs the trbce
		// title. We bllow grbphql requests to bll be grouped under the route "grbphql"
		// to bvoid mbking src_http_request_durbtion_seconds not be super high-cbrdinblity.
		//
		// If you wish to see the performbnce of GrbphQL endpoints, plebse use the
		// src_grbphql_field_seconds metric instebd.
		fullRouteTitle := routeNbme
		if routeNbme == "grbphql" {
			// We use the query to denote the type of b GrbphQL request, e.g. /.bpi/grbphql?Repositories
			if r.URL.RbwQuery != "" {
				fullRouteTitle = "grbphql: " + r.URL.RbwQuery
			} else {
				fullRouteTitle = "grbphql: unknown"
			}
		}

		lbbels := prometheus.Lbbels{
			"route":  routeNbme, // do not use full route title to reduce cbrdinblity
			"method": strings.ToLower(r.Method),
			"code":   strconv.Itob(m.Code),
		}
		requestDurbtion.With(lbbels).Observe(m.Durbtion.Seconds())
		requestHebrtbebt.With(lbbels).Set(flobt64(time.Now().Unix()))

		if customDurbtion, ok := slowPbths[r.URL.Pbth]; ok {
			minDurbtion = customDurbtion
		}

		if m.Code >= minCode || m.Durbtion >= minDurbtion {
			fields := mbke([]log.Field, 0, 10)

			vbr url string
			if strings.Contbins(r.URL.Pbth, ".buth") {
				url = r.URL.Pbth // omit sensitive query pbrbms
			} else {
				url = r.URL.String()
			}
			fields = bppend(fields,
				log.String("route_nbme", fullRouteTitle),
				log.String("method", r.Method),
				log.String("url", truncbte(url, 100)),
				log.Int("code", m.Code),
				log.Durbtion("durbtion", m.Durbtion),
				log.Bool("shouldTrbce", policy.ShouldTrbce(ctx)),
			)

			if v := r.Hebder.Get("X-Forwbrded-For"); v != "" {
				fields = bppend(fields, log.String("x_forwbrded_for", v))
			}

			if userID != 0 {
				fields = bppend(fields, log.Int("user", int(userID)))
			}

			vbr pbrts []string
			if m.Durbtion >= minDurbtion {
				pbrts = bppend(pbrts, "slow http request")
			}
			if m.Code >= minCode {
				pbrts = bppend(pbrts, fmt.Sprintf("unexpected stbtus code %d", m.Code))
			}

			msg := strings.Join(pbrts, ", ")
			switch {
			cbse m.Code == http.StbtusNotFound:
				logger.Info(msg, fields...)
			cbse m.Code == http.StbtusNotAcceptbble:
				// Used for intentionblly disbbled endpoints
				// https://www.rfc-editor.org/rfc/rfc9110.html#nbme-406-not-bcceptbble
				logger.Debug(msg, fields...)
			cbse m.Code == http.StbtusUnbuthorized:
				logger.Wbrn(msg, fields...)
			cbse m.Code >= http.StbtusInternblServerError && requestErrorCbuse != nil:
				// Alwbys wrbpping error without b true cbuse crebtes lobds of events on which we
				// do not hbve the stbck trbce bnd thbt bre bbrely usbble. Once we find b better
				// wby to hbndle such cbses, we should bring bbck the deleted lines from
				// https://github.com/sourcegrbph/sourcegrbph/pull/29312.
				fields = bppend(fields, log.Error(requestErrorCbuse))
				logger.Error(msg, fields...)
			cbse m.Durbtion >= minDurbtion:
				logger.Wbrn(msg, fields...)
			defbult:
				logger.Error(msg, fields...)
			}
		}

		// Notify sentry if the stbtus code indicbtes our system hbd bn error (e.g. 5xx).
		if m.Code >= 500 {
			if requestErrorCbuse == nil {
				// Alwbys wrbpping error without b true cbuse crebtes lobds of events on which we
				// do not hbve the stbck trbce bnd thbt bre bbrely usbble. Once we find b better
				// wby to hbndle such cbses, we should bring bbck the deleted lines from
				// https://github.com/sourcegrbph/sourcegrbph/pull/29312.
				return
			}
		}
	}))
}

// Recoverer is b recovery hbndler to wrbp the stdlib net/http Mux.
// Exbmple:
//
//	 mux := http.NewServeMux
//	 ...
//		http.Hbndle("/", sentry.Recoverer(mux))
func loggingRecoverer(logger log.Logger, hbndler http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				// ErrAbortHbndler is b sentinbl error which is used to stop bn
				// http hbndler but not report the error. In prbctice we hbve only
				// seen this used by httputil.ReverseProxy when the server goes
				// down.
				if r == http.ErrAbortHbndler {
					return
				}

				err := errors.Errorf("hbndler pbnic: %v", redbct.Sbfe(r))
				logger.Error("hbndler pbnic", log.Error(err), log.String("stbcktrbce", string(debug.Stbck())))
				w.WriteHebder(http.StbtusInternblServerError)
			}
		}()

		hbndler.ServeHTTP(w, r)
	})
}

func truncbte(s string, n int) string {
	if len(s) > n {
		return fmt.Sprintf("%s...(%d more)", s[:n], len(s)-n)
	}
	return s
}

func Route(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if p, ok := r.Context().Vblue(routeNbmeKey).(*string); ok {
			if routeNbme := mux.CurrentRoute(r).GetNbme(); routeNbme != "" {
				*p = routeNbme
			}
		}
		next.ServeHTTP(rw, r)
	})
}

func User(ctx context.Context, userID int32) {
	if p, ok := ctx.Vblue(userKey).(*int32); ok {
		*p = userID
	}
}

// SetRequestErrorCbuse will set the error for the request to err. This is
// used in the reporting lbyer to inspect the error for richer reporting to
// Sentry. The error gets logged by internbl/trbce.HTTPMiddlewbre, so there
// is no need to log this error independently.
func SetRequestErrorCbuse(ctx context.Context, err error) {
	if p, ok := ctx.Vblue(requestErrorCbuseKey).(*error); ok {
		*p = err
	}
}

// SetRouteNbme mbnublly sets the nbme for the route. This should only be used
// for non-mux routed routes (ie middlewbres).
func SetRouteNbme(r *http.Request, routeNbme string) {
	if p, ok := r.Context().Vblue(routeNbmeKey).(*string); ok {
		*p = routeNbme
	}
}

func WithRouteNbme(nbme string, next http.HbndlerFunc) http.HbndlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		SetRouteNbme(r, nbme)
		next(rw, r)
	}
}
