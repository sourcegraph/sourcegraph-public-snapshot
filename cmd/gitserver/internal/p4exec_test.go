package internal

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestServer_handleP4Exec(t *testing.T) {
	defaultMockRunCommand := func(ctx context.Context, cmd *exec.Cmd) (int, error) {
		switch cmd.Args[1] {
		case "users":
			_, _ = cmd.Stdout.Write([]byte("admin <admin@joe-perforce-server> (admin) accessed 2021/01/31"))
			_, _ = cmd.Stderr.Write([]byte("teststderr"))
			return 42, errors.New("the answer to life the universe and everything")
		}
		return 0, nil
	}

	t.Cleanup(func() {
		executil.UpdateRunCommandMock(nil)
	})

	startServer := func(t *testing.T) (handler http.Handler, client proto.GitserverServiceClient, cleanup func()) {
		t.Helper()

		logger := logtest.Scoped(t)

		s := &Server{
			Logger:                  logger,
			ReposDir:                t.TempDir(),
			ObservationCtx:          observation.TestContextTB(t),
			skipCloneForTests:       true,
			DB:                      dbmocks.NewMockDB(),
			RecordingCommandFactory: wrexec.NewNoOpRecordingCommandFactory(),
			Locker:                  NewRepositoryLocker(),
		}

		server := defaults.NewServer(logger)
		proto.RegisterGitserverServiceServer(server, &GRPCServer{Server: s})
		handler = grpc.MultiplexHandlers(server, s.Handler())

		srv := httptest.NewServer(handler)

		u, _ := url.Parse(srv.URL)
		conn, err := defaults.Dial(u.Host, logger.Scoped("gRPC client"))
		if err != nil {
			t.Fatalf("failed to dial: %v", err)
		}

		client = proto.NewGitserverServiceClient(conn)

		return handler, client, func() {
			srv.Close()
			conn.Close()
			server.Stop()
		}
	}

	t.Run("gRPC", func(t *testing.T) {
		readAll := func(execClient proto.GitserverService_P4ExecClient) ([]byte, error) {
			var buf bytes.Buffer
			for {
				resp, err := execClient.Recv()
				if errors.Is(err, io.EOF) {
					return buf.Bytes(), nil
				}

				if err != nil {
					return buf.Bytes(), err
				}

				_, err = buf.Write(resp.GetData())
				if err != nil {
					t.Fatalf("failed to write data: %v", err)
				}
			}
		}

		t.Run("users", func(t *testing.T) {
			executil.UpdateRunCommandMock(defaultMockRunCommand)

			_, client, closeFunc := startServer(t)
			t.Cleanup(closeFunc)

			stream, err := client.P4Exec(context.Background(), &proto.P4ExecRequest{
				Args: [][]byte{[]byte("users")},
			})
			if err != nil {
				t.Fatalf("failed to call P4Exec: %v", err)
			}

			data, err := readAll(stream)
			s, ok := status.FromError(err)
			if !ok {
				t.Fatal("received non-status error from p4exec call")
			}

			if diff := cmp.Diff("the answer to life the universe and everything", s.Message()); diff != "" {
				t.Fatalf("unexpected error in stream (-want +got):\n%s", diff)
			}

			expectedData := []byte("admin <admin@joe-perforce-server> (admin) accessed 2021/01/31")

			if diff := cmp.Diff(expectedData, data); diff != "" {
				t.Fatalf("unexpected data (-want +got):\n%s", diff)
			}
		})

		t.Run("empty request", func(t *testing.T) {
			executil.UpdateRunCommandMock(defaultMockRunCommand)

			_, client, closeFunc := startServer(t)
			t.Cleanup(closeFunc)

			stream, err := client.P4Exec(context.Background(), &proto.P4ExecRequest{})
			if err != nil {
				t.Fatalf("failed to call P4Exec: %v", err)
			}

			_, err = readAll(stream)
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument error, got %v", err)
			}
		})

		t.Run("disallowed command", func(t *testing.T) {

			executil.UpdateRunCommandMock(defaultMockRunCommand)

			_, client, closeFunc := startServer(t)
			t.Cleanup(closeFunc)

			stream, err := client.P4Exec(context.Background(), &proto.P4ExecRequest{
				Args: [][]byte{[]byte("bad_command")},
			})
			if err != nil {
				t.Fatalf("failed to call P4Exec: %v", err)
			}

			_, err = readAll(stream)
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument error, got %v", err)
			}
		})

		t.Run("context cancelled", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			executil.UpdateRunCommandMock(func(ctx context.Context, _ *exec.Cmd) (int, error) {
				// fake a context cancellation that occurs while the process is running

				cancel()
				return 0, ctx.Err()
			})

			_, client, closeFunc := startServer(t)
			t.Cleanup(closeFunc)

			stream, err := client.P4Exec(ctx, &proto.P4ExecRequest{
				Args: [][]byte{[]byte("users")},
			})
			if err != nil {
				t.Fatalf("failed to call P4Exec: %v", err)
			}

			_, err = readAll(stream)
			if !(errors.Is(err, context.Canceled) || status.Code(err) == codes.Canceled) {
				t.Fatalf("expected context cancelation error, got %v", err)
			}
		})

	})

	t.Run("HTTP", func(t *testing.T) {

		tests := []Test{
			{
				Name:         "Command",
				Request:      newRequest("POST", "/p4-exec", strings.NewReader(`{"args": ["users"]}`)),
				ExpectedCode: http.StatusOK,
				ExpectedBody: "admin <admin@joe-perforce-server> (admin) accessed 2021/01/31",
				ExpectedTrailers: http.Header{
					"X-Exec-Error":       {"the answer to life the universe and everything"},
					"X-Exec-Exit-Status": {"42"},
					"X-Exec-Stderr":      {"teststderr"},
				},
			},
			{
				Name:         "Error",
				Request:      newRequest("POST", "/p4-exec", strings.NewReader(`{"args": ["bad_command"]}`)),
				ExpectedCode: http.StatusBadRequest,
				ExpectedBody: "subcommand \"bad_command\" is not allowed",
			},
			{
				Name:         "EmptyBody",
				Request:      newRequest("POST", "/p4-exec", nil),
				ExpectedCode: http.StatusBadRequest,
				ExpectedBody: `EOF`,
			},
			{
				Name:         "EmptyInput",
				Request:      newRequest("POST", "/p4-exec", strings.NewReader("{}")),
				ExpectedCode: http.StatusBadRequest,
				ExpectedBody: `args must be greater than or equal to 1`,
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				executil.UpdateRunCommandMock(defaultMockRunCommand)

				handler, _, closeFunc := startServer(t)
				t.Cleanup(closeFunc)

				w := httptest.ResponseRecorder{Body: new(bytes.Buffer)}
				handler.ServeHTTP(&w, test.Request)

				res := w.Result()
				if res.StatusCode != test.ExpectedCode {
					t.Errorf("wrong status: expected %d, got %d", test.ExpectedCode, w.Code)
				}

				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				if strings.TrimSpace(string(body)) != test.ExpectedBody {
					t.Errorf("wrong body: expected %q, got %q", test.ExpectedBody, string(body))
				}

				for k, v := range test.ExpectedTrailers {
					if got := res.Trailer.Get(k); got != v[0] {
						t.Errorf("wrong trailer %q: expected %q, got %q", k, v[0], got)
					}
				}
			})
		}
	})
}
