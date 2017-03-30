package main

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	// Import for side effect of setting SGPATH env var.
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/websocket"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/sourcegraph/zap"
	zapgit "github.com/sourcegraph/zap/pkg/git"
	zapgitutil "github.com/sourcegraph/zap/pkg/gitutil"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
)

/*
ZAP SERVER ARCHITECTURE

     +---------------------+
     |                     |
     | User's editor w/Zap |
     |                     |
     +-----^---------------+
        |  |
        |  |
        |  |
 +------v------------------------+      +-----------------------------------+
 |                               |      |                                   |
 |  User's Zap local server      |      | User's browser at Sourcegraph.com |
 |                               |      |                                   |
 +-----^-------------------------+      +-------^---------------------------+
    |  |                                     |  |
    |  |                                     |  |
    |  |                                     |  |
 +--v-------------+                          |  |
 |                <--------------------------+  |
 | HTTP /.api/zap |                             |
 |                +-----------------------------+
 +-----^----------+
    |  |
    |  |
    |  |
 +--v---------------------+
 |                        |
 | Zap server (this file) |
 |                        |
 +------------------------+


LIFECYCLE OF A ZAP OPERATION

1. User opens a file "f" containing "abc".
2. User types "x" in the file in their editor.
3. Their editor's Zap editor extension notices the change and sends op {edit: {f: [3, "x"]}} to their Zap local server (running on their machine).
4. Their Zap local server sends the op to the server specified by its ZAP_SERVER env var, which points to Sourcegraph's HTTP/WebSocket https://sourcegraph.com/.api/zap endpoint.
5. If the Zap local server isn't already in an active WebSocket connection with Sourcegraph: The /.api/zap HTTP endpoint forwards the connection to this Zap remote server (in this file), which performs the WebSocket upgrade with the user's Zap local server.
6. The Zap remote server applies the op in-memory and stores a snapshot on the gitserver.
7. The Zap remote server broadcasts the op to other connected clients and sends an ack to the original user's Zap local server.
8. The Zap local server forwards the ack to the user's editor.

*/

var (
	profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")
	listenURLStr = os.ExpandEnv(env.Get("ZAP_SERVER_LISTEN", "ws://${SGPATH}/zap", "zap server listen URL (ws:///abspath or ws://host:port)"))
)

var websocketUpgrader = websocket.Upgrader{
	// We already do an origin check in httpapi websocket proxy, so we can accept
	// requests here without checking. This is fine since this service should be
	// deployed behind the firewall.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// dogfoodGitClient is the git client we use, just while dogfooding. We only
// use it in the access check to ensure we have cloned the repo onto
// gitserver-zap. We can't use DefaultClient, since that has NoCreds set to
// true.
var dogfoodGitClient = gitserver.NewClient(strings.Fields(os.Getenv("SRC_GIT_SERVERS")))

func main() {
	env.Lock()
	env.HandleHelpFlag()
	traceutil.InitTracer()
	gitserver.DefaultClient.NoCreds = true
	auth.InitSessionStore(conf.AppURL.Scheme == "https")

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
		<-c
		os.Exit(0)
	}()

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
		stdlog.Printf("Profiler available on %s/pprof", profBindAddr)
	}

	addr, lis, err := listen(listenURLStr)
	if err != nil {
		stdlog.Fatal("zap:", err)
	}
	fmt.Fprintln(os.Stderr, "zap: listening on", addr)

	ctx := context.Background()
	zapServer := zap.NewServer(zapServerBackend)
	zapServer.Start(ctx)
	go stdlog.Fatal(http.Serve(lis, httptrace.TraceRoute(auth.TrustedActorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: Upgrade: %s [client: %s]\n", err, r.RemoteAddr)
			http.Error(w, "WebSocket upgrade error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		<-zapServer.Accept(r.Context(), websocketjsonrpc2.NewObjectStream(c))
	})))))
	select {}
}

func listen(urlStr string) (string, net.Listener, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", nil, err
	}
	if u.Scheme != "ws" {
		return "", nil, fmt.Errorf("bad listen URL %s (must be ws:///abspath or ws://host:port)", urlStr)
	}
	if u.Host == "" {
		_ = os.Remove(u.Path)
		lis, err := net.Listen("unix", u.Path)
		if err != nil {
			return "", nil, err
		}
		return urlStr, lis, nil
	}
	lis, err := net.Listen("tcp", u.Host)
	if err != nil {
		return "", nil, err
	}
	return "ws://" + lis.Addr().String(), lis, nil
}

var zapServerBackend = zapgit.ServerBackend{
	CanAccessRepo: func(ctx context.Context, log log.Logger, repo string) (ok bool, err error) {
		logResult := func(ok bool, err error) {
			actor := auth.ActorFromContext(ctx)
			var f func(string, ...interface{})
			if ok {
				f = log15.Info
			} else if err != nil {
				f = log15.Error
			} else {
				f = log15.Warn
			}
			f("Zap: CanAccessRepo", "repo", repo, "login", actor.Login, "uid", actor.UID, "canAccess", ok, "err", err)
		}
		defer func() {
			logResult(ok, err)
		}()

		// A Repos.GetByURI call with a nil error indicates the actor
		// has access to the repo.
		//
		// 🚨 SECURITY: While Zap is in dogfooding, we are OK allowing 🚨
		// anyone with read access to also have write access to Zap
		// repos. Currently we have no way to allow Zap reads but not
		// writes.
		sRepo, err := backend.Repos.GetByURI(ctx, repo)
		if err != nil {
			return false, err
		}

		// Dogfooding: Do a quick git command just to ensure we have
		// the repo cloned. Remove before launch.
		go func() {
			cmd := dogfoodGitClient.Command("git", "rev-parse", "--show-toplevel")
			cmd.Repo = sRepo
			cmd.Run(ctx)
		}()

		return true, nil
	},
	OpenBareRepo: func(ctx context.Context, log log.Logger, repo string) (zapgit.ServerRepo, error) {
		actor := auth.ActorFromContext(ctx)
		log15.Info("Zap: OpenRepo", "repo", repo, "login", actor.Login, "uid", actor.UID)

		_, err := backend.Repos.GetByURI(ctx, repo)
		if err != nil {
			return nil, err
		}
		return zapgitutil.BareRepo{
			GitExecutor: gitserverExecutor{
				repoPath: repo,
			},
		}, nil
	},
	CanAutoCreateRepo: func() bool { return true },
}

type gitserverExecutor struct {
	repoPath string
}

func (e gitserverExecutor) Exec(input []byte, args ...string) ([]byte, error) {
	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = &sourcegraph.Repo{URI: e.repoPath}
	cmd.Input = input
	stdout, stderr, err := cmd.DividedOutput(context.Background())
	if err != nil {
		return nil, gitError(args, err, stderr)
	}
	if len(stderr) != 0 {
		return nil, gitError(args, errors.New("unexpected output on stderr"), stderr)
	}
	return stdout, nil
}

func gitError(args []string, err error, stderr []byte) error {
	return fmt.Errorf("command failed: git %s: %s\n%s", strings.Join(args, " "), err, stderr)
}
