package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/accesslog"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/streamio"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func (gs *GRPCServer) P4Exec(req *proto.P4ExecRequest, ss proto.GitserverService_P4ExecServer) error {
	arguments := byteSlicesToStrings(req.GetArgs()) //nolint:staticcheck

	if len(arguments) < 1 {
		return status.Error(codes.InvalidArgument, "args must be greater than or equal to 1")
	}

	subCommand := arguments[0]

	// Make sure the subcommand is explicitly allowed
	allowlist := []string{"protects", "groups", "users", "group", "changes"}
	allowed := false
	for _, c := range allowlist {
		if subCommand == c {
			allowed = true
			break
		}
	}
	if !allowed {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("subcommand %q is not allowed", subCommand))
	}

	p4home, err := gitserverfs.MakeP4HomeDir(gs.Server.ReposDir)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// Log which actor is accessing p4-exec.
	//
	// p4-exec is currently only used for fetching user based permissions information
	// so, we don't have a repo name.
	accesslog.Record(ss.Context(), "<no-repo>",
		log.String("p4user", req.GetP4User()), //nolint:staticcheck
		log.String("p4port", req.GetP4Port()), //nolint:staticcheck
		log.Strings("args", arguments),
	)

	// Make sure credentials are valid before heavier operation
	err = perforce.P4TestWithTrust(ss.Context(), p4home, req.GetP4Port(), req.GetP4User(), req.GetP4Passwd()) //nolint:staticcheck
	if err != nil {
		if ctxErr := ss.Context().Err(); ctxErr != nil {
			return status.FromContextError(ctxErr).Err()
		}

		return status.Error(codes.InvalidArgument, err.Error())
	}

	w := streamio.NewWriter(func(p []byte) error {
		return ss.Send(&proto.P4ExecResponse{
			Data: p,
		})
	})

	var r p4ExecRequest
	r.FromProto(req)

	return gs.doP4Exec(ss.Context(), gs.Server.Logger, &r, "unknown-grpc-client", w)
}

func (gs *GRPCServer) doP4Exec(ctx context.Context, logger log.Logger, req *p4ExecRequest, userAgent string, w io.Writer) error {
	execStatus := gs.Server.p4Exec(ctx, logger, req, userAgent, w)

	if execStatus.ExitStatus != 0 || execStatus.Err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return status.FromContextError(ctxErr).Err()
		}

		gRPCStatus := codes.Unknown
		if strings.Contains(execStatus.Err.Error(), "signal: killed") {
			gRPCStatus = codes.Aborted
		}

		s, err := status.New(gRPCStatus, execStatus.Err.Error()).WithDetails(&proto.ExecStatusPayload{
			StatusCode: int32(execStatus.ExitStatus),
			Stderr:     execStatus.Stderr,
		})
		if err != nil {
			gs.Server.Logger.Error("failed to marshal status", log.Error(err))
			return err
		}
		return s.Err()
	}

	return nil
}

func (s *Server) handleP4Exec(w http.ResponseWriter, r *http.Request) {
	var req p4ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Args) < 1 {
		http.Error(w, "args must be greater than or equal to 1", http.StatusBadRequest)
		return
	}

	// Make sure the subcommand is explicitly allowed
	allowlist := []string{"protects", "groups", "users", "group", "changes"}
	allowed := false
	for _, arg := range allowlist {
		if req.Args[0] == arg {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, fmt.Sprintf("subcommand %q is not allowed", req.Args[0]), http.StatusBadRequest)
		return
	}

	p4home, err := gitserverfs.MakeP4HomeDir(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log which actor is accessing p4-exec.
	//
	// p4-exec is currently only used for fetching user based permissions information
	// so, we don't have a repo name.
	accesslog.Record(r.Context(), "<no-repo>",
		log.String("p4user", req.P4User),
		log.String("p4port", req.P4Port),
		log.Strings("args", req.Args),
	)

	// Make sure credentials are valid before heavier operation
	err = perforce.P4TestWithTrust(r.Context(), p4home, req.P4Port, req.P4User, req.P4Passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.p4execHTTP(w, r, &req)
}

func (s *Server) p4execHTTP(w http.ResponseWriter, r *http.Request, req *p4ExecRequest) {
	logger := s.Logger.Scoped("p4exec")

	// Flush writes more aggressively than standard net/http so that clients
	// with a context deadline see as much partial response body as possible.
	if fw := newFlushingResponseWriter(logger, w); fw != nil {
		w = fw
		defer fw.Close()
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()

	w.Header().Set("Trailer", "X-Exec-Error")
	w.Header().Add("Trailer", "X-Exec-Exit-Status")
	w.Header().Add("Trailer", "X-Exec-Stderr")
	w.WriteHeader(http.StatusOK)

	execStatus := s.p4Exec(ctx, logger, req, r.UserAgent(), w)
	w.Header().Set("X-Exec-Error", errorString(execStatus.Err))
	w.Header().Set("X-Exec-Exit-Status", strconv.Itoa(execStatus.ExitStatus))
	w.Header().Set("X-Exec-Stderr", execStatus.Stderr)

}

func (s *Server) p4Exec(ctx context.Context, logger log.Logger, req *p4ExecRequest, userAgent string, w io.Writer) execStatus {
	start := time.Now()
	var cmdStart time.Time // set once we have ensured commit
	exitStatus := executil.UnsetExitStatus
	var stdoutN, stderrN int64
	var status string
	var execErr error

	// Instrumentation
	{
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		args := strings.Join(req.Args, " ")

		var tr trace.Trace
		tr, ctx = trace.New(ctx, "p4exec."+cmd, attribute.String("port", req.P4Port))
		tr.SetAttributes(attribute.String("args", args))
		logger = logger.WithTrace(trace.Context(ctx))

		execRunning.WithLabelValues(cmd).Inc()
		defer func() {
			tr.AddEvent("done",
				attribute.String("status", status),
				attribute.Int64("stdout", stdoutN),
				attribute.Int64("stderr", stderrN),
			)
			tr.SetError(execErr)
			tr.End()

			duration := time.Since(start)
			execRunning.WithLabelValues(cmd).Dec()
			execDuration.WithLabelValues(cmd, status).Observe(duration.Seconds())

			var cmdDuration time.Duration
			if !cmdStart.IsZero() {
				cmdDuration = time.Since(cmdStart)
			}

			isSlow := cmdDuration > 30*time.Second
			if honey.Enabled() || traceLogs || isSlow {
				act := actor.FromContext(ctx)
				ev := honey.NewEvent("gitserver-p4exec")
				ev.SetSampleRate(honeySampleRate(cmd, act))
				ev.AddField("p4port", req.P4Port)
				ev.AddField("cmd", cmd)
				ev.AddField("args", args)
				ev.AddField("actor", act.UIDString())
				ev.AddField("client", userAgent)
				ev.AddField("duration_ms", duration.Milliseconds())
				ev.AddField("stdout_size", stdoutN)
				ev.AddField("stderr_size", stderrN)
				ev.AddField("exit_status", exitStatus)
				ev.AddField("status", status)
				if execErr != nil {
					ev.AddField("error", execErr.Error())
				}
				if !cmdStart.IsZero() {
					ev.AddField("cmd_duration_ms", cmdDuration.Milliseconds())
				}

				if traceID := trace.ID(ctx); traceID != "" {
					ev.AddField("traceID", traceID)
					ev.AddField("trace", trace.URL(traceID, conf.DefaultClient()))
				}

				_ = ev.Send()

				if traceLogs {
					logger.Debug("TRACE gitserver p4exec", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
				}
				if isSlow {
					logger.Warn("Long p4exec request", log.Object("ev.Fields", mapToLoggerField(ev.Fields())...))
				}
			}
		}()
	}

	p4home, err := gitserverfs.MakeP4HomeDir(s.ReposDir)
	if err != nil {
		return execStatus{ExitStatus: -1, Err: err}
	}

	var stderrBuf bytes.Buffer
	stdoutW := &writeCounter{w: w}
	stderrW := &writeCounter{w: &limitWriter{W: &stderrBuf, N: 1024}}

	cmdStart = time.Now()
	cmd := exec.CommandContext(ctx, "p4", req.Args...)
	cmd.Env = append(os.Environ(),
		"P4PORT="+req.P4Port,
		"P4USER="+req.P4User,
		"P4PASSWD="+req.P4Passwd,
		"HOME="+p4home,
	)
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	exitStatus, execErr = executil.RunCommand(ctx, s.RecordingCommandFactory.Wrap(ctx, s.Logger, cmd))

	status = strconv.Itoa(exitStatus)
	stdoutN = stdoutW.n
	stderrN = stderrW.n

	stderr := stderrBuf.String()

	return execStatus{
		ExitStatus: exitStatus,
		Stderr:     stderr,
		Err:        execErr,
	}
}

// p4ExecRequest is a request to execute a p4 command with given arguments.
//
// Note that this request is deserialized by both gitserver and the frontend's
// internal proxy route and any major change to this structure will need to be
// reconciled in both places.
type p4ExecRequest struct {
	P4Port   string   `json:"p4port"`
	P4User   string   `json:"p4user"`
	P4Passwd string   `json:"p4passwd"`
	Args     []string `json:"args"`
}

func (r *p4ExecRequest) ToProto() *proto.P4ExecRequest {
	return &proto.P4ExecRequest{
		P4Port:   r.P4Port,
		P4User:   r.P4User,
		P4Passwd: r.P4Passwd,
		Args:     stringsToByteSlices(r.Args),
	}
}

func (r *p4ExecRequest) FromProto(p *proto.P4ExecRequest) {
	*r = p4ExecRequest{
		P4Port:   p.GetP4Port(),                    //nolint:staticcheck
		P4User:   p.GetP4User(),                    //nolint:staticcheck
		P4Passwd: p.GetP4Passwd(),                  //nolint:staticcheck
		Args:     byteSlicesToStrings(p.GetArgs()), //nolint:staticcheck
	}
}

func stringsToByteSlices(in []string) [][]byte {
	res := make([][]byte, len(in))
	for i, s := range in {
		res[i] = []byte(s)
	}
	return res
}
