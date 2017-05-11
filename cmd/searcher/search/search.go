// search is a service which exposes an API to text search a repo at a
// specific commit.
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

// Params are the input for a search request. Most of the fields are based on
// PatternInfo used in vscode.
type Params struct {
	// Repo is which repository to search. eg "github.com/gorilla/mux"
	Repo string

	// Commit is which commit to search. It is required to be resolved,
	// not a ref like HEAD or master. eg
	// "599cba5e7b6137d46ddf58fb1765f5d928e69604"
	Commit string

	// Pattern is the search query. It is a regular expression if IsRegExp
	// is true, otherwise a fixed string. eg "route variable"
	Pattern string

	// IsRegExp if true will treat the Pattern as a regular expression.
	IsRegExp bool

	// IsWordMatch if true will only match the pattern at word boundaries.
	IsWordMatch bool

	// IsCaseSensitive if false will ignore the case of text and pattern
	// when finding matches.
	IsCaseSensitive bool

	// ExcludePattern is a glob pattern that should not match the returned files.
	// eg '**/node_modules'
	ExcludePattern string

	// Include pattern is a glob pattern that should match the returned files.
	// eg '**/*.go' to search go files.
	IncludePattern string

	// FileMatchLimit limits the number of files with matches that are returned.
	FileMatchLimit int
}

// Response represents the response from a Search request.
type Response struct {
	Matches []FileMatch

	// LimitHit is true if Matches may not include all FileMatches.
	LimitHit bool
}

// FileMatch is the struct used by vscode to receive search results
type FileMatch struct {
	Path        string
	LineMatches []LineMatch

	// LimitHit is true if LineMatches may not include all LineMatches.
	LimitHit bool
}

// LineMatch is the struct used by vscode to receive search results for a line.
type LineMatch struct {
	// Preview is the matched line.
	Preview string

	// LineNumber is the 0-based line number. Note: Our editors present
	// 1-based line numbers, but internally vscode uses 0-based.
	LineNumber int

	// OffsetAndLengths is a slice of 2-tuples (Offset, Length)
	// representing each match on a line.
	OffsetAndLengths [][]int

	// LimitHit is true if OffsetAndLengths may not include all OffsetAndLengths.
	LimitHit bool
}

// ServeHTTP handles HTTP based search requests
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	running.Inc()
	defer running.Dec()

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	var p Params
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
		if !isBadRequest(err) {
			log.Printf("internal error serving %#+v: %s", p, err)
			code = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), code)
		return
	}
	if matches == nil {
		// Return an empty list
		matches = make([]FileMatch, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	resp := Response{
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

func (s *Service) search(ctx context.Context, p *Params) (matches []FileMatch, limitHit bool, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Search")
	ext.Component.Set(span, "service")
	span.SetTag("repo", p.Repo)
	span.SetTag("commit", p.Commit)
	span.SetTag("pattern", p.Pattern)
	span.SetTag("isRegExp", strconv.FormatBool(p.IsRegExp))
	span.SetTag("isWordMatch", strconv.FormatBool(p.IsWordMatch))
	span.SetTag("isCaseSensitive", strconv.FormatBool(p.IsCaseSensitive))
	span.SetTag("fileMatchLimit", p.FileMatchLimit)
	defer func(start time.Time) {
		code := "200"
		if err != nil {
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

	rg, err := compile(p)
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

func validateParams(p *Params) error {
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
