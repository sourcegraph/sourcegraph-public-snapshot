// Package replace is a service exposing an API to replace file contents in a repo.
// It streams back results with JSON lines.
//
// Architecture Notes:
// - The following are the same as cmd/searcher/search.go:
// * Archive is fetched from gitserver
// * Simple HTTP API exposed
// * Currently no concept of authorization
// * On disk cache of fetched archives to reduce load on gitserver
//
// - Here is where replacer.go differs
// * Pass the zip file path to external replacer tool(s) after validating
// * Read tool stdout and write it out on the HTTP connection
// * Input from stdout is expected to use JSON lines format, but the format isn't checked here: line-buffering is done on the frontend

package replace

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/net/trace"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/replacer/protocol"
	"github.com/sourcegraph/sourcegraph/internal/store"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/schema"
)

type Service struct {
	Store *store.Store
	Log   log15.Logger
}

type ExternalTool struct {
	Name       string
	BinaryPath string
}

// Configure the command line options and return the command to execute using an external tool
func (t *ExternalTool) command(ctx context.Context, spec *protocol.RewriteSpecification, zipPath string) (cmd *exec.Cmd, err error) {
	switch t.Name {
	case "comby":
		_, err = exec.LookPath("comby")
		if err != nil {
			return nil, errors.New("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
		}

		var args []string
		args = append(args, spec.MatchTemplate, spec.RewriteTemplate)

		if spec.FileExtension != "" {
			args = append(args, spec.FileExtension)
		}

		args = append(args, "-zip", zipPath, "-json-lines", "-json-only-diff")

		if spec.DirectoryExclude != "" {
			args = append(args, "-exclude-dir", spec.DirectoryExclude)
		}

		log15.Info(fmt.Sprintf("running command: comby %q", strings.Join(args[:], " ")))
		return exec.CommandContext(ctx, t.BinaryPath, args...), nil

	default:
		return nil, errors.Errorf("Unknown external replace tool %q.", t.Name)
	}
}

var decoder = schema.NewDecoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
}

// ServeHTTP handles HTTP based replace requests
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	running.Inc()
	defer running.Dec()

	err := r.ParseForm()
	if err != nil {
		log15.Info("Didn't parse" + err.Error())
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

	deadlineHit, err := s.replace(ctx, &p, w, r)
	if err != nil {
		code := http.StatusInternalServerError
		if isBadRequest(err) || ctx.Err() == context.Canceled {
			code = http.StatusBadRequest
		} else if isTemporary(err) {
			code = http.StatusServiceUnavailable
		} else {
			log.Printf("internal error serving %#+v: %s", p, err)
		}
		http.Error(w, err.Error(), code)
		return
	}

	if deadlineHit {
		log15.Info("Deadline hit")
		http.Error(w, "Deadline hit", http.StatusRequestTimeout)
		return
	}
}

func (s *Service) replace(ctx context.Context, p *protocol.Request, w http.ResponseWriter, r *http.Request) (deadlineHit bool, err error) {
	tr := trace.New("replace", fmt.Sprintf("%s@%s", p.Repo, p.Commit))
	tr.LazyPrintf("%s", p.RewriteSpecification)

	span, ctx := opentracing.StartSpanFromContext(ctx, "Replace")
	ext.Component.Set(span, "service")
	span.SetTag("repo", p.Repo)
	span.SetTag("url", p.URL)
	span.SetTag("commit", p.Commit)
	span.SetTag("rewriteSpecification", p.RewriteSpecification)
	defer func(start time.Time) {
		code := "200"
		// We often have canceled and timed out requests. We do not want to
		// record them as errors to avoid noise
		if ctx.Err() == context.Canceled {
			code = "canceled"
			span.SetTag("err", err)
		} else if ctx.Err() == context.DeadlineExceeded {
			code = "timedout"
			span.SetTag("err", err)
			deadlineHit = true
			err = nil // error is fully described by deadlineHit=true return value
		} else if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
			if isBadRequest(err) {
				code = "400"
			} else if isTemporary(err) {
				code = "503"
			} else {
				code = "500"
			}
		}
		tr.LazyPrintf("code=%s deadlineHit=%v", code, deadlineHit)
		tr.Finish()
		requestTotal.WithLabelValues(code).Inc()
		span.SetTag("deadlineHit", deadlineHit)
		span.Finish()

		if s.Log != nil {
			s.Log.Debug("replace request", "repo", p.Repo, "commit", p.Commit, "rewriteSpecification", p.RewriteSpecification, "code", code, "duration", time.Since(start), "err", err)
		}
	}(time.Now())

	if p.FetchTimeout == "" {
		p.FetchTimeout = "500ms"
	}
	fetchTimeout, err := time.ParseDuration(p.FetchTimeout)
	if err != nil {
		return false, err
	}
	prepareCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	getZf := func() (string, *store.ZipFile, error) {
		path, err := s.Store.PrepareZip(prepareCtx, p.GitserverRepo(), p.Commit)
		if err != nil {
			return "", nil, err
		}
		zf, err := s.Store.ZipCache.Get(path)
		return path, zf, err
	}

	zipPath, zf, err := store.GetZipFileWithRetry(getZf)
	if err != nil {
		return false, err
	}
	defer zf.Close()

	nFiles := uint64(len(zf.Files))
	bytes := int64(len(zf.Data))
	tr.LazyPrintf("files=%d bytes=%d", nFiles, bytes)
	span.LogFields(
		otlog.Uint64("archive.files", nFiles),
		otlog.Int64("archive.size", bytes))
	archiveFiles.Observe(float64(nFiles))
	archiveSize.Observe(float64(bytes))

	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	t := &ExternalTool{
		Name:       "comby",
		BinaryPath: "comby",
	}

	cmd, err := t.command(ctx, &p.RewriteSpecification, zipPath)
	if err != nil {
		log15.Info("Invalid command: " + err.Error())
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log15.Info("Could not connect to command stdout: " + err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		log15.Info("Error starting command: " + err.Error())
		return false, errors.New(err.Error())
	}

	_, err = io.Copy(w, stdout)
	if err != nil {
		log15.Info("Error copying external command output to HTTP writer: " + err.Error())
		return
	}

	if err := cmd.Wait(); err != nil {
		log15.Info("Error after executing command: " + string(err.(*exec.ExitError).Stderr))
	}

	return false, nil
}

func validateParams(p *protocol.Request) error {
	if p.Repo == "" {
		return errors.New("Repo must be non-empty")
	}
	// Surprisingly this is the same sanity check used in the git source.
	if len(p.Commit) != 40 {
		return errors.Errorf("Commit must be resolved (Commit=%q)", p.Commit)
	}

	if p.RewriteSpecification.MatchTemplate == "" {
		return errors.New("MatchTemplate must be non-empty")
	}
	return nil
}

const megabyte = float64(1000 * 1000)

var (
	running = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "replacer",
		Subsystem: "service",
		Name:      "running",
		Help:      "Number of running search requests.",
	})
	archiveSize = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "replacer",
		Subsystem: "service",
		Name:      "archive_size_bytes",
		Help:      "Observes the size when an archive is searched.",
		Buckets:   []float64{1 * megabyte, 10 * megabyte, 100 * megabyte, 500 * megabyte, 1000 * megabyte, 5000 * megabyte},
	})
	archiveFiles = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "replacer",
		Subsystem: "service",
		Name:      "archive_files",
		Help:      "Observes the number of files when an archive is searched.",
		Buckets:   []float64{100, 1000, 10000, 50000, 100000},
	})
	requestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "replacer",
		Subsystem: "service",
		Name:      "request_total",
		Help:      "Number of returned replace requests.",
	}, []string{"code"})
)

func init() {
	prometheus.MustRegister(running)
	prometheus.MustRegister(archiveSize)
	prometheus.MustRegister(archiveFiles)
	prometheus.MustRegister(requestTotal)
}

func isBadRequest(err error) bool {
	e, ok := errors.Cause(err).(interface {
		BadRequest() bool
	})
	return ok && e.BadRequest()
}

func isTemporary(err error) bool {
	e, ok := errors.Cause(err).(interface {
		Temporary() bool
	})
	return ok && e.Temporary()
}
