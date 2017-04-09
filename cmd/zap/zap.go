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
	"sync"
	"syscall"

	// Import for side effect of setting SGPATH env var.
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/websocket"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	zapgit "github.com/sourcegraph/zap/pkg/git"
	"github.com/sourcegraph/zap/server"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
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
	zapServer := server.New(zapServerBackend)
	if err := zapServer.Start(ctx); err != nil {
		stdlog.Fatal(err)
	}
	go stdlog.Fatal(http.Serve(lis, traceutil.TraceRoute(actor.TrustedActorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: Upgrade: %s [client: %s]\n", err, r.RemoteAddr)
			http.Error(w, "WebSocket upgrade error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		<-zapServer.Accept(withCanAccessRepos(r.Context()), websocketjsonrpc2.NewObjectStream(c))
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

// canAccessRepos is stored in a JSON-RPC connection's parent context
// and remembers repos that (ServerBackend).CanAccessRepo has
// succeeded on.
type canAccessRepos struct {
	sync.Mutex
	// Repos is the set of repos that yielded a "true" result from
	// (ServerBackend).CanAccessRepo during this connection.
	repos map[string]struct{}
}

type contextKey int

const canAccessReposContextKey contextKey = iota

// withCanAccessRepos augments this context with the ability to store
// repos that a (ServerBackend).CanAccessRepo check has succeeded on.
func withCanAccessRepos(ctx context.Context) context.Context {
	return context.WithValue(ctx, canAccessReposContextKey, &canAccessRepos{repos: map[string]struct{}{}})
}

func cachedCanAccessRepo(ctx context.Context, repo string) bool {
	v := ctx.Value(canAccessReposContextKey).(*canAccessRepos)
	v.Lock()
	defer v.Unlock()
	_, ok := v.repos[repo]
	return ok
}

func cacheSetCanAccessRepo(ctx context.Context, repo string, ok bool) {
	if !ok {
		panic("per-connection CanAccessRepo caching only supports positive results")
	}
	v := ctx.Value(canAccessReposContextKey).(*canAccessRepos)
	v.Lock()
	defer v.Unlock()
	v.repos[repo] = struct{}{}
}

var zapServerBackend = zapgit.ServerBackend{
	CanAccessRepo: func(ctx context.Context, log log.Logger, repo string) (ok bool, err error) {
		if cachedCanAccessRepo(ctx, repo) {
			return true, nil
		}

		logResult := func(ok bool, err error) {
			actor := actor.FromContext(ctx)
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
		if _, err := backend.Repos.GetByURI(ctx, repo); err != nil {
			return false, err
		}
		ok = true
		cacheSetCanAccessRepo(ctx, repo, ok)
		return ok, nil
	},
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
