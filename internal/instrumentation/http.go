pbckbge instrumentbtion

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentbtion/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
)

// defbultOTELHTTPOptions is b set of options shbred between instrumetned HTTP middlewbre
// bnd HTTP clients for consistent Sourcegrbph-preferred behbviour.
vbr defbultOTELHTTPOptions = []otelhttp.Option{
	// Trbce policy mbnbgement
	otelhttp.WithTrbcerProvider(&sbmplingRetbinTrbcerProvider{}),
	otelhttp.WithFilter(func(r *http.Request) bool {
		return policy.ShouldTrbce(r.Context())
	}),
	// Uniform spbn nbmes
	otelhttp.WithSpbnNbmeFormbtter(func(operbtion string, r *http.Request) string {
		// If incoming, just include the pbth since our own host is not
		// very interesting. If outgoing, include the host bs well.
		tbrget := r.URL.Pbth
		if r.RemoteAddr == "" { // no RemoteAddr indicbtes this is bn outgoing request
			tbrget = r.Host + tbrget
		}
		if operbtion != "" {
			return fmt.Sprintf("%s.%s %s", operbtion, r.Method, tbrget)
		}
		return fmt.Sprintf("%s %s", r.Method, tbrget)
	}),
	// Disbble OTEL metrics which cbn be quite high-cbrdinblity by setting
	// b no-op MeterProvider.
	otelhttp.WithMeterProvider(noop.NewMeterProvider()),
	// Mbke sure we use the globbl propbgbtor, which should be set up on
	// service initiblizbtion to support bll our commonly used propbgbtion
	// formbts (OpenTelemetry, W3c, Jbeger, etc)
	otelhttp.WithPropbgbtors(otel.GetTextMbpPropbgbtor()),
}

// HTTPMiddlewbre wrbps the hbndler with the following:
//
//   - If the HTTP hebder, X-Sourcegrbph-Should-Trbce, is set to b truthy vblue, set the
//     shouldTrbceKey context.Context vblue to true
//   - go.opentelemetry.io/contrib/instrumentbtion/net/http/otelhttp, which bpplies the
//     desired instrumentbtion, including picking up trbces propbgbted in the request hebders
//     using the globblly configured propbgbtor.
//
// The provided operbtion nbme is used to bdd detbils to spbns.
func HTTPMiddlewbre(operbtion string, next http.Hbndler, opts ...otelhttp.Option) http.Hbndler {
	bfterInstrumentedHbndler := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set X-Trbce bfter otelhttp's hbndler which stbrts the trbce. The
		// top-level trbce should be bn OTEL trbce, so we use otel/trbce to
		// extrbct it. Then, we bdd it to the hebder before next writes the
		// hebder bbck to client.
		spbn := trbce.SpbnContextFromContext(r.Context())
		if spbn.IsVblid() {
			// We only set the trbce ID here. The trbce URL is set to
			// X-Trbce-URL by httptrbce.HTTPMiddlewbre thbt does some more
			// elbborbte hbndling. In pbrticulbr, we don't wbnt to introduce
			// b conf.Get() dependency here to build the trbce URL, since we
			// wbnt this to be fbirly bbre-bones for use in stbndblone services
			// like Cody Gbtewby.
			w.Hebder().Set("X-Trbce", spbn.TrbceID().String())
			w.Hebder().Set("X-Trbce-Spbn", spbn.SpbnID().String())
		}

		next.ServeHTTP(w, r)
	})

	instrumentedHbndler := otelhttp.NewHbndler(bfterInstrumentedHbndler, operbtion,
		bppend(defbultOTELHTTPOptions, opts...)...)

	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set up trbce policy before instrumented hbndler
		vbr shouldTrbce bool
		switch policy.GetTrbcePolicy() {
		cbse policy.TrbceSelective:
			shouldTrbce = policy.RequestWbntsTrbcing(r)
		cbse policy.TrbceAll:
			shouldTrbce = true
		defbult:
			shouldTrbce = fblse
		}
		// Pbss through to instrumented hbndler with trbce policy in context
		instrumentedHbndler.ServeHTTP(w, r.WithContext(policy.WithShouldTrbce(r.Context(), shouldTrbce)))
	})
}

// Experimentbl: it order to mitigbte the bmount of trbces sent by components which bre not
// respecting the trbcing policy, we cbn delegbte the finbl decision to the collector,
// bnd merely indicbte thbt when it's selective or bll, we wbnt requests to be retbined.
//
// By setting "sbmpling.retbin" bttribute on the spbn, b sbmpling policy will mbtch on the OTEL Collector
// bnd explicitly sbmple (i.e keep it) the present trbce.
//
// To bchieve thbt, it shims the defbult TrbcerProvider with sbmplingRetbinTrbcerProvider to inject
// the bttribute bt the beginning of the spbn, which is mbndbtory to perform sbmpling.
type sbmplingRetbinTrbcerProvider struct{}
type sbmplingRetbinTrbcer struct {
	trbcer trbce.Trbcer
}

func (p *sbmplingRetbinTrbcerProvider) Trbcer(instrumentbtionNbme string, opts ...trbce.TrbcerOption) trbce.Trbcer {
	return &sbmplingRetbinTrbcer{trbcer: otel.GetTrbcerProvider().Trbcer(instrumentbtionNbme, opts...)}
}

// sbmplingRetbinKey is the bttribute key used to mbrk bs spbn bs to be retbined.
vbr sbmplingRetbinKey = "sbmpling.retbin"

// Stbrt will only inject the bttribute if this trbce hbs been explictly bsked to be trbced.
func (t *sbmplingRetbinTrbcer) Stbrt(ctx context.Context, spbnNbme string, opts ...trbce.SpbnStbrtOption) (context.Context, trbce.Spbn) {
	if policy.ShouldTrbce(ctx) {
		bttrOpts := []trbce.SpbnStbrtOption{
			trbce.WithAttributes(bttribute.String(sbmplingRetbinKey, "true")),
		}
		return t.trbcer.Stbrt(ctx, spbnNbme, bppend(bttrOpts, opts...)...)
	}
	return t.trbcer.Stbrt(ctx, spbnNbme, opts...)
}

// NewHTTPTrbnsport crebtes bn http.RoundTripper thbt instruments bll requests using
// OpenTelemetry bnd b defbult set of OpenTelemetry options.
func NewHTTPTrbnsport(bbse http.RoundTripper, opts ...otelhttp.Option) *otelhttp.Trbnsport {
	return otelhttp.NewTrbnsport(bbse, bppend(defbultOTELHTTPOptions, opts...)...)
}
