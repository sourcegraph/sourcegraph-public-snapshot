package server

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/mxk/go-flowrate/flowrate"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

var uploadPackArgs = []string{
	// Partial clones/fetches
	"-c", "uploadpack.allowFilter=true",

	// Can fetch any object. Used in case of race between a resolve ref and a
	// fetch of a commit. Safe to do, since this is only used internally.
	"-c", "uploadpack.allowAnySHA1InWant=true",

	"upload-pack",

	"--stateless-rpc", "--strict",
}

// gitServiceHandler is a smart Git HTTP transfer protocol as documented at
// https://www.git-scm.com/docs/http-protocol.
//
// This allows users to clone any git repo. We only support the smart
// protocol. We aim to support modern git features such as protocol v2 to
// minimize traffic.
type gitServiceHandler struct {
	// Dir is a funcion which takes a repository name and returns an absolute
	// path to the GIT_DIR for it.
	Dir func(string) string

	// CommandHook if non-nil will run with the git upload command before we
	// start the command.
	//
	// This allows the command to be modified before running. In practice
	// sourcegraph.com will add a flowrated writer for Stdout to treat our
	// internal networks more kindly.
	CommandHook func(*exec.Cmd)

	// Trace if non-nil is called at the start of serving a request. It will
	// call the returned function when done executing. If the executation
	// failed, it will pass in a non-nil error.
	Trace func(svc, repo, protocol string) func(error)
}

func (s *gitServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only support clones and fetches (git upload-pack). /info/refs sets the
	// service field.
	if svcQ := r.URL.Query().Get("service"); svcQ != "" && svcQ != "git-upload-pack" {
		http.Error(w, "only support service git-upload-pack", http.StatusBadRequest)
		return
	}

	var repo, svc string
	for _, suffix := range []string{"/info/refs", "/git-upload-pack"} {
		if strings.HasSuffix(r.URL.Path, suffix) {
			svc = suffix
			repo = strings.TrimSuffix(r.URL.Path, suffix)
			repo = strings.TrimPrefix(repo, "/")
			break
		}
	}

	dir := s.Dir(repo)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		http.Error(w, "repository not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "failed to stat repo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	body := r.Body
	defer body.Close()

	if r.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(body)
		if err != nil {
			http.Error(w, "malformed payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer gzipReader.Close()

		body = gzipReader
	}

	// err is set if we fail to run command or have an unexpected svc. It is
	// captured for tracing.
	var err error
	if s.Trace != nil {
		done := s.Trace(svc, repo, r.Header.Get("Git-Protocol"))
		defer func() {
			done(err)
		}()
	}

	args := append([]string{}, uploadPackArgs...)
	switch svc {
	case "/info/refs":
		w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
		_, _ = w.Write(packetWrite("# service=git-upload-pack\n"))
		_, _ = w.Write([]byte("0000"))
		args = append(args, "--advertise-refs")
	case "/git-upload-pack":
		w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	default:
		err = fmt.Errorf("unexpected subpath (want /info/refs or /git-upload-pack): %q", svc)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	args = append(args, dir)

	env := os.Environ()
	if protocol := r.Header.Get("Git-Protocol"); protocol != "" {
		env = append(env, "GIT_PROTOCOL="+protocol)
	}

	cmd := exec.CommandContext(r.Context(), "git", args...)
	cmd.Env = env
	cmd.Stdout = w
	cmd.Stdin = body

	if s.CommandHook != nil {
		s.CommandHook(cmd)
	}

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("error running git service command args=%q: %w", args, err)
		_, _ = w.Write([]byte("\n" + err.Error() + "\n"))
	}
}

func packetWrite(str string) []byte {
	s := strconv.FormatInt(int64(len(str)+4), 16)
	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}
	return []byte(s + str)
}

// flowrateWriter limits the write rate of w to 1 Gbps.
//
// We are cloning repositories from within the same network from another
// Sourcegraph service (zoekt-indexserver). This can end up being so fast that
// we harm our own network connectivity. In the case of zoekt-indexserver and
// gitserver running on the same host machine, we can even reach up to ~100
// Gbps and effectively DoS the Docker network, temporarily disrupting other
// containers running on the host.
//
// Google Compute Engine has a network bandwidth of about 1.64 Gbps
// between nodes, and AWS varies widely depending on instance type.
// We play it safe and default to 1 Gbps here (~119 MiB/s), which
// means we can fetch a 1 GiB archive in ~8.5 seconds.
func flowrateWriter(w io.Writer) io.Writer {
	const megabit = int64(1000 * 1000)
	const limit = 1000 * megabit // 1 Gbps
	return flowrate.NewWriter(w, limit)
}

func (s *Server) gitServiceHandler() *gitServiceHandler {
	return &gitServiceHandler{
		Dir: func(d string) string {
			return string(s.dir(api.RepoName(d)))
		},

		// Limit rate of stdout from git.
		CommandHook: func(cmd *exec.Cmd) {
			cmd.Stdout = flowrateWriter(cmd.Stdout)
		},

		Trace: func(svc, repo, protocol string) func(error) {
			start := time.Now()
			metricServiceRunning.WithLabelValues(svc).Inc()
			return func(err error) {
				metricServiceRunning.WithLabelValues(svc).Dec()
				metricServiceDuration.WithLabelValues(svc).Observe(time.Since(start).Seconds())

				if err != nil {
					log15.Error("gitservice.ServeHTTP", "svc", svc, "repo", repo, "protocol", protocol, "duration", time.Since(start), "error", err.Error())
				} else if traceLogs {
					log15.Debug("TRACE gitserver git service", "svc", svc, "repo", repo, "protocol", protocol, "duration", time.Since(start))
				}
			}
		},
	}
}

var (
	metricServiceDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_gitserver_gitservice_duration_seconds",
		Help:    "A histogram of latencies for the git service (upload-pack for internal clones) endpoint.",
		Buckets: prometheus.ExponentialBuckets(.1, 5, 5), // 100ms -> 62s
	}, []string{"type"})

	metricServiceRunning = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_gitserver_gitservice_running",
		Help: "A histogram of latencies for the git service (upload-pack for internal clones) endpoint.",
	}, []string{"type"})
)
