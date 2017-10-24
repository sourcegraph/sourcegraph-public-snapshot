// Package search is a service which exposes an API to text search a repo at
// a specific commit.
//
// Architecture Notes:
// * Archive is fetched from gitserver
// * Simple HTTP API exposed
// * Currently no concept of authorization
// * On disk cache of fetched archives to reduce load on gitserver
// * Run search on archive. Rely on OS file buffers
// * Simple to scale up since stateless
// * Use ingress with affinity to increase local cache hit ratio
package search

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searcher/protocol"

	"github.com/pkg/errors"

	"github.com/gorilla/schema"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
)

// Service is the search service. It is an http.Handler.
type Service struct {
	Store *Store

	// RequestLog if non-nil will log info per valid search request.
	RequestLog *log.Logger
}

var decoder = schema.NewDecoder()

// ServeHTTP handles HTTP based search requests
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	running.Inc()
	defer running.Dec()

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	var p protocol.Request
	err = decoder.Decode(&p, r.Form)
	if err != nil {
		http.Error(w, "failed to decode form: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err = validateParams(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	matches, limitHit, err := s.search(r.Context(), &p)
	if err != nil {
		code := http.StatusBadRequest
		// Log errors not caused by the client.
		if !isBadRequest(err) && r.Context().Err() != context.Canceled {
			log.Printf("internal error serving %#+v: %s", p, err)
			code = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), code)
		return
	}
	if matches == nil {
		// Return an empty list
		matches = make([]protocol.FileMatch, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	resp := protocol.Response{
		Matches:  matches,
		LimitHit: limitHit,
	}
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		// We may have already started writing to w
		http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) search(ctx context.Context, p *protocol.Request) (matches []protocol.FileMatch, limitHit bool, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Search")
	ext.Component.Set(span, "service")
	span.SetTag("repo", p.Repo)
	span.SetTag("commit", p.Commit)
	span.SetTag("pattern", p.Pattern)
	span.SetTag("isRegExp", strconv.FormatBool(p.IsRegExp))
	span.SetTag("isWordMatch", strconv.FormatBool(p.IsWordMatch))
	span.SetTag("isCaseSensitive", strconv.FormatBool(p.IsCaseSensitive))
	span.SetTag("pathPatternsAreRegExps", strconv.FormatBool(p.PathPatternsAreRegExps))
	span.SetTag("pathPatternsAreCaseSensitive", strconv.FormatBool(p.PathPatternsAreCaseSensitive))
	span.SetTag("fileMatchLimit", p.FileMatchLimit)
	defer func(start time.Time) {
		code := "200"
		// We often have canceled requests. We do not want to
		// record them as errors to avoid noise
		if ctx.Err() == context.Canceled {
			code = "canceled"
			span.SetTag("err", err)
		} else if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
			if isBadRequest(err) {
				code = "400"
			} else {
				code = "500"
			}
		}
		requestTotal.WithLabelValues(code).Inc()
		span.SetTag("matches", len(matches))
		span.SetTag("limitHit", limitHit)
		span.Finish()
		if s.RequestLog != nil {
			errS := ""
			if err != nil {
				errS = " error=" + strconv.Quote(err.Error())
			}
			s.RequestLog.Printf("search request repo=%v commit=%v pattern=%q isRegExp=%v isWordMatch=%v isCaseSensitive=%v matches=%d duration=%v%s", p.Repo, p.Commit, p.Pattern, p.IsRegExp, p.IsWordMatch, p.IsCaseSensitive, len(matches), time.Since(start), errS)
		}
	}(time.Now())

	rg, err := compile(&p.PatternInfo)
	if err != nil {
		return nil, false, badRequestError{err.Error()}
	}

	ar, err := s.Store.openReader(ctx, p.Repo, p.Commit)
	if err != nil {
		return nil, false, err
	}
	defer ar.Close()

	return concurrentFind(ctx, rg, ar.Reader(), p.FileMatchLimit)
}

func validateParams(p *protocol.Request) error {
	if p.Repo == "" {
		return errors.New("Repo must be non-empty")
	}
	// Surprisingly this is the same sanity check used in the git source.
	if len(p.Commit) != 40 {
		return errors.Errorf("Commit must be resolved (Commit=%q)", p.Commit)
	}
	if p.Pattern == "" {
		return errors.New("Pattern must be non-empty")
	}
	return nil
}

var (
	running = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "searcher",
		Subsystem: "service",
		Name:      "running",
		Help:      "Number of running search requests.",
	})
	requestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "searcher",
		Subsystem: "service",
		Name:      "request_total",
		Help:      "Number of returned search requests.",
	}, []string{"code"})
)

func init() {
	prometheus.MustRegister(running)
	prometheus.MustRegister(requestTotal)
}

type badRequestError struct{ msg string }

func (e badRequestError) Error() string    { return e.msg }
func (e badRequestError) BadRequest() bool { return true }

func isBadRequest(err error) bool {
	e, ok := errors.Cause(err).(interface {
		BadRequest() bool
	})
	return ok && e.BadRequest()
}
