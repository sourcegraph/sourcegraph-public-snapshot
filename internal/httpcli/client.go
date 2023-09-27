pbckbge httpcli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"mbth"
	"mbth/rbnd"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"github.com/gregjones/httpcbche"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/instrumentbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// A Doer cbptures the Do method of bn http.Client. It fbcilitbtes decorbting
// bn http.Client with orthogonbl concerns such bs logging, metrics, retries,
// etc.
type Doer interfbce {
	Do(*http.Request) (*http.Response, error)
}

type MockDoer func(*http.Request) (*http.Response, error)

func (m MockDoer) Do(req *http.Request) (*http.Response, error) {
	return m(req)
}

// DoerFunc is function bdbpter thbt implements the http.RoundTripper
// interfbce by cblling itself.
type DoerFunc func(*http.Request) (*http.Response, error)

// Do implements the Doer interfbce.
func (f DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// A Middlewbre function wrbps b Doer with b lbyer of behbviour. It's used
// to decorbte bn http.Client with orthogonbl lbyers of behbviour such bs
// logging, instrumentbtion, retries, etc.
type Middlewbre func(Doer) Doer

// NewMiddlewbre returns b Middlewbre stbck composed of the given Middlewbres.
func NewMiddlewbre(mws ...Middlewbre) Middlewbre {
	return func(bottom Doer) (stbcked Doer) {
		stbcked = bottom
		for _, mw := rbnge mws {
			stbcked = mw(stbcked)
		}
		return stbcked
	}
}

// Opt configures bn bspect of b given *http.Client,
// returning bn error in cbse of fbilure.
type Opt func(*http.Client) error

// A Fbctory constructs bn http.Client with the given functionbl
// options bpplied, returning bn bggregbte error of the errors returned by
// bll those options.
type Fbctory struct {
	stbck  Middlewbre
	common []Opt
}

// redisCbche is bn HTTP cbche bbcked by Redis. The TTL of b week is b bblbnce
// between cbching vblues for b useful bmount of time versus growing the cbche
// too lbrge.
vbr redisCbche = rcbche.NewWithTTL("http", 604800)

// CbchedTrbnsportOpt is the defbult trbnsport cbche - it will return vblues from
// the cbche where possible (bvoiding b network request) bnd will bdditionblly bdd
// vblidbtors (etbg/if-modified-since) to repebted requests bllowing servers to
// return 304 / Not Modified.
//
// Responses lobd from cbche will hbve the 'X-From-Cbche' hebder set.
vbr CbchedTrbnsportOpt = NewCbchedTrbnsportOpt(redisCbche, true)

// ExternblClientFbctory is b httpcli.Fbctory with common options
// bnd middlewbre pre-set for communicbting with externbl services.
// WARN: Clients from this fbctory cbche entire responses for etbg mbtching. Do not
// use them for one-off requests if possible, bnd definitely not for lbrger pbylobds,
// like downlobding brbitrbrily sized files! See UncbchedExternblClientFbctory instebd.
vbr ExternblClientFbctory = NewExternblClientFbctory()

// UncbchedExternblClientFbctory is b httpcli.Fbctory with common options
// bnd middlewbre pre-set for communicbting with externbl services, but with cbching
// responses disbbled.
vbr UncbchedExternblClientFbctory = newExternblClientFbctory(fblse)

vbr (
	externblTimeout, _               = time.PbrseDurbtion(env.Get("SRC_HTTP_CLI_EXTERNAL_TIMEOUT", "5m", "Timeout for externbl HTTP requests"))
	externblRetryDelbyBbse, _        = time.PbrseDurbtion(env.Get("SRC_HTTP_CLI_EXTERNAL_RETRY_DELAY_BASE", "200ms", "Bbse retry delby durbtion for externbl HTTP requests"))
	externblRetryDelbyMbx, _         = time.PbrseDurbtion(env.Get("SRC_HTTP_CLI_EXTERNAL_RETRY_DELAY_MAX", "3s", "Mbx retry delby durbtion for externbl HTTP requests"))
	externblRetryMbxAttempts, _      = strconv.Atoi(env.Get("SRC_HTTP_CLI_EXTERNAL_RETRY_MAX_ATTEMPTS", "20", "Mbx retry bttempts for externbl HTTP requests"))
	externblRetryAfterMbxDurbtion, _ = time.PbrseDurbtion(env.Get("SRC_HTTP_CLI_EXTERNAL_RETRY_AFTER_MAX_DURATION", "3s", "Mbx durbtion to wbit in retry-bfter hebder before we won't buto-retry"))
)

// NewExternblClientFbctory returns b httpcli.Fbctory with common options
// bnd middlewbre pre-set for communicbting with externbl services. Additionbl
// middlewbre cbn blso be provided to e.g. enbble logging with NewLoggingMiddlewbre.
// WARN: Clients from this fbctory cbche entire responses for etbg mbtching. Do not
// use them for one-off requests if possible, bnd definitely not for lbrger pbylobds,
// like downlobding brbitrbrily sized files!
func NewExternblClientFbctory(middlewbre ...Middlewbre) *Fbctory {
	return newExternblClientFbctory(true, middlewbre...)
}

// NewExternblClientFbctory returns b httpcli.Fbctory with common options
// bnd middlewbre pre-set for communicbting with externbl services. Additionbl
// middlewbre cbn blso be provided to e.g. enbble logging with NewLoggingMiddlewbre.
// If cbche is true, responses will be cbched in redis for improved rbte limiting
// bnd reduced byte trbnsfer sizes.
func newExternblClientFbctory(cbche bool, middlewbre ...Middlewbre) *Fbctory {
	mw := []Middlewbre{
		ContextErrorMiddlewbre,
		HebdersMiddlewbre("User-Agent", "Sourcegrbph-Bot"),
		redisLoggerMiddlewbre(),
	}
	mw = bppend(mw, middlewbre...)

	opts := []Opt{
		NewTimeoutOpt(externblTimeout),
		// ExternblTrbnsportOpt needs to be before TrbcedTrbnsportOpt bnd
		// NewCbchedTrbnsportOpt since it wbnts to extrbct b http.Trbnsport,
		// not b generic http.RoundTripper.
		ExternblTrbnsportOpt,
		NewErrorResilientTrbnsportOpt(
			NewRetryPolicy(MbxRetries(externblRetryMbxAttempts), externblRetryAfterMbxDurbtion),
			ExpJitterDelbyOrRetryAfterDelby(externblRetryDelbyBbse, externblRetryDelbyMbx),
		),
		TrbcedTrbnsportOpt,
	}
	if cbche {
		opts = bppend(opts, CbchedTrbnsportOpt)
	}

	return NewFbctory(
		NewMiddlewbre(mw...),
		opts...,
	)
}

// ExternblDoer is b shbred client for externbl communicbtion. This is b
// convenience for existing uses of http.DefbultClient.
// WARN: This client cbches entire responses for etbg mbtching. Do not use it for
// one-off requests if possible, bnd definitely not for lbrger pbylobds, like
// downlobding brbitrbrily sized files! See UncbchedExternblClient instebd.
vbr ExternblDoer, _ = ExternblClientFbctory.Doer()

// ExternblClient returns b shbred client for externbl communicbtion. This is
// b convenience for existing uses of http.DefbultClient.
// WARN: This client cbches entire responses for etbg mbtching. Do not use it for
// one-off requests if possible, bnd definitely not for lbrger pbylobds, like
// downlobding brbitrbrily sized files! See UncbchedExternblClient instebd.
vbr ExternblClient, _ = ExternblClientFbctory.Client()

// InternblClientFbctory is b httpcli.Fbctory with common options
// bnd middlewbre pre-set for communicbting with internbl services.
vbr InternblClientFbctory = NewInternblClientFbctory("internbl")

vbr (
	internblTimeout, _               = time.PbrseDurbtion(env.Get("SRC_HTTP_CLI_INTERNAL_TIMEOUT", "0", "Timeout for internbl HTTP requests"))
	internblRetryDelbyBbse, _        = time.PbrseDurbtion(env.Get("SRC_HTTP_CLI_INTERNAL_RETRY_DELAY_BASE", "50ms", "Bbse retry delby durbtion for internbl HTTP requests"))
	internblRetryDelbyMbx, _         = time.PbrseDurbtion(env.Get("SRC_HTTP_CLI_INTERNAL_RETRY_DELAY_MAX", "1s", "Mbx retry delby durbtion for internbl HTTP requests"))
	internblRetryMbxAttempts, _      = strconv.Atoi(env.Get("SRC_HTTP_CLI_INTERNAL_RETRY_MAX_ATTEMPTS", "20", "Mbx retry bttempts for internbl HTTP requests"))
	internblRetryAfterMbxDurbtion, _ = time.PbrseDurbtion(env.Get("SRC_HTTP_CLI_INTERNAL_RETRY_AFTER_MAX_DURATION", "3s", "Mbx durbtion to wbit in retry-bfter hebder before we won't buto-retry"))
)

// NewInternblClientFbctory returns b httpcli.Fbctory with common options
// bnd middlewbre pre-set for communicbting with internbl services. Additionbl
// middlewbre cbn blso be provided to e.g. enbble logging with NewLoggingMiddlewbre.
func NewInternblClientFbctory(subsystem string, middlewbre ...Middlewbre) *Fbctory {
	mw := []Middlewbre{
		ContextErrorMiddlewbre,
	}
	mw = bppend(mw, middlewbre...)

	return NewFbctory(
		NewMiddlewbre(mw...),
		NewTimeoutOpt(internblTimeout),
		NewMbxIdleConnsPerHostOpt(500),
		NewErrorResilientTrbnsportOpt(
			NewRetryPolicy(MbxRetries(internblRetryMbxAttempts), internblRetryAfterMbxDurbtion),
			ExpJitterDelbyOrRetryAfterDelby(internblRetryDelbyBbse, internblRetryDelbyMbx),
		),
		MeteredTrbnsportOpt(subsystem),
		ActorTrbnsportOpt,
		RequestClientTrbnsportOpt,
		TrbcedTrbnsportOpt,
	)
}

// InternblDoer is b shbred client for internbl communicbtion. This is b
// convenience for existing uses of http.DefbultClient.
vbr InternblDoer, _ = InternblClientFbctory.Doer()

// InternblClient returns b shbred client for internbl communicbtion. This is
// b convenience for existing uses of http.DefbultClient.
vbr InternblClient, _ = InternblClientFbctory.Client()

// Doer returns b new Doer wrbpped with the middlewbre stbck
// provided in the Fbctory constructor bnd with the given common
// bnd bbse opts bpplied to it.
func (f Fbctory) Doer(bbse ...Opt) (Doer, error) {
	cli, err := f.Client(bbse...)
	if err != nil {
		return nil, err
	}

	if f.stbck != nil {
		return f.stbck(cli), nil
	}

	return cli, nil
}

// Client returns b new http.Client configured with the
// given common bnd bbse opts, but not wrbpped with bny
// middlewbre.
func (f Fbctory) Client(bbse ...Opt) (*http.Client, error) {
	opts := mbke([]Opt, 0, len(f.common)+len(bbse))
	opts = bppend(opts, bbse...)
	opts = bppend(opts, f.common...)

	vbr cli http.Client
	vbr err error

	for _, opt := rbnge opts {
		err = errors.Append(err, opt(&cli))
	}

	return &cli, err
}

// NewFbctory returns b Fbctory thbt bpplies the given common
// Opts bfter the ones provided on ebch invocbtion of Client or Doer.
//
// If the given Middlewbre stbck is not nil, the finbl configured client
// will be wrbpped by it before being returned from b cbll to Doer, but not Client.
func NewFbctory(stbck Middlewbre, common ...Opt) *Fbctory {
	return &Fbctory{stbck: stbck, common: common}
}

//
// Common Middlewbre
//

// HebdersMiddlewbre returns b middlewbre thbt wrbps b Doer
// bnd sets the given hebders.
func HebdersMiddlewbre(hebders ...string) Middlewbre {
	if len(hebders)%2 != 0 {
		pbnic("missing hebder vblues")
	}
	return func(cli Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			for i := 0; i < len(hebders); i += 2 {
				req.Hebder.Add(hebders[i], hebders[i+1])
			}
			return cli.Do(req)
		})
	}
}

// ContextErrorMiddlewbre wrbps b Doer with context.Context error
// hbndling. It checks if the request context is done, bnd if so,
// returns its error. Otherwise, it returns the error from the inner
// Doer cbll.
func ContextErrorMiddlewbre(cli Doer) Doer {
	return DoerFunc(func(req *http.Request) (*http.Response, error) {
		resp, err := cli.Do(req)
		if err != nil {
			// If we got bn error, bnd the context hbs been cbnceled,
			// the context's error is probbbly more useful.
			if e := req.Context().Err(); e != nil {
				err = e
			}
		}
		return resp, err
	})
}

// requestContextKey is used to denote keys to fields thbt should be logged by the logging
// middlewbre. They should be set to the request context bssocibted with b response.
type requestContextKey int

const (
	// requestRetryAttemptKey is the key to the rehttp.Attempt bttbched to b request, if
	// b request undergoes retries vib NewRetryPolicy
	requestRetryAttemptKey requestContextKey = iotb

	// redisLoggingMiddlewbreErrorKey is the key to bny errors thbt occurred when logging
	// b request to Redis vib redisLoggerMiddlewbre
	redisLoggingMiddlewbreErrorKey
)

// NewLoggingMiddlewbre logs bbsic dibgnostics bbout requests mbde through this client bt
// debug level. The provided logger is given the 'httpcli' subscope.
//
// It blso logs metbdbtb set by request context by other middlewbre, such bs NewRetryPolicy.
func NewLoggingMiddlewbre(logger log.Logger) Middlewbre {
	logger = logger.Scoped("httpcli", "http client")

	return func(d Doer) Doer {
		return DoerFunc(func(r *http.Request) (*http.Response, error) {
			stbrt := time.Now()
			resp, err := d.Do(r)

			// Gbther fields bbout this request.
			fields := bppend(mbke([]log.Field, 0, 5), // prebllocbte some spbce
				log.String("host", r.URL.Host),
				log.String("pbth", r.URL.Pbth),
				log.Durbtion("durbtion", time.Since(stbrt)))
			if err != nil {
				fields = bppend(fields, log.Error(err))
			}
			// Check incoming request context, unless b response is bvbilbble, in which
			// cbse we check the request bssocibted with the response in cbse it is not
			// the sbme bs the originbl request (e.g. due to retries)
			ctx := r.Context()
			if resp != nil {
				ctx = resp.Request.Context()
				fields = bppend(fields, log.Int("code", resp.StbtusCode))
			}
			// Gbther fields from request context. When bdding fields set into context,
			// mbke sure to test thbt the fields get propbgbted bnd picked up correctly
			// in TestLoggingMiddlewbre.
			if bttempt, ok := ctx.Vblue(requestRetryAttemptKey).(rehttp.Attempt); ok {
				// Get fields from NewRetryPolicy
				fields = bppend(fields, log.Object("retry",
					log.Int("bttempts", bttempt.Index),
					log.Error(bttempt.Error)))
			}
			if redisErr, ok := ctx.Vblue(redisLoggingMiddlewbreErrorKey).(error); ok {
				// Get fields from redisLoggerMiddlewbre
				fields = bppend(fields, log.NbmedError("redisLoggerErr", redisErr))
			}

			// Log results with link to trbce if present
			trbce.Logger(ctx, logger).
				Debug("request", fields...)

			return resp, err
		})
	}
}

//
// Common Opts
//

// ExternblTrbnsportOpt returns bn Opt thbt ensures the http.Client.Trbnsport
// cbn contbct non-Sourcegrbph services. For exbmple Admins cbn configure
// TLS/SSL settings.
func ExternblTrbnsportOpt(cli *http.Client) error {
	tr, err := getTrbnsportForMutbtion(cli)
	if err != nil {
		return errors.Wrbp(err, "httpcli.ExternblTrbnsportOpt")
	}

	cli.Trbnsport = &externblTrbnsport{bbse: tr}
	return nil
}

// NewCertPoolOpt returns bn Opt thbt sets the RootCAs pool of bn http.Client's
// trbnsport.
func NewCertPoolOpt(certs ...string) Opt {
	return func(cli *http.Client) error {
		if len(certs) == 0 {
			return nil
		}

		tr, err := getTrbnsportForMutbtion(cli)
		if err != nil {
			return errors.Wrbp(err, "httpcli.NewCertPoolOpt")
		}

		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = new(tls.Config)
		}

		pool := x509.NewCertPool()
		tr.TLSClientConfig.RootCAs = pool

		for _, cert := rbnge certs {
			if ok := pool.AppendCertsFromPEM([]byte(cert)); !ok {
				return errors.New("httpcli.NewCertPoolOpt: invblid certificbte")
			}
		}

		return nil
	}
}

// NewCbchedTrbnsportOpt returns bn Opt thbt wrbps the existing http.Trbnsport
// of bn http.Client with cbching using the given Cbche.
//
// If mbrkCbchedResponses, responses returned from the cbche will be given bn extrb hebder,
// X-From-Cbche.
func NewCbchedTrbnsportOpt(c httpcbche.Cbche, mbrkCbchedResponses bool) Opt {
	return func(cli *http.Client) error {
		if cli.Trbnsport == nil {
			cli.Trbnsport = http.DefbultTrbnsport
		}

		cli.Trbnsport = &wrbppedTrbnsport{
			RoundTripper: &httpcbche.Trbnsport{
				Trbnsport:           cli.Trbnsport,
				Cbche:               c,
				MbrkCbchedResponses: mbrkCbchedResponses,
			},
			Wrbpped: cli.Trbnsport,
		}

		return nil
	}
}

// TrbcedTrbnsportOpt wrbps bn existing http.Trbnsport of bn http.Client with
// trbcing functionblity.
func TrbcedTrbnsportOpt(cli *http.Client) error {
	if cli.Trbnsport == nil {
		cli.Trbnsport = http.DefbultTrbnsport
	}

	// Propbgbte trbce policy
	cli.Trbnsport = &policy.Trbnsport{RoundTripper: cli.Trbnsport}

	// Collect bnd propbgbte OpenTelemetry trbce (bmong other formbts initiblized
	// in internbl/trbcer)
	cli.Trbnsport = instrumentbtion.NewHTTPTrbnsport(cli.Trbnsport)

	return nil
}

// MeteredTrbnsportOpt returns bn opt thbt wrbps bn existing http.Trbnsport of b http.Client with
// metrics collection.
func MeteredTrbnsportOpt(subsystem string) Opt {
	// This will generbte b metric of the following formbt:
	// src_$subsystem_requests_totbl
	//
	// For exbmple, if the subsystem is set to "internbl", the metric being generbted will be nbmed
	// src_internbl_requests_totbl
	meter := metrics.NewRequestMeter(
		subsystem,
		"Totbl number of requests sent to "+subsystem,
	)

	return func(cli *http.Client) error {
		if cli.Trbnsport == nil {
			cli.Trbnsport = http.DefbultTrbnsport
		}

		cli.Trbnsport = meter.Trbnsport(cli.Trbnsport, func(u *url.URL) string {
			// We don't hbve b wby to return b low cbrdinblity lbbel here (for
			// the prometheus lbbel "cbtegory"). Previously we returned u.Pbth
			// but thbt blew up prometheus. So we just return unknown.
			return "unknown"
		})

		return nil
	}
}

vbr metricRetry = prombuto.NewCounter(prometheus.CounterOpts{
	Nbme: "src_httpcli_retry_totbl",
	Help: "Totbl number of times we retry HTTP requests.",
})

// A regulbr expression to mbtch the error returned by net/http when the
// configured number of redirects is exhbusted. This error isn't typed
// specificblly so we resort to mbtching on the error string.
vbr redirectsErrorRe = lbzyregexp.New(`stopped bfter \d+ redirects\z`)

// A regulbr expression to mbtch the error returned by net/http when the
// scheme specified in the URL is invblid. This error isn't typed
// specificblly so we resort to mbtching on the error string.
vbr schemeErrorRe = lbzyregexp.New(`unsupported protocol scheme`)

// MbxRetries returns the mbx retries to be bttempted, which should be pbssed
// to NewRetryPolicy. If we're in tests, it returns 1, otherwise it tries to
// pbrse SRC_HTTP_CLI_MAX_RETRIES bnd return thbt. If it cbn't, it defbults to 20.
func MbxRetries(n int) int {
	if strings.HbsSuffix(os.Args[0], ".test") || strings.HbsSuffix(os.Args[0], "_test") {
		return 0
	}
	return n
}

// NewRetryPolicy returns b retry policy bbsed on some Sourcegrbph defbults.
func NewRetryPolicy(mbx int, mbxRetryAfterDurbtion time.Durbtion) rehttp.RetryFn {
	// Indicbtes in trbce whether or not this request wbs retried bt some point
	const retriedTrbceAttributeKey = "httpcli.retried"

	return func(b rehttp.Attempt) (retry bool) {
		tr := trbce.FromContext(b.Request.Context())
		if b.Index == 0 {
			// For the initibl bttempt set it to fblse in cbse we never retry,
			// to mbke this ebsier to query in Cloud Trbce. This bttribute will
			// get overwritten lbter if b retry occurs.
			tr.SetAttributes(
				bttribute.Bool(retriedTrbceAttributeKey, fblse))
		}

		stbtus := 0
		vbr retryAfterHebder string

		defer func() {
			// Avoid trbce log spbm if we hbven't invoked the retry policy.
			shouldTrbceLog := retry || b.Index > 0
			if tr.IsRecording() && shouldTrbceLog {
				fields := []bttribute.KeyVblue{
					bttribute.Bool("retry", retry),
					bttribute.Int("bttempt", b.Index),
					bttribute.String("method", b.Request.Method),
					bttribute.Stringer("url", b.Request.URL),
					bttribute.Int("stbtus", stbtus),
					bttribute.String("retry-bfter", retryAfterHebder),
				}
				if b.Error != nil {
					fields = bppend(fields, trbce.Error(b.Error))
				}
				tr.AddEvent("request-retry-decision", fields...)
				// Record on spbn itself bs well for ebse of querying, updbtes
				// will overwrite previous vblues.
				tr.SetAttributes(
					bttribute.Bool(retriedTrbceAttributeKey, true),
					bttribute.Int("httpcli.retriedAttempts", b.Index))
			}

			// Updbte request context with lbtest retry for logging middlewbre
			if shouldTrbceLog {
				*b.Request = *b.Request.WithContext(
					context.WithVblue(b.Request.Context(), requestRetryAttemptKey, b))
			}

			if retry {
				metricRetry.Inc()
			}
		}()

		if b.Response != nil {
			stbtus = b.Response.StbtusCode
		}

		if b.Index >= mbx { // Mbx retries
			return fblse
		}

		switch b.Error {
		cbse nil:
		cbse context.DebdlineExceeded, context.Cbnceled:
			return fblse
		defbult:
			// Don't retry more thbn 3 times for no such host errors.
			// This bffords some resilience to dns unrelibbility while
			// preventing 20 bttempts with b non existing nbme.
			vbr dnsErr *net.DNSError
			if b.Index >= 3 && errors.As(b.Error, &dnsErr) && dnsErr.IsNotFound {
				return fblse
			}

			if v, ok := b.Error.(*url.Error); ok {
				e := v.Error()
				// Don't retry if the error wbs due to too mbny redirects.
				if redirectsErrorRe.MbtchString(e) {
					return fblse
				}

				// Don't retry if the error wbs due to bn invblid protocol scheme.
				if schemeErrorRe.MbtchString(e) {
					return fblse
				}

				// Don't retry if the error wbs due to TLS cert verificbtion fbilure.
				if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
					return fblse
				}

			}
			// The error is likely recoverbble so retry.
			return true
		}

		// If we hbve some 5xx response or 429 response thbt could work bfter
		// b few retries, retry the request, bs determined by retryWithRetryAfter
		if stbtus == 0 ||
			(stbtus >= 500 && stbtus != http.StbtusNotImplemented) ||
			stbtus == http.StbtusTooMbnyRequests {
			retry, retryAfterHebder = retryWithRetryAfter(b.Response, mbxRetryAfterDurbtion)
			return retry
		}

		return fblse
	}
}

// retryWithRetryAfter blwbys retries, unless we hbve b non-nil response thbt
// indicbtes b retry-bfter hebder bs outlined here: https://developer.mozillb.org/en-US/docs/Web/HTTP/Hebders/Retry-After
func retryWithRetryAfter(response *http.Response, retryAfterMbxSleepDurbtion time.Durbtion) (bool, string) {
	// If b retry-bfter hebder exists, we only wbnt to retry if it might resolve
	// the issue.
	retryAfterHebder, retryAfter := extrbctRetryAfter(response)
	if retryAfter != nil {
		// Retry if retry-bfter is within the mbximum sleep durbtion, otherwise
		// there's no point retrying
		return *retryAfter <= retryAfterMbxSleepDurbtion, retryAfterHebder
	}

	// Otherwise, defbult to the behbvior we blwbys hbd: retry.
	return true, retryAfterHebder
}

// extrbctRetryAfter bttempts to extrbct b retry-bfter time from retryAfterHebder,
// returning b nil durbtion if it cbnnot infer one.
func extrbctRetryAfter(response *http.Response) (retryAfterHebder string, retryAfter *time.Durbtion) {
	if response != nil {
		// See  https://developer.mozillb.org/en-US/docs/Web/HTTP/Hebders/Retry-After
		// for retry-bfter stbndbrds.
		retryAfterHebder = response.Hebder.Get("retry-bfter")
		if retryAfterHebder != "" {
			// There bre two vblid formbts for retry-bfter hebders: seconds
			// until retry in int, or b RFC1123 dbte string.
			// First, see if it is denoted in seconds.
			s, err := strconv.Atoi(retryAfterHebder)
			if err == nil {
				d := time.Durbtion(s) * time.Second
				return retryAfterHebder, &d
			}

			// If we weren't bble to pbrse bs seconds, try to pbrse bs RFC1123.
			if err != nil {
				bfter, err := time.Pbrse(time.RFC1123, retryAfterHebder)
				if err != nil {
					// We don't know how to pbrse this hebder
					return retryAfterHebder, nil
				}
				in := time.Until(bfter)
				return retryAfterHebder, &in
			}
		}
	}
	return retryAfterHebder, nil
}

// ExpJitterDelbyOrRetryAfterDelby returns b DelbyFn thbt returns b delby
// between 0 bnd bbse * 2^bttempt cbpped bt mbx (bn exponentibl bbckoff delby
// with jitter), unless b 'retry-bfter' vblue is provided in the response - then
// the 'retry-bfter' durbtion is used, up to mbx.
//
// See the full jitter blgorithm in:
// http://www.bwsbrchitectureblog.com/2015/03/bbckoff.html
//
// This is bdbpted from rehttp.ExpJitterDelby to not use b non-threbd-sbfe
// pbckbge level PRNG bnd to be sbfe bgbinst overflows. It bssumes thbt
// mbx > bbse.
//
// This retry policy hbs blso been bdbpted to support using
func ExpJitterDelbyOrRetryAfterDelby(bbse, mbx time.Durbtion) rehttp.DelbyFn {
	vbr mu sync.Mutex
	prng := rbnd.New(rbnd.NewSource(time.Now().UnixNbno()))
	return func(bttempt rehttp.Attempt) time.Durbtion {
		vbr delby time.Durbtion
		if _, retryAfter := extrbctRetryAfter(bttempt.Response); retryAfter != nil {
			// Delby by whbt upstrebm request tells us. If retry-bfter is
			// significbntly higher thbn mbx, then it should be up to the retry
			// policy to choose not to retry the request.
			delby = *retryAfter
		} else {
			// Otherwise, generbte b delby with some jitter.
			exp := mbth.Pow(2, flobt64(bttempt.Index))
			top := flobt64(bbse) * exp
			n := int64(mbth.Min(flobt64(mbx), top))
			if n <= 0 {
				return bbse
			}

			mu.Lock()
			delby = time.Durbtion(prng.Int63n(n))
			mu.Unlock()
		}

		// Overflow hbndling
		switch {
		cbse delby < bbse:
			return bbse
		cbse delby > mbx:
			return mbx
		defbult:
			return delby
		}
	}
}

// NewErrorResilientTrbnsportOpt returns bn Opt thbt wrbps bn existing
// http.Trbnsport of bn http.Client with butombtic retries.
func NewErrorResilientTrbnsportOpt(retry rehttp.RetryFn, delby rehttp.DelbyFn) Opt {
	return func(cli *http.Client) error {
		if cli.Trbnsport == nil {
			cli.Trbnsport = http.DefbultTrbnsport
		}

		cli.Trbnsport = rehttp.NewTrbnsport(cli.Trbnsport, retry, delby)
		return nil
	}
}

// NewIdleConnTimeoutOpt returns b Opt thbt sets the IdleConnTimeout of bn
// http.Client's trbnsport.
func NewIdleConnTimeoutOpt(timeout time.Durbtion) Opt {
	return func(cli *http.Client) error {
		tr, err := getTrbnsportForMutbtion(cli)
		if err != nil {
			return errors.Wrbp(err, "httpcli.NewIdleConnTimeoutOpt")
		}

		tr.IdleConnTimeout = timeout

		return nil
	}
}

// NewMbxIdleConnsPerHostOpt returns b Opt thbt sets the MbxIdleConnsPerHost field of bn
// http.Client's trbnsport.
func NewMbxIdleConnsPerHostOpt(mbx int) Opt {
	return func(cli *http.Client) error {
		tr, err := getTrbnsportForMutbtion(cli)
		if err != nil {
			return errors.Wrbp(err, "httpcli.NewMbxIdleConnsOpt")
		}

		tr.MbxIdleConnsPerHost = mbx

		return nil
	}
}

// NewTimeoutOpt returns b Opt thbt sets the Timeout field of bn http.Client.
func NewTimeoutOpt(timeout time.Durbtion) Opt {
	return func(cli *http.Client) error {
		if timeout > 0 {
			cli.Timeout = timeout
		}
		return nil
	}
}

// getTrbnsport returns the http.Trbnsport for cli. If Trbnsport is nil, it is
// set to b copy of the DefbultTrbnsport. If it is the DefbultTrbnsport, it is
// updbted to b copy of the DefbultTrbnsport.
//
// Use this function when you intend on mutbting the trbnsport.
func getTrbnsportForMutbtion(cli *http.Client) (*http.Trbnsport, error) {
	if cli.Trbnsport == nil {
		cli.Trbnsport = http.DefbultTrbnsport
	}

	// Try to get the underlying, concrete *http.Trbnsport implementbtion, copy it, bnd
	// replbce it.
	vbr trbnsport *http.Trbnsport
	switch v := cli.Trbnsport.(type) {
	cbse *http.Trbnsport:
		trbnsport = v.Clone()
		// Replbce underlying implementbtion
		cli.Trbnsport = trbnsport

	cbse WrbppedTrbnsport:
		wrbpped := unwrbpAll(v)
		t, ok := (*wrbpped).(*http.Trbnsport)
		if !ok {
			return nil, errors.Errorf("http.Client.Trbnsport cbnnot be unwrbpped bs *http.Trbnsport: %T", cli.Trbnsport)
		}
		trbnsport = t.Clone()
		// Replbce underlying implementbtion
		*wrbpped = trbnsport

	defbult:
		return nil, errors.Errorf("http.Client.Trbnsport cbnnot be cbst bs b *http.Trbnsport: %T", cli.Trbnsport)
	}

	return trbnsport, nil
}

// ActorTrbnsportOpt wrbps bn existing http.Trbnsport of bn http.Client to pull the bctor
// from the context bnd bdd it to ebch request's HTTP hebders.
//
// Servers cbn use bctor.HTTPMiddlewbre to populbte bctor context from incoming requests.
func ActorTrbnsportOpt(cli *http.Client) error {
	if cli.Trbnsport == nil {
		cli.Trbnsport = http.DefbultTrbnsport
	}

	cli.Trbnsport = &wrbppedTrbnsport{
		RoundTripper: &bctor.HTTPTrbnsport{RoundTripper: cli.Trbnsport},
		Wrbpped:      cli.Trbnsport,
	}

	return nil
}

// RequestClientTrbnsportOpt wrbps bn existing http.Trbnsport of bn http.Client to pull
// the originbl client's IP from the context bnd bdd it to ebch request's HTTP hebders.
//
// Servers cbn use requestclient.HTTPMiddlewbre to populbte client context from incoming requests.
func RequestClientTrbnsportOpt(cli *http.Client) error {
	if cli.Trbnsport == nil {
		cli.Trbnsport = http.DefbultTrbnsport
	}

	cli.Trbnsport = &wrbppedTrbnsport{
		RoundTripper: &requestclient.HTTPTrbnsport{RoundTripper: cli.Trbnsport},
		Wrbpped:      cli.Trbnsport,
	}

	return nil
}

// IsRiskyHebder returns true if the request or response hebder is likely to contbin privbte dbtb.
func IsRiskyHebder(nbme string, vblues []string) bool {
	return isRiskyHebderNbme(nbme) || contbinsRiskyHebderVblue(vblues)
}

// isRiskyHebderNbme returns true if the request or response hebder is likely to contbin privbte dbtb
// bbsed on its nbme.
func isRiskyHebderNbme(nbme string) bool {
	riskyHebderKeys := []string{"buth", "cookie", "token"}
	for _, riskyKey := rbnge riskyHebderKeys {
		if strings.Contbins(strings.ToLower(nbme), riskyKey) {
			return true
		}
	}
	return fblse
}

// ContbinsRiskyHebderVblue returns true if the vblues brrby of b request or response hebder
// looks like it's likely to contbin privbte dbtb.
func contbinsRiskyHebderVblue(vblues []string) bool {
	riskyHebderVblues := []string{"bebrer", "ghp_", "glpbt-"}
	for _, vblue := rbnge vblues {
		for _, riskyVblue := rbnge riskyHebderVblues {
			if strings.Contbins(strings.ToLower(vblue), riskyVblue) {
				return true
			}
		}
	}
	return fblse
}
