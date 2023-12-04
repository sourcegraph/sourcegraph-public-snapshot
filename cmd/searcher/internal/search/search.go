// Package search is a service which exposes an API to text search a repo at
// a specific commit.
//
// Architecture Notes:
//   - Archive is fetched from gitserver
//   - Simple HTTP API exposed
//   - Currently no concept of authorization
//   - On disk cache of fetched archives to reduce load on gitserver
//   - Run search on archive. Rely on OS file buffers
//   - Simple to scale up since stateless
//   - Use ingress with affinity to increase local cache hit ratio
package search

import (
	"context"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	// numWorkers is how many concurrent readerGreps run in the case of
	// regexSearch, and the number of parallel workers in the case of
	// structuralSearch.
	numWorkers = 8
)

// Service is the search service. It is an http.Handler.
type Service struct {
	Store *Store
	Log   log.Logger

	Indexed zoekt.Streamer

	// GitDiffSymbols returns the stdout of running "git diff -z --name-status
	// --no-renames commitA commitB" against repo.
	//
	// TODO Git client should be exposing a better API here.
	GitDiffSymbols func(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) ([]byte, error)

	// MaxTotalPathsLength is the maximum sum of lengths of all paths in a
	// single call to git archive. This mainly needs to be less than ARG_MAX
	// for the exec.Command on gitserver.
	MaxTotalPathsLength int
}

// ServeHTTP handles HTTP based search requests
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var p protocol.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		http.Error(w, "failed to decode form: "+err.Error(), http.StatusBadRequest)
		return
	}

	if !p.PatternMatchesContent && !p.PatternMatchesPath {
		// BACKCOMPAT: Old frontends send neither of these fields, but we still want to
		// search file content in that case.
		p.PatternMatchesContent = true
	}
	if err := validateParams(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.streamSearch(ctx, w, p)
}

// isNetOpError returns true if net.OpError is contained in err. This is
// useful to ignore errors when the connection has gone away.
func isNetOpError(err error) bool {
	return errors.HasType(err, (*net.OpError)(nil))
}

func (s *Service) streamSearch(ctx context.Context, w http.ResponseWriter, p protocol.Request) {
	if p.Limit == 0 {
		// No limit for streaming search since upstream limits
		// will either be sent in the request, or propagated by
		// a cancelled context.
		p.Limit = math.MaxInt32
	}
	eventWriter, err := streamhttp.NewWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var bufMux sync.Mutex
	matchesBuf := streamhttp.NewJSONArrayBuf(32*1024, func(data []byte) error {
		return eventWriter.EventBytes("matches", data)
	})
	onMatches := func(match protocol.FileMatch) {
		bufMux.Lock()
		if err := matchesBuf.Append(match); err != nil && !isNetOpError(err) {
			s.Log.Warn("failed appending match to buffer", log.Error(err))
		}
		bufMux.Unlock()
	}

	ctx, cancel, stream := newLimitedStream(ctx, p.Limit, onMatches)
	defer cancel()

	err = s.search(ctx, &p, stream)
	doneEvent := searcher.EventDone{
		LimitHit: stream.LimitHit(),
	}
	if err != nil {
		doneEvent.Error = err.Error()
	}

	// Flush remaining matches before sending a different event
	if err := matchesBuf.Flush(); err != nil && !isNetOpError(err) {
		s.Log.Warn("failed to flush matches", log.Error(err))
	}
	if err := eventWriter.Event("done", doneEvent); err != nil && !isNetOpError(err) {
		s.Log.Warn("failed to send done event", log.Error(err))
	}
}

func (s *Service) search(ctx context.Context, p *protocol.Request, sender matchSender) (err error) {
	metricRunning.Inc()
	defer metricRunning.Dec()

	var tr trace.Trace
	tr, ctx = trace.New(ctx, "search",
		p.Repo.Attr(),
		p.Commit.Attr(),
		attribute.String("url", p.URL),
		attribute.String("pattern", p.Pattern),
		attribute.Bool("isRegExp", p.IsRegExp),
		attribute.StringSlice("languages", p.Languages),
		attribute.Bool("isWordMatch", p.IsWordMatch),
		attribute.Bool("isCaseSensitive", p.IsCaseSensitive),
		attribute.Bool("pathPatternsAreCaseSensitive", p.PathPatternsAreCaseSensitive),
		attribute.Int("limit", p.Limit),
		attribute.Bool("patternMatchesContent", p.PatternMatchesContent),
		attribute.Bool("patternMatchesPath", p.PatternMatchesPath),
		attribute.String("select", p.Select))
	defer tr.End()
	defer func(start time.Time) {
		code := "200"
		// We often have canceled and timed out requests. We do not want to
		// record them as errors to avoid noise
		if ctx.Err() == context.Canceled {
			code = "canceled"
			tr.SetError(err)
		} else if err != nil {
			tr.SetError(err)
			if errcode.IsBadRequest(err) {
				code = "400"
			} else if errcode.IsTemporary(err) {
				code = "503"
			} else {
				code = "500"
			}
		}
		metricRequestTotal.WithLabelValues(code).Inc()
		tr.AddEvent("done",
			attribute.String("code", code),
			attribute.Int("matches.len", sender.SentCount()),
			attribute.Bool("limitHit", sender.LimitHit()),
		)
		s.Log.Debug("search request",
			log.String("repo", string(p.Repo)),
			log.String("commit", string(p.Commit)),
			log.String("pattern", p.Pattern),
			log.Bool("isRegExp", p.IsRegExp),
			log.Bool("isStructuralPat", p.IsStructuralPat),
			log.Strings("languages", p.Languages),
			log.Bool("isWordMatch", p.IsWordMatch),
			log.Bool("isCaseSensitive", p.IsCaseSensitive),
			log.Bool("patternMatchesContent", p.PatternMatchesContent),
			log.Bool("patternMatchesPath", p.PatternMatchesPath),
			log.Int("matches", sender.SentCount()),
			log.String("code", code),
			log.Duration("duration", time.Since(start)),
			log.Error(err))
	}(time.Now())

	if p.IsStructuralPat && p.Indexed {
		// Execute the new structural search path that directly calls Zoekt.
		// TODO use limit in indexed structural search
		return structuralSearchWithZoekt(ctx, s.Log, s.Indexed, p, sender)
	}

	// Compile pattern before fetching from store incase it is bad.
	var rg *readerGrep
	if !p.IsStructuralPat {
		rg, err = compile(&p.PatternInfo)
		if err != nil {
			return badRequestError{err.Error()}
		}
	}

	if p.FetchTimeout == time.Duration(0) {
		p.FetchTimeout = 500 * time.Millisecond
	}
	prepareCtx, cancel := context.WithTimeout(ctx, p.FetchTimeout)
	defer cancel()

	getZf := func() (string, *zipFile, error) {
		path, err := s.Store.PrepareZip(prepareCtx, p.Repo, p.Commit)
		if err != nil {
			return "", nil, err
		}
		zf, err := s.Store.zipCache.Get(path)
		return path, zf, err
	}

	// Hybrid search only works with our normal searcher code path, not
	// structural search.
	hybrid := !p.IsStructuralPat
	if hybrid {
		logger := logWithTrace(ctx, s.Log).Scoped("hybrid").With(
			log.String("repo", string(p.Repo)),
			log.String("commit", string(p.Commit)),
		)

		unsearched, ok, err := s.hybrid(ctx, logger, p, sender)
		if err != nil {
			// error logging is done inside of s.hybrid so we just return
			// error here.
			return errors.Wrap(err, "hybrid search failed")
		}
		if !ok {
			logger.Debug("hybrid search is falling back to normal unindexed search")
		} else {
			// now we only need to search unsearched
			if len(unsearched) == 0 {
				// indexed search did it all
				return nil
			}

			getZf = func() (string, *zipFile, error) {
				path, err := s.Store.PrepareZipPaths(prepareCtx, p.Repo, p.Commit, unsearched)
				if err != nil {
					return "", nil, err
				}
				zf, err := s.Store.zipCache.Get(path)
				return path, zf, err
			}
		}
	}

	zipPath, zf, err := getZipFileWithRetry(getZf)
	if err != nil {
		return errors.Wrap(err, "failed to get archive")
	}
	defer zf.Close()

	nFiles := uint64(len(zf.Files))
	bytes := int64(len(zf.Data))
	tr.AddEvent("archive",
		attribute.Int64("archive.files", int64(nFiles)),
		attribute.Int64("archive.size", bytes))
	metricArchiveFiles.Observe(float64(nFiles))
	metricArchiveSize.Observe(float64(bytes))

	if p.IsStructuralPat {
		return filteredStructuralSearch(ctx, s.Log, zipPath, zf, &p.PatternInfo, p.Repo, sender)
	} else {
		return regexSearch(ctx, rg, zf, p.PatternMatchesContent, p.PatternMatchesPath, p.IsNegated, sender)
	}
}

func validateParams(p *protocol.Request) error {
	if p.Repo == "" {
		return errors.New("Repo must be non-empty")
	}
	// Surprisingly this is the same sanity check used in the git source.
	if len(p.Commit) != 40 {
		return errors.Errorf("Commit must be resolved (Commit=%q)", p.Commit)
	}
	if p.Pattern == "" && p.ExcludePattern == "" && len(p.IncludePatterns) == 0 {
		return errors.New("At least one of pattern and include/exclude pattners must be non-empty")
	}
	if p.IsNegated && p.IsStructuralPat {
		return errors.New("Negated patterns are not supported for structural searches")
	}
	return nil
}

const megabyte = float64(1000 * 1000)

var (
	metricRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "searcher_service_running",
		Help: "Number of running search requests.",
	})
	metricArchiveSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "searcher_service_archive_size_bytes",
		Help:    "Observes the size when an archive is searched.",
		Buckets: []float64{1 * megabyte, 10 * megabyte, 100 * megabyte, 500 * megabyte, 1000 * megabyte, 5000 * megabyte},
	})
	metricArchiveFiles = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "searcher_service_archive_files",
		Help:    "Observes the number of files when an archive is searched.",
		Buckets: []float64{100, 1000, 10000, 50000, 100000},
	})
	metricRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "searcher_service_request_total",
		Help: "Number of returned search requests.",
	}, []string{"code"})
)

type badRequestError struct{ msg string }

func (e badRequestError) Error() string    { return e.msg }
func (e badRequestError) BadRequest() bool { return true }
