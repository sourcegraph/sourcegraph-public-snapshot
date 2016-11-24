package cli

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/keegancsmith/tmpfriend"
	"sourcegraph.com/sourcegraph/sourcegraph/app"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	app_router "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/loghandlers"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/middleware"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/graphstoreutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/sysreq"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/events"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	"sourcegraph.com/sourcegraph/sourcegraph/services/httpapi"
	"sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/services/repoupdater"
	srclib "sourcegraph.com/sourcegraph/srclib/cli"
)

var (
	logLevel       = env.Get("SRC_LOG_LEVEL", "info", "upper log level to restrict log output to (dbug, dbug-dev, info, warn, error, crit)")
	trace          = env.Get("SRC_LOG_TRACE", "HTTP", "space separated list of trace logs to show. Options: all, HTTP, build, github")
	traceThreshold = env.Get("SRC_LOG_TRACE_THRESHOLD", "", "show traces that take longer than this")

	httpAddr  = env.Get("SRC_HTTP_ADDR", ":3080", "HTTP listen address for app, REST API, and gRPC API")
	httpsAddr = env.Get("SRC_HTTPS_ADDR", ":3443", "HTTPS (TLS) listen address for app, REST API, and gRPC API")

	profBindAddr = env.Get("SRC_PROF_HTTP", ":6060", "net/http/pprof http bind address")

	appURL     = env.Get("SRC_APP_URL", "http://<http-addr>", "publicly accessible URL to web app (e.g., what you type into your browser)")
	enableHSTS = env.Get("SG_ENABLE_HSTS", "false", "enable HTTP Strict Transport Security")

	certFile = env.Get("SRC_TLS_CERT", "", "certificate file for TLS")
	keyFile  = env.Get("SRC_TLS_KEY", "", "key file for TLS")

	idKeyData = env.Get("SRC_ID_KEY_DATA", "", "identity key file data")

	reposDir   = os.ExpandEnv(env.Get("SRC_REPOS_DIR", "$SGPATH/repos", "root dir containing repos"))
	gitservers = env.Get("SRC_GIT_SERVERS", "", "addresses of the remote gitservers; a local gitserver process is used by default")

	graphstoreRoot = os.ExpandEnv(env.Get("SRC_GRAPHSTORE_ROOT", "$SGPATH/repos", "root dir, HTTP VFS (http[s]://...), or S3 bucket (s3://...) in which to store graph data"))
)

func init() {
	srclib.CacheLocalRepo = false
}

func configureAppURL() (*url.URL, error) {
	var hostPort string
	if strings.HasPrefix(httpAddr, ":") {
		// Prepend localhost if HTTP listen addr is just a port.
		hostPort = "localhost" + httpAddr
	} else {
		hostPort = httpAddr
	}
	if appURL == "" {
		appURL = "http://<http-addr>"
	}
	appURL = strings.Replace(appURL, "<http-addr>", hostPort, -1)

	u, err := url.Parse(appURL)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func Main() error {
	log.SetFlags(0)
	log.SetPrefix("")

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			log.Print("Build information:")
			b, err := json.MarshalIndent(buildvar.All, "", "  ")
			if err != nil {
				return err
			}
			log.Print(string(b))
			log.Print()

			env.PrintHelp()

			log.Print()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			for _, st := range sysreq.Check(ctx, skippedSysReqs()) {
				log.Printf("%s:", st.Name)
				if st.OK() {
					log.Print("\tOK")
					continue
				}
				if st.Skipped {
					log.Print("\tSkipped")
					continue
				}
				if st.Problem != "" {
					log.Print("\t" + st.Problem)
				}
				if st.Err != nil {
					log.Printf("\tError: %s", st.Err)
				}
				if st.Fix != "" {
					log.Printf("\tPossible fix: %s", st.Fix)
				}
			}

			return nil
		}
	}

	cleanup := tmpfriend.SetupOrNOOP()
	defer cleanup()

	logHandler := log15.StderrHandler

	// We have some noisey debug logs, so to aid development we have a
	// special dbug level which excludes the noisey logs
	if logLevel == "dbug-dev" {
		logLevel = "dbug"
		logHandler = log15.FilterHandler(loghandlers.NotNoisey, logHandler)
	}

	// Filter trace logs
	d, _ := time.ParseDuration(traceThreshold)
	logHandler = log15.FilterHandler(loghandlers.Trace(strings.Fields(trace), d), logHandler)

	// Filter log output by level.
	lvl, err := log15.LvlFromString(logLevel)
	if err != nil {
		return err
	}
	log15.Root().SetHandler(log15.LvlFilterHandler(lvl, logHandler))

	traceutil.InitTracer()

	// Don't proceed if system requirements are missing, to avoid
	// presenting users with a half-working experience.
	if err := checkSysReqs(context.Background(), os.Stderr); err != nil {
		return err
	}

	log15.Debug("GraphStore", "at", graphstoreRoot)

	for _, f := range cli.ServeInit {
		f()
	}

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
		log15.Debug("Profiler available", "on", fmt.Sprintf("%s/pprof", profBindAddr))
	}

	app.Init()

	conf.AppURL, err = configureAppURL()
	if err != nil {
		return err
	}

	backend.SetGraphStore(graphstoreutil.New(graphstoreRoot, nil))

	// Server identity keypair
	if idKeyData != "" {
		auth.ActiveIDKey, err = auth.FromString(idKeyData)
		if err != nil {
			return err
		}
	} else {
		log15.Warn("Using default ID key.")
	}

	runGitserver()

	sm := http.NewServeMux()
	newRouter := func() *mux.Router {
		router := mux.NewRouter()
		// httpctx.Base will clear the context for us
		router.KeepContext = true
		return router
	}
	subRouter := func(r *mux.Route) *mux.Router {
		router := r.Subrouter()
		// httpctx.Base will clear the context for us
		router.KeepContext = true
		return router
	}
	sm.Handle("/.api/", gziphandler.GzipHandler(httpapi.NewHandler(router.New(subRouter(newRouter().PathPrefix("/.api/"))))))
	sm.Handle("/", gziphandler.GzipHandler(handlerutil.NewHandlerWithCSRFProtection(app.NewHandler(app_router.New(newRouter())))))
	assets.Mount(sm)

	if (certFile != "" || keyFile != "") && httpsAddr == "" {
		return errors.New("HTTPS listen address must be specified if TLS cert and key are set")
	}
	useTLS := certFile != "" || keyFile != ""

	if useTLS && conf.AppURL.Scheme == "http" {
		log15.Warn("TLS is enabled but app url scheme is http", "appURL", conf.AppURL)
	}

	if !useTLS && conf.AppURL.Scheme == "https" {
		log15.Warn("TLS is disabled but app url scheme is https", "appURL", conf.AppURL)
	}

	mw := []handlerutil.Middleware{middleware.HealthCheck, middleware.RealIP, middleware.NoCacheByDefault}
	if v, _ := strconv.ParseBool(enableHSTS); v {
		mw = append(mw, middleware.StrictTransportSecurity)
	}
	mw = append(mw, middleware.SecureHeader)
	mw = append(mw, httptrace.Middleware)
	mw = append(mw, middleware.BlackHole)
	mw = append(mw, middleware.SourcegraphComGoGetHandler)

	// Start background workers that receive input from main app.
	//
	// It's safe (and better) to start this before starting the
	// HTTP(S) web server to avoid a brief moment where the web server
	// is started, but the listeners haven't started yet.
	//
	//
	// Start event listeners.
	initializeEventListeners()

	srv := &http.Server{
		Handler:      handlerutil.WithMiddleware(sm, mw...),
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	// Start HTTP server.
	if httpAddr != "" {
		l, err := net.Listen("tcp", httpAddr)
		if err != nil {
			return err
		}

		log15.Debug("HTTP running", "on", httpAddr)
		go func() { log.Fatal(srv.Serve(l)) }()
	}

	// Start HTTPS server.
	if useTLS && httpsAddr != "" {
		l, err := net.Listen("tcp", httpsAddr)
		if err != nil {
			return err
		}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}

		l = tls.NewListener(l, &tls.Config{
			NextProtos:   []string{"h2"},
			Certificates: []tls.Certificate{cert},
		})

		log15.Debug("HTTPS running", "on", httpsAddr)
		go func() { log.Fatal(srv.Serve(l)) }()
	}

	// Connection test
	log15.Info(fmt.Sprintf("✱ Sourcegraph running at %s", appURL))

	// Start background repo updater worker.
	repoUpdaterCtx, err := authenticateScopedContext(context.Background(), []string{"internal:repoupdater"})
	if err != nil {
		return err
	}
	repoupdater.RepoUpdater.Start(repoUpdaterCtx)

	// HACK(keegancsmith) async is the only user of this at the moment,
	// but other background workers that need access to the store will
	// likely pop up in the future. We need to make this less hacky
	internalServerCtx := func(name string) (context.Context, error) {
		ctx := context.Background()
		scope := "internal:" + name
		ctx = auth.WithActor(ctx, &auth.Actor{Scope: map[string]bool{scope: true}})
		ctx, err = authenticateScopedContext(ctx, []string{scope})
		if err != nil {
			return nil, err
		}
		return ctx, nil
	}

	// Start background async workers. It is a service, so needs the
	// stores setup in the context
	asyncCtx, err := internalServerCtx("async")
	if err != nil {
		return err
	}
	backend.StartAsyncWorkers(asyncCtx)

	select {}
}

// authenticateScopedContext adds a token with the specified scope to the given
// context. This context can only make gRPC calls that are permitted for the given
// scope. See the accesscontrol package for information about different scopes.
func authenticateScopedContext(ctx context.Context, scopes []string) (context.Context, error) {
	scopeMap := make(map[string]bool)
	for _, s := range scopes {
		scopeMap[s] = true
	}
	a := &auth.Actor{
		Scope: scopeMap,
	}
	ctx = github.NewContextWithAuthedClient(ctx)
	return auth.WithActor(ctx, a), nil
}

// initializeEventListeners creates special scoped contexts and passes them to
// event listeners.
func initializeEventListeners() {
	for _, l := range events.GetRegisteredListeners() {
		listenerCtx, err := authenticateScopedContext(auth.WithActor(context.Background(), &auth.Actor{}), l.Scopes())
		if err != nil {
			log.Fatalf("Could not initialize listener context: %v", err)
		} else {
			l.Start(listenerCtx)
		}
	}
}

// runGitserver either connects to gitservers specified in gitservers, if any.
// Otherwise it starts a single local gitserver and connects to it.
func runGitserver() {
	if gitservers == "none" {
		return
	}

	gitservers := strings.Fields(gitservers)
	if len(gitservers) != 0 {
		for _, addr := range gitservers {
			gitserver.DefaultClient.Connect(addr)
		}
		return
	}

	stdoutReader, stdoutWriter := io.Pipe()
	go func() {
		cmd := exec.Command("gitserver")
		cmd.Env = append(os.Environ(),
			"SRC_AUTO_TERMINATE=true",
			"SRC_REPOS_DIR="+os.ExpandEnv(reposDir),
		)
		_, err := cmd.StdinPipe() // keep stdin from closing
		if err != nil {
			log.Fatalf("git-server failed: %s", err)
		}
		cmd.Stdout = io.MultiWriter(os.Stdout, stdoutWriter)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("git-server failed: %s", err)
		}
		log.Fatal("git-server has exited")
	}()

	r := bufio.NewReader(stdoutReader)
	line, err := r.ReadString('\n')
	if err != nil {
		log.Fatalf("git-server stdout read failed: %s", err)
	}
	addr := line[strings.LastIndexByte(line, ' ')+1 : len(line)-1]
	go io.Copy(ioutil.Discard, stdoutReader) // drain pipe

	gitserver.DefaultClient.Connect(addr)
}
