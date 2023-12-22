package internal

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	execRunning = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_gitserver_exec_running",
		Help: "number of gitserver.GitCommand running concurrently.",
	}, []string{"cmd"})
	execDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_gitserver_exec_duration_seconds",
		Help:    "gitserver.GitCommand latencies in seconds.",
		Buckets: trace.UserLatencyBuckets,
	}, []string{"cmd", "status"})
	blockedCommandExecutedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_exec_blocked_command_received",
		Help: "Incremented each time a command not in the allowlist for gitserver is executed",
	})
)

type NotFoundError struct {
	Payload *protocol.NotFoundPayload
}

func (e *NotFoundError) Error() string { return "not found" }

type execStatus struct {
	ExitStatus int
	Stderr     string
	Err        error
}

var ErrInvalidCommand = errors.New("invalid command")

// TODO: eseliger
// exec runs a git command. After the first write to w, it must not return an error.
// TODO(@camdencheek): once gRPC is the only consumer of this, do everything with errors
// because gRPC can handle trailing errors on a stream.
func (s *Server) exec(ctx context.Context, logger log.Logger, req *protocol.ExecRequest, userAgent string, w io.Writer) (execStatus, error) {
	// 🚨 SECURITY: Ensure that only commands in the allowed list are executed.
	// See https://github.com/sourcegraph/security-issues/issues/213.

	repoPath := string(protocol.NormalizeRepo(req.Repo))
	repoDir := filepath.Join(s.ReposDir, filepath.FromSlash(repoPath))

	if !gitdomain.IsAllowedGitCmd(logger, req.Args, repoDir) {
		blockedCommandExecutedCounter.Inc()
		return execStatus{}, ErrInvalidCommand
	}

	if !req.NoTimeout {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, executil.ShortGitCommandTimeout(req.Args))
		defer cancel()
	}

	start := time.Now()
	var cmdStart time.Time // set once we have ensured commit
	exitStatus := executil.UnsetExitStatus
	var stdoutN, stderrN int64
	var status string
	var execErr error
	ensureRevisionStatus := "noop"

	req.Repo = protocol.NormalizeRepo(req.Repo)
	repoName := req.Repo

	// Instrumentation
	{
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		args := strings.Join(req.Args, " ")

		var tr trace.Trace
		tr, ctx = trace.New(ctx, "exec."+cmd, repoName.Attr())
		tr.SetAttributes(
			attribute.String("args", args),
			attribute.String("ensure_revision", req.EnsureRevision),
		)
		logger = logger.WithTrace(trace.Context(ctx))

		execRunning.WithLabelValues(cmd).Inc()
		defer func() {
			tr.AddEvent(
				"done",
				attribute.String("status", status),
				attribute.Int64("stdout", stdoutN),
				attribute.Int64("stderr", stderrN),
				attribute.String("ensure_revision_status", ensureRevisionStatus),
			)
			tr.SetError(execErr)
			tr.End()

			duration := time.Since(start)
			execRunning.WithLabelValues(cmd).Dec()
			execDuration.WithLabelValues(cmd, status).Observe(duration.Seconds())

			var cmdDuration time.Duration
			var fetchDuration time.Duration
			if !cmdStart.IsZero() {
				cmdDuration = time.Since(cmdStart)
				fetchDuration = cmdStart.Sub(start)
			}

			isSlow := cmdDuration > shortGitCommandSlow(req.Args)
			isSlowFetch := fetchDuration > 10*time.Second
			if honey.Enabled() || traceLogs || isSlow || isSlowFetch {
				act := actor.FromContext(ctx)
				ev := honey.NewEvent("gitserver-exec")
				ev.SetSampleRate(honeySampleRate(cmd, act))
				ev.AddField("repo", repoName)
				ev.AddField("cmd", cmd)
				ev.AddField("args", args)
				ev.AddField("actor", act.UIDString())
				ev.AddField("ensure_revision", req.EnsureRevision)
				ev.AddField("ensure_revision_status", ensureRevisionStatus)
				ev.AddField("client", userAgent)
				ev.AddField("duration_ms", duration.Milliseconds())
				ev.AddField("stdin_size", len(req.Stdin))
				ev.AddField("stdout_size", stdoutN)
				ev.AddField("stderr_size", stderrN)
				ev.AddField("exit_status", exitStatus)
				ev.AddField("status", status)
				if execErr != nil {
					ev.AddField("error", execErr.Error())
				}
				if !cmdStart.IsZero() {
					ev.AddField("cmd_duration_ms", cmdDuration.Milliseconds())
					ev.AddField("fetch_duration_ms", fetchDuration.Milliseconds())
				}

				if traceID := trace.ID(ctx); traceID != "" {
					ev.AddField("traceID", traceID)
					ev.AddField("trace", trace.URL(traceID, conf.DefaultClient()))
				}

				if honey.Enabled() {
					_ = ev.Send()
				}

				if traceLogs {
					logger.Debug("TRACE gitserver exec", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
				}
				if isSlow {
					logger.Warn("Long exec request", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
				}
				if isSlowFetch {
					logger.Warn("Slow fetch/clone for exec request", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
				}
			}
		}()
	}

	if notFoundPayload, cloned := s.maybeStartClone(ctx, logger, repoName); !cloned {
		if notFoundPayload.CloneInProgress {
			status = "clone-in-progress"
		} else {
			status = "repo-not-found"
		}

		return execStatus{}, &NotFoundError{notFoundPayload}
	}

	dir := gitserverfs.RepoDirFromName(s.ReposDir, repoName)
	if s.ensureRevision(ctx, repoName, req.EnsureRevision, dir) {
		ensureRevisionStatus = "fetched"
	}

	// Special-case `git rev-parse HEAD` requests. These are invoked by search queries for every repo in scope.
	// For searches over large repo sets (> 1k), this leads to too many child process execs, which can lead
	// to a persistent failure mode where every exec takes > 10s, which is disastrous for gitserver performance.
	if len(req.Args) == 2 && req.Args[0] == "rev-parse" && req.Args[1] == "HEAD" {
		if resolved, err := git.QuickRevParseHead(dir); err == nil && gitdomain.IsAbsoluteRevision(resolved) {
			_, _ = w.Write([]byte(resolved))
			return execStatus{}, nil
		}
	}

	// Special-case `git symbolic-ref HEAD` requests. These are invoked by resolvers determining the default branch of a repo.
	// For searches over large repo sets (> 1k), this leads to too many child process execs, which can lead
	// to a persistent failure mode where every exec takes > 10s, which is disastrous for gitserver performance.
	if len(req.Args) == 2 && req.Args[0] == "symbolic-ref" && req.Args[1] == "HEAD" {
		if resolved, err := git.QuickSymbolicRefHead(dir); err == nil {
			_, _ = w.Write([]byte(resolved))
			return execStatus{}, nil
		}
	}

	var stderrBuf bytes.Buffer
	stdoutW := &writeCounter{w: w}
	stderrW := &writeCounter{w: &limitWriter{W: &stderrBuf, N: 1024}}

	cmdStart = time.Now()
	cmd := s.RecordingCommandFactory.Command(ctx, s.Logger, string(repoName), "git", req.Args...)
	dir.Set(cmd.Unwrap())
	cmd.Unwrap().Stdout = stdoutW
	cmd.Unwrap().Stderr = stderrW
	cmd.Unwrap().Stdin = bytes.NewReader(req.Stdin)

	exitStatus, execErr = executil.RunCommand(ctx, cmd)

	status = strconv.Itoa(exitStatus)
	stdoutN = stdoutW.n
	stderrN = stderrW.n

	stderr := stderrBuf.String()
	s.logIfCorrupt(ctx, repoName, dir, stderr)

	return execStatus{
		Err:        execErr,
		Stderr:     stderr,
		ExitStatus: exitStatus,
	}, nil
}
