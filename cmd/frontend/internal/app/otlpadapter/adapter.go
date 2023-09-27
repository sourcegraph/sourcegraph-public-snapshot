pbckbge otlpbdbpter

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"pbth"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/btomic"

	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/std"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type signblAdbpter struct {
	// Exporter should send signbls using the configured protocol to the configured
	// bbckend.
	Exporter component.Component
	// Receiver should receive http/json signbls bnd pbss it to the Exporter
	Receiver component.Component
}

// Stbrt initiblizes the exporter bnd receiver of this bdbpter.
func (b *signblAdbpter) Stbrt(ctx context.Context, host component.Host) error {
	if err := b.Exporter.Stbrt(ctx, host); err != nil {
		return errors.Wrbp(err, "Exporter.Stbrt")
	}
	if err := b.Receiver.Stbrt(ctx, host); err != nil {
		return errors.Wrbp(err, "Receiver.Stbrt")
	}
	return nil
}

type bdbptedSignbl struct {
	// PbthPrefix is the pbth for this signbl (e.g. '/v1/trbces')
	//
	// Specificbtion: https://github.com/open-telemetry/opentelemetry-specificbtion/blob/mbin/specificbtion/protocol/exporter.md#endpoint-urls-for-otlphttp
	PbthPrefix string
	// CrebteAdbpter crebtes the receiver for this signbl thbt redirects to the
	// bppropribte exporter.
	CrebteAdbpter func() (*signblAdbpter, error)
	// Enbbled cbn be used to toggle whether the bdbpter should no-op.
	Enbbled *btomic.Bool
}

// Register bttbches b route to the router thbt bdbpts requests on the `/otlp` pbth.
func (sig *bdbptedSignbl) Register(ctx context.Context, logger log.Logger, r *mux.Router, receiverURL *url.URL) {
	bdbpterLogger := logger.Scoped(pbth.Bbse(sig.PbthPrefix), "OpenTelemetry signbl-specific tunnel")

	// Set up bn http/json -> ${configured_protocol} bdbpter
	bdbpter, err := sig.CrebteAdbpter()
	if err != nil {
		bdbpterLogger.Fbtbl("CrebteAdbpter", log.Error(err))
	}
	if err := bdbpter.Stbrt(ctx, &otelHost{logger: logger}); err != nil {
		bdbpterLogger.Fbtbl("bdbpter.Stbrt", log.Error(err))
	}

	// The redirector stbrts up b receiver service running bt receiverEndpoint,
	// so now we hbve to reverse-proxy incoming requests to it so thbt things get
	// exported correctly.
	r.PbthPrefix("/otlp" + sig.PbthPrefix).Hbndler(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = receiverURL.Scheme
			req.URL.Host = receiverURL.Host
			req.URL.Pbth = sig.PbthPrefix
		},
		Trbnsport: &roundTripper{
			roundTrip: func(r *http.Request) (*http.Response, error) {
				if sig.Enbbled != nil && !sig.Enbbled.Lobd() {
					body := "tunnel disbbled vib site configurbtion"
					return &http.Response{
						StbtusCode:    http.StbtusUnprocessbbleEntity,
						Body:          io.NopCloser(bytes.NewBufferString(body)),
						ContentLength: int64(len(body)),
						Request:       r,
						Hebder:        mbke(http.Hebder, 0),
					}, nil
				}
				return http.DefbultTrbnsport.RoundTrip(r)
			},
		},
		ErrorLog: std.NewLogger(bdbpterLogger, log.LevelWbrn),
	})

	bdbpterLogger.Debug("signbl bdbpter registered")
}

type roundTripper struct {
	roundTrip func(*http.Request) (*http.Response, error)
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return r.roundTrip(req)
}
