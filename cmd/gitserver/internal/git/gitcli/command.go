package gitcli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/sourcegraph/sourcegraph/internal/bytesize"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/memcmd"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
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
		Buckets: prometheus.ExponentialBucketsRange(0.01, 60.0, 12),
	}, []string{"cmd"})
	highMemoryCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_gitserver_exec_high_memory_usage_count",
		Help: "gitcli.GitCommand high memory usage by subcommand",
	}, []string{"cmd"})

	memoryObservationEnabled = env.MustGetBool("GITSERVER_MEMORY_OBSERVATION_ENABLED", false, "enable memory observation for gitserver commands")
)

type commandOpts struct {
	arguments []string

	stdin io.Reader
}

func optsFromFuncs(optFns ...CommandOptionFunc) commandOpts {
	var opts commandOpts
	for _, optFn := range optFns {
		optFn(&opts)
	}
	return opts
}

type CommandOptionFunc func(*commandOpts)

// WithArguments sets the given arguments to the command arguments.
func WithArguments(args ...string) CommandOptionFunc {
	return func(o *commandOpts) {
		o.arguments = args
	}
}

// WithStdin specifies the reader to use for the command's stdin input.
func WithStdin(stdin io.Reader) CommandOptionFunc {
	return func(o *commandOpts) {
		o.stdin = stdin
	}
}

const gitCommandDefaultTimeout = time.Minute

func (g *gitCLIBackend) NewCommand(ctx context.Context, optFns ...CommandOptionFunc) (_ io.ReadCloser, err error) {
	opts := optsFromFuncs(optFns...)

	tr, ctx := trace.New(ctx, "gitcli.NewCommand",
		attribute.StringSlice("args", opts.arguments),
		attribute.String("dir", g.dir.Path()),
	)
	defer func() {
		if err != nil {
			tr.EndWithErr(&err)
		}
	}()

	logger := g.logger.WithTrace(trace.Context(ctx))

	if !IsAllowedGitCmd(logger, opts.arguments) {
		blockedCommandExecutedCounter.Inc()
		return nil, errBadGitCommand
	}

	if len(opts.arguments) == 0 {
		// Technically can't happen because IsAllowedGitCmd catches this, but
		// if someone ever touches that logic, let's be safe before we access
		// args[0].
		return nil, errors.New("provide arguments")
	}

	subCmd := opts.arguments[0]

	// If no deadline is set, use the default git command timeout.
	cancel := func() {}
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, gitCommandDefaultTimeout)
	}

	cmd := exec.CommandContext(ctx, "git", opts.arguments...)
	cmd.Cancel = func() error {
		// Send SIGKILL to the process group instead of just the process
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	g.dir.Set(cmd)

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// We use setpgid here so that child processes live in their own process groups.
	// This is helpful for two things:
	// - We can kill a process group to make sure that any subprocesses git might spawn
	//   will also receive the termination signal. The standard go implementation sends
	//   a SIGKILL only to the process itself. By using process groups, we can tell all
	//   children to shut down as well.
	// - We want to track maxRSS for tracking purposes to identify memory usage by command
	//   and linux tracks the maxRSS as "the maximum resident set size used (in kilobytes)"
	//   of the process in the process group that had the highest maximum resident set size.
	//   Read: If we don't use a separate process group here, we usually get the maxRSS from
	//   the process with the biggest memory usage in the process group, which is gitserver.
	//   So we cannot track the memory well. This is leaky, as it only tracks the largest sub-
	//   process, but it gives us a good indication of the general resource consumption.
	cmd.SysProcAttr.Setpgid = true

	stderr, stderrBuf := stderrBuffer()
	cmd.Stderr = stderr

	wrappedCmd := g.rcf.WrapWithRepoName(ctx, logger, g.repoName, cmd)

	stdout, err := wrappedCmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "failed to create stdout pipe")
	}

	if opts.stdin != nil {
		cmd.Stdin = opts.stdin
	}

	cmdStart := time.Now()

	if err := wrappedCmd.Start(); err != nil {
		cancel()
		return nil, errors.Wrap(err, "failed to start git process")
	}

	observer := memcmd.NewNoOpObserver()

	if memoryObservationEnabled {
		maybeObs, err := memcmd.NewDefaultObserver(ctx, cmd)
		if err != nil {
			logger.Warn("failed to create memory observer, defaulting to no-op", log.Error(err))
		}

		observer = maybeObs
	}
	observer.Start()
	defer func() {
		if err != nil {
			observer.Stop() // stop the observer if we return an error
		}
	}()

	execRunning.WithLabelValues(subCmd).Inc()

	cr := &cmdReader{
		ctx:            ctx,
		subCmd:         subCmd,
		ctxCancel:      cancel,
		cmdStart:       cmdStart,
		stdout:         stdout,
		cmd:            wrappedCmd,
		stderr:         stderrBuf,
		repoName:       g.repoName,
		logger:         logger,
		gitDir:         g.dir,
		tr:             tr,
		memoryObserver: observer,
	}

	return cr, nil
}

// errBadGitCommand is returned from the git CLI backend if the arguments provided
// are not allowed.
var errBadGitCommand = errors.New("bad git command, not allowed")

func newCommandFailedError(ctx context.Context, err error, cmd wrexec.Cmder, stderr []byte) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return &commandFailedError{
		Inner:      err,
		args:       cmd.Unwrap().Args,
		Stderr:     stderr,
		ExitStatus: cmd.Unwrap().ProcessState.ExitCode(),
	}
}

type commandFailedError struct {
	Stderr     []byte
	ExitStatus int
	Inner      error
	args       []string
}

func (e *commandFailedError) Unwrap() error {
	return e.Inner
}

func (e *commandFailedError) Error() string {
	return fmt.Sprintf("git command %v failed with status code %d (output: %q)", e.args, e.ExitStatus, e.Stderr)
}

type cmdReader struct {
	stdout         io.Reader
	ctx            context.Context
	ctxCancel      context.CancelFunc
	subCmd         string
	cmdStart       time.Time
	cmd            wrexec.Cmder
	stderr         *bytes.Buffer
	logger         log.Logger
	gitDir         common.GitDir
	repoName       api.RepoName
	mu             sync.Mutex
	tr             trace.Trace
	err            error
	waitOnce       sync.Once
	memoryObserver memcmd.Observer
}

func (rc *cmdReader) Read(p []byte) (n int, err error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	n, err = rc.stdout.Read(p)
	// If the command has finished, we close the stdout pipe and wait on the command
	// to free any leftover resources. If it errored, this will return the command
	// error from Read.
	if err == io.EOF {
		if err := rc.waitCmd(); err != nil {
			return n, err
		}
	}
	return n, err
}

func (rc *cmdReader) Close() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	return rc.waitCmd()
}

func (rc *cmdReader) waitCmd() error {
	// Waiting on a command should only happen once, so
	// we synchronize all potential calls to Read and Close
	// here, and memoize the error.
	rc.waitOnce.Do(func() {
		rc.err = rc.cmd.Wait()
		rc.memoryObserver.Stop()

		if rc.err != nil {
			if checkMaybeCorruptRepo(rc.logger, rc.gitDir, rc.repoName, rc.stderr.String()) {
				rc.err = common.ErrRepoCorrupted{Reason: rc.stderr.String()}
			} else {
				rc.err = newCommandFailedError(rc.ctx, rc.err, rc.cmd, rc.stderr.Bytes())
			}
		}

		rc.trace()
		rc.tr.EndWithErr(&rc.err)
		rc.ctxCancel()
	})

	return rc.err
}

const highMemoryUsageThreshold = 500 * bytesize.MiB

func (rc *cmdReader) trace() {
	duration := time.Since(rc.cmdStart)

	execRunning.WithLabelValues(rc.subCmd).Dec()
	execDuration.WithLabelValues(rc.subCmd).Observe(duration.Seconds())

	processState := rc.cmd.Unwrap().ProcessState
	var sysUsage syscall.Rusage
	s, ok := processState.SysUsage().(*syscall.Rusage)
	if ok {
		sysUsage = *s
	}

	memUsage, memoryError := rc.memoryObserver.MaxMemoryUsage()
	if memoryError != nil {
		if !(isContextErr(memoryError) && isContextErr(rc.ctx.Err())) {
			// If the context was canceled, we don't log the error as it's expected.
			rc.logger.Warn("failed to get max memory usage", log.Error(memoryError))
		}
	}

	isSlow := duration > shortGitCommandSlow(rc.cmd.Unwrap().Args)

	isHighMem := memUsage > highMemoryUsageThreshold

	if isHighMem {
		highMemoryCounter.WithLabelValues(rc.subCmd).Inc()
	}

	if honey.Enabled() || isSlow || isHighMem {
		act := actor.FromContext(rc.ctx)
		ev := honey.NewEvent("gitserver-exec")
		ev.SetSampleRate(HoneySampleRate(rc.subCmd, act))
		ev.AddField("repo", rc.repoName)
		ev.AddField("cmd", rc.subCmd)
		ev.AddField("args", rc.cmd.Unwrap().Args)
		ev.AddField("actor", act.UIDString())
		ev.AddField("exit_status", processState.ExitCode())
		if rc.err != nil {
			ev.AddField("error", rc.err.Error())
		}
		ev.AddField("cmd_duration_ms", duration.Milliseconds())
		ev.AddField("user_time_ms", processState.UserTime().Milliseconds())
		ev.AddField("system_time_ms", processState.SystemTime().Milliseconds())
		ev.AddField("cmd_ru_maxrss_kib", memUsage>>10)
		ev.AddField("cmd_ru_maxrss_human_readable", humanize.Bytes(uint64(memUsage)))
		ev.AddField("cmd_ru_minflt", sysUsage.Minflt)
		ev.AddField("cmd_ru_majflt", sysUsage.Majflt)
		ev.AddField("cmd_ru_inblock", sysUsage.Inblock)
		ev.AddField("cmd_ru_oublock", sysUsage.Oublock)

		if traceID := trace.ID(rc.ctx); traceID != "" {
			ev.AddField("traceID", traceID)
			ev.AddField("trace", trace.URL(traceID))
		}

		if honey.Enabled() {
			_ = ev.Send()
		}

		if isSlow {
			rc.logger.Warn("Long exec request", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
		}
		if isHighMem {
			rc.logger.Warn("High memory usage exec request", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
		}
	}

	rc.tr.SetAttributes(attribute.Int("exit_code", processState.ExitCode()))
	rc.tr.SetAttributes(attribute.Int64("cmd_duration_ms", duration.Milliseconds()))
	rc.tr.SetAttributes(attribute.Int64("user_time_ms", processState.UserTime().Milliseconds()))
	rc.tr.SetAttributes(attribute.Int64("system_time_ms", processState.SystemTime().Milliseconds()))
	rc.tr.SetAttributes(attribute.Int64("cmd_ru_maxrss_kib", int64(memUsage>>10)))
	rc.tr.SetAttributes(attribute.String("cmd_ru_maxrss_human_readable", humanize.Bytes(uint64(memUsage))))
	rc.tr.SetAttributes(attribute.Int64("cmd_ru_minflt", sysUsage.Minflt))
	rc.tr.SetAttributes(attribute.Int64("cmd_ru_majflt", sysUsage.Majflt))
	rc.tr.SetAttributes(attribute.Int64("cmd_ru_inblock", sysUsage.Inblock))
	rc.tr.SetAttributes(attribute.Int64("cmd_ru_oublock", sysUsage.Oublock))
}

const maxStderrCapture = 1024

// stderrBuffer sets up a limited buffer to capture stderr for error reporting.
func stderrBuffer() (io.Writer, *bytes.Buffer) {
	stderrBuf := bytes.NewBuffer(make([]byte, 0, maxStderrCapture))
	stderr := &limitWriter{W: stderrBuf, N: maxStderrCapture}
	return stderr, stderrBuf
}

// limitWriter is a io.Writer that writes to an W but discards after N bytes.
type limitWriter struct {
	W io.Writer // underling writer
	N int       // max bytes remaining
}

func (l *limitWriter) Write(p []byte) (int, error) {
	if l.N <= 0 {
		return len(p), nil
	}
	origLen := len(p)
	if len(p) > l.N {
		p = p[:l.N]
	}
	n, err := l.W.Write(p)
	l.N -= n
	if l.N <= 0 {
		// If we have written limit bytes, then we can include the discarded
		// part of p in the count.
		n = origLen
	}
	return n, err
}

func checkMaybeCorruptRepo(logger log.Logger, girDir common.GitDir, repo api.RepoName, stderr string) bool {
	if !stdErrIndicatesCorruption(stderr) {
		return false
	}

	logger = logger.With(log.String("repo", string(repo)), log.String("repo", string(repo)))
	logger.Warn("marking repo for re-cloning due to stderr output indicating repo corruption", log.String("stderr", stderr))

	// We set a flag in the config for the cleanup janitor job to fix. The janitor
	// runs every minute.
	// We use a background context here to record corruption events even when the
	// context has since been cancelled.
	err := markRepoMaybeCorrupt(girDir)
	if err != nil {
		logger.Error("failed to set maybeCorruptRepo flag", log.Error(err))
	}

	return true
}

// markRepoMaybeCorrupt ensures a file called sourcegraph.maybeCorruptRepo exists
// in the repo root, and makes sure the files mtime set to the current time.
func markRepoMaybeCorrupt(gitDir common.GitDir) error {
	p := gitDir.Path(RepoMaybeCorruptFlagFilepath)

	f, err := os.Create(p)
	if err != nil && !os.IsExist(err) {
		return err
	}
	_ = f.Close()
	return os.Chtimes(p, time.Time{}, time.Now())
}

// RepoMaybeCorruptFlagFilepath is a magic file we add to the root of a git repo
// to signal that a repo may be corrupt on disk, when the git stderr output indicates
// potential corruption.
const RepoMaybeCorruptFlagFilepath = ".sourcegraph-maybe-corrupt-repo"

var (
	// objectOrPackFileCorruptionRegex matches stderr lines from git which indicate
	// that a repository's packfiles or commit objects might be corrupted.
	//
	// See https://github.com/sourcegraph/sourcegraph/issues/6676 for more
	// context.
	objectOrPackFileCorruptionRegex = lazyregexp.NewPOSIX(`^error: (Could not read|packfile) `)

	// objectOrPackFileCorruptionRegex matches stderr lines from git which indicate that
	// git's supplemental commit-graph might be corrupted.
	//
	// See https://github.com/sourcegraph/sourcegraph/issues/37872 for more
	// context.
	commitGraphCorruptionRegex = lazyregexp.NewPOSIX(`^fatal: commit-graph requires overflow generation data but has none`)
)

// stdErrIndicatesCorruption returns true if the provided stderr output from a git command indicates
// that there might be repository corruption.
func stdErrIndicatesCorruption(stderr string) bool {
	return objectOrPackFileCorruptionRegex.MatchString(stderr) || commitGraphCorruptionRegex.MatchString(stderr)
}

// shortGitCommandSlow returns the threshold for regarding an git command as
// slow. Some commands such as "git archive" are inherently slower than "git
// rev-parse", so this will return an appropriate threshold given the command.
func shortGitCommandSlow(args []string) time.Duration {
	if len(args) < 1 {
		return time.Second
	}
	switch args[0] {
	case "archive":
		return 1 * time.Minute

	case "blame", "ls-tree", "log", "show":
		return 5 * time.Second

	case "fetch":
		return 10 * time.Second

	default:
		return 2500 * time.Millisecond
	}
}

// mapToLoggerField translates a map to log context fields.
func mapToLoggerField(m map[string]any) []log.Field {
	LogFields := []log.Field{}

	for i, v := range m {

		LogFields = append(LogFields, log.String(i, fmt.Sprint(v)))
	}

	return LogFields
}

// Send 1 in 16 events to honeycomb. This is hardcoded since we only use this
// for Sourcegraph.com.
//
// 2020-05-29 1 in 4. We are currently at the top tier for honeycomb (before
// enterprise) and using double our quota. This gives us room to grow. If you
// find we keep bumping this / missing data we care about we can look into
// more dynamic ways to sample in our application code.
//
// 2020-07-20 1 in 16. Again hitting very high usage. Likely due to recent
// scaling up of the indexed search cluster. Will require more investigation,
// but we should probably segment user request path traffic vs internal batch
// traffic.
//
// 2020-11-02 Dynamically sample. Again hitting very high usage. Same root
// cause as before, scaling out indexed search cluster. We update our sampling
// to instead be dynamic, since "rev-parse" is 12 times more likely than the
// next most common command.
//
// 2021-08-20 over two hours we did 128 * 128 * 1e6 rev-parse requests
// internally. So we update our sampling to heavily downsample internal
// rev-parse, while upping our sampling for non-internal.
// https://ui.honeycomb.io/sourcegraph/datasets/gitserver-exec/result/67e4bLvUddg
//
// 2024-02-23 we are now capturing all execs done in honeycomb, including
// internal stuff like config and janitor jobs. In particular "config" is now
// running as often as rev-parse. rev-list is also higher than most so we
// include it in the big sample rate.
func HoneySampleRate(cmd string, actor *actor.Actor) uint {
	// HACK(keegan) 2022-11-02 IsInternal on sourcegraph.com is always
	// returning false. For now I am also marking it internal if UID is not
	// set to work around us hammering honeycomb.
	internal := actor.IsInternal() || actor.UID == 0
	switch {
	case (cmd == "rev-parse" || cmd == "rev-list" || cmd == "config") && internal:
		return 1 << 14 // 16384

	case internal:
		// we care more about user requests, so downsample internal more.
		return 16

	default:
		return 8
	}
}

func isContextErr(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
