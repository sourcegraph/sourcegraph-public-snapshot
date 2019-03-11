package httpcli

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"

	"github.com/gregjones/httpcache"
	"github.com/hashicorp/go-multierror"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
)

// A Doer captures the Do method of an http.Client. It faciliates decorating
// an http.Client with orthogonal concerns such as logging, metrics, retries,
// etc.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// DoerFunc is function adapter that implements the http.RoundTripper
// interface by calling itself.
type DoerFunc func(*http.Request) (*http.Response, error)

// Do implements the Doer interface.
func (f DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// A Middleware function wraps a Doer with a layer of behaviour. It's used
// to decorate an http.Client with orthogonal layers of behaviour such as
// logging, instrumentation, retries, etc.
type Middleware func(Doer) Doer

// NewMiddleware returns a Middleware stack composed of the given Middlewares.
func NewMiddleware(mws ...Middleware) Middleware {
	return func(bottom Doer) (stacked Doer) {
		stacked = bottom
		for _, mw := range mws {
			stacked = mw(stacked)
		}
		return stacked
	}
}

// A Opt configures an aspect of a given *http.Client,
// returning an error in case of failure.
type Opt func(*http.Client) error

// A Factory constructs an http.Client with the given functional
// options applied, returning an aggreagte error of the errors returned by
// all those options.
type Factory func(...Opt) (Doer, error)

// NewClient returns a new http.Client from the factory with the given opts
// applied to it.
func (f Factory) NewClient(opts ...Opt) (Doer, error) {
	return f(opts...)
}

// NewFactory returns a Factory that applies the given common
// Opts after the ones provided on each invocation of New.
//
// If the given Middleware stack is not nil, the final configured client
// will be wrapped by it before being returned.
func NewFactory(stack Middleware, common ...Opt) Factory {
	return func(base ...Opt) (do Doer, _ error) {
		opts := make([]Opt, 0, len(common)+len(base))
		opts = append(opts, base...)
		opts = append(opts, common...)

		var cli http.Client
		var err *multierror.Error

		for _, opt := range opts {
			err = multierror.Append(err, opt(&cli))
		}

		do = &cli
		if stack != nil {
			do = stack(do)
		}

		return do, err.ErrorOrNil()
	}
}

//
// Common Opts
//

// NewCertPoolOpt returns a Opt that sets the RootCAs pool of an http.Client's
// transport.
func NewCertPoolOpt(pool *x509.CertPool) Opt {
	return func(cli *http.Client) error {
		tr, ok := cli.Transport.(*http.Transport)
		if !ok {
			return errors.New("httpcli.NewCertPoolOpt: http.Client.Transport is not an *http.Transport")
		}

		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = new(tls.Config)
		}

		tr.TLSClientConfig.RootCAs = pool

		return nil
	}
}

// NewCachedTransportOpt returns an Opt that wraps the existing http.Transport
// of an http.Client with caching using the given Cache.
func NewCachedTransportOpt(c httpcache.Cache, markCachedResponses bool) Opt {
	return func(cli *http.Client) error {
		if cli.Transport == nil {
			cli.Transport = http.DefaultTransport
		}

		cli.Transport = &httpcache.Transport{
			Transport:           cli.Transport,
			Cache:               c,
			MarkCachedResponses: markCachedResponses,
		}

		return nil
	}
}

// TracedTransportOpt wraps an existing http.Transport of an http.Client with
// tracing functionality.
func TracedTransportOpt(cli *http.Client) error {
	if cli.Transport == nil {
		cli.Transport = http.DefaultTransport
	}

	cli.Transport = &nethttp.Transport{RoundTripper: cli.Transport}
	return nil
}
