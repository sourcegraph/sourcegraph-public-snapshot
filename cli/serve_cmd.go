package cli

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/keegancsmith/tmpfriend"
	"github.com/soheilhy/cmux"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/sourcegraph/app"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	app_router "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/sharedsecret"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/client"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/middleware"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/srccmd"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/httpapi"
	"sourcegraph.com/sourcegraph/sourcegraph/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/snapshotprof"
	"sourcegraph.com/sourcegraph/sourcegraph/server"
	"sourcegraph.com/sourcegraph/sourcegraph/services/events"
	"sourcegraph.com/sourcegraph/sourcegraph/services/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/services/worker"
	"sourcegraph.com/sourcegraph/sourcegraph/util/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/util/metricutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/statsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
	"sourcegraph.com/sqs/pbtypes"
)

// Stripped down help message presented to most users (full help message can be
// gotten with -a or --help-all).
var shortHelpMessage = `Usage:
  src serve [serve-OPTIONS]

Starts an HTTP server serving the app and API.

[serve command options]
          --http-addr=                           HTTP listen address for app, REST API, and gRPC API (:3080)
          --https-addr=                          HTTPS (TLS) listen address for app, REST API, and gRPC API (:3443)
          --prof-http=BIND-ADDR                  net/http/pprof http bind address (:6060)
          --app-url=                             publicly accessible URL to web app (e.g., what you type into your browser) (http://<http-addr>)
          --reload                               reload templates, etc. on each request (dev mode)
          --no-worker                            do not start background worker
          --tls-cert=                            certificate file (for TLS)
          --tls-key=                             key file (for TLS)
      -i, --id-key=                              identity key file
          --id-key-data=                         identity key file data (overrides -i/--id-key) [$SRC_ID_KEY_DATA]
          --dequeue-msec=                        if no builds are dequeued, sleep up to this many msec before trying again (1000)

    Help Options:
      -h, --help                                 Show common serve flags
      -a, --help-all                             Show all serve flags

    Local:
          --local.clcache=                       how often to refresh the commit-log cache in seconds; if 0, then no cache is used (0)
          --local.clcachesize=                   number of commits to cache on refresh (500)
`

func init() {
	// We will register our own custom help group.
	cli.CustomHelpCmds = append(cli.CustomHelpCmds, "serve")

	c, err := cli.CLI.AddCommand("serve",
		"start web server",
		`
Starts an HTTP server serving the app and API.`,
		&serveCmdInst,
	)
	if err != nil {
		log.Fatal(err)
	}
	c.SubcommandsOptional = true
	cli.Serve = c

	// Build the group.
	var help struct {
		ShowHelp    func() error `short:"h" long:"help" description:"Show common serve flags"`
		ShowHelpAll func() error `short:"a" long:"help-all" description:"Show all serve flags"`
	}
	help.ShowHelp = func() error {
		return &flags.Error{
			Type:    flags.ErrHelp,
			Message: shortHelpMessage,
		}
	}
	help.ShowHelpAll = func() error {
		var b bytes.Buffer
		cli.CLI.WriteHelp(&b)
		return &flags.Error{
			Type:    flags.ErrHelp,
			Message: b.String(),
		}
	}

	// Add the group to the command.
	_, err = c.AddGroup("Help Options", "", &help)
	if err != nil {
		log.Fatal(err)
	}
}

// ServeCmdPrivate holds the parameters containing private data about the
// instance. These fields will not be forwarded with the other metrics.
type ServeCmdPrivate struct {
	CertFile string `long:"tls-cert" description:"certificate file (for TLS)" env:"SRC_TLS_CERT"`
	KeyFile  string `long:"tls-key" description:"key file (for TLS)" env:"SRC_TLS_KEY"`

	IDKeyFile string `short:"i" long:"id-key" description:"identity key file" default:"$SGPATH/id.pem" env:"SRC_ID_KEY"`
	IDKeyData string `long:"id-key-data" description:"identity key file data (overrides -i/--id-key)" env:"SRC_ID_KEY_DATA"`
}

var serveCmdInst ServeCmd

type ServeCmd struct {
	HTTPAddr  string `long:"http-addr" default:":3080" description:"HTTP listen address for app, REST API, and gRPC API" env:"SRC_HTTP_ADDR"`
	HTTPSAddr string `long:"https-addr" default:":3443" description:"HTTPS (TLS) listen address for app, REST API, and gRPC API" env:"SRC_HTTPS_ADDR"`

	ProfBindAddr string `long:"prof-http" default:":6060" description:"net/http/pprof http bind address" value-name:"BIND-ADDR" env:"SRC_PROF_HTTP"`

	AppURL string `long:"app-url" default:"http://<http-addr>" description:"publicly accessible URL to web app (e.g., what you type into your browser)" env:"SRC_APP_URL"`

	RedirectToHTTPS bool `long:"app.redirect-to-https" description:"redirect HTTP requests to the equivalent HTTPS URL" env:"SG_FORCE_HTTPS"`

	NoWorker bool `long:"no-worker" description:"do not start background worker" env:"SRC_NO_WORKER"`

	// Flags containing sensitive information must be added to this struct.
	ServeCmdPrivate

	Prefetch bool `long:"prefetch" description:"prefetch directory children" env:"SRC_PREFETCH"`

	worker.WorkCmd

	GraphStoreOpts `group:"Graph data storage (defs, refs, etc.)" namespace:"graphstore"`

	NoInitialOnboarding bool `long:"no-initial-onboarding" description:"don't add sample repositories to server during initial server setup" env:"SRC_NO_INITIAL_ONBOARDING"`

	RegisterURL string `long:"register" description:"register this server as a client of another Sourcegraph server (empty to disable)" value-name:"URL" default:"https://sourcegraph.com"`

	ReposDir   string `long:"fs.repos-dir" description:"root dir containing repos" default:"$SGPATH/repos" env:"SRC_REPOS_DIR"`
	GitServers string `long:"git-servers" description:"addresses of the remote git servers; a local git server process is used by default" env:"SRC_GIT_SERVERS"`
}

func (c *ServeCmd) configureAppURL() (*url.URL, error) {
	var hostPort string
	if strings.HasPrefix(c.HTTPAddr, ":") {
		// Prepend localhost if HTTP listen addr is just a port.
		hostPort = "localhost" + c.HTTPAddr
	} else {
		hostPort = c.HTTPAddr
	}
	if c.AppURL == "" {
		c.AppURL = "http://<http-addr>"
	}
	c.AppURL = strings.Replace(c.AppURL, "<http-addr>", hostPort, -1)

	appURL, err := url.Parse(c.AppURL)
	if err != nil {
		return nil, err
	}

	// Endpoint defaults to the AppURL.
	if client.Endpoint.URL == "" {
		client.Endpoint.URL = appURL.String()

		// Reset client.Ctx to use new endpoint.
		client.Ctx = WithClientContext(context.Background())
	}

	return appURL, nil
}

func (c *ServeCmd) Execute(args []string) error {
	cleanup := tmpfriend.SetupOrNOOP()
	defer cleanup()

	logHandler := log15.StderrHandler
	if globalOpt.VerbosePkg != "" {
		logHandler = log15.MatchFilterHandler("pkg", globalOpt.VerbosePkg, log15.StderrHandler)
	}

	// We have some noisey debug logs, so to aid development we have a
	// special dbug level which excludes the noisey logs
	if globalOpt.LogLevel == "dbug-dev" {
		globalOpt.LogLevel = "dbug"
		logHandler = log15.FilterHandler(noiseyLogFilter, logHandler)
	}

	// Filter log output by level.
	lvl, err := log15.LvlFromString(globalOpt.LogLevel)
	if err != nil {
		return err
	}
	log15.Root().SetHandler(log15.LvlFilterHandler(lvl, logHandler))

	// Snapshotters allow us to regularly capture profile data for later
	// analysis. Not enabled by default since it is similiar to debug logs
	if p := os.Getenv("SG_SNAPSHOTPROF_PATH"); p != "" {
		runSnapshotProfiler(p)
	}

	// Don't proceed if system requirements are missing, to avoid
	// presenting users with a half-working experience.
	if err := checkSysReqs(client.Ctx, os.Stderr); err != nil {
		return err
	}

	// Clear auth specified on the CLI. If we didn't do this, then the
	// app, git, and HTTP API would all inherit the process's owner's
	// current auth. This is undesirable, unexpected, and could lead
	// to unintentionally leaking private info.
	client.Credentials.SetAccessToken("")
	client.Credentials.AuthFile = ""

	c.GraphStoreOpts.expandEnv()
	log15.Debug("GraphStore", "at", c.GraphStoreOpts.Root)

	for _, f := range cli.ServeInit {
		f()
	}

	if c.ProfBindAddr != "" {
		startDebugServer(c.ProfBindAddr)
	}

	var (
		sharedCtxFuncs []func(context.Context) context.Context
		serverCtxFuncs []func(context.Context) context.Context = cli.ServerContext
		clientCtxFuncs []func(context.Context) context.Context = cli.ClientContext
	)

	// graphstore
	serverCtxFuncs = append(serverCtxFuncs, c.GraphStoreOpts.context)

	app.Init()

	appURL, err := c.configureAppURL()
	if err != nil {
		return err
	}

	// Shared context setup between client and server.
	sharedCtxFunc := func(ctx context.Context) context.Context {
		for _, f := range sharedCtxFuncs {
			ctx = f(ctx)
		}
		ctx = conf.WithURL(ctx, appURL)
		return ctx
	}
	clientCtxFunc := func(ctx context.Context) context.Context {
		ctx = sharedCtxFunc(ctx)
		for _, f := range clientCtxFuncs {
			ctx = f(ctx)
		}
		ctx = WithClientContext(ctx)
		return ctx
	}
	serverCtxFunc := func(ctx context.Context) context.Context {
		ctx = sharedCtxFunc(ctx)
		for _, f := range serverCtxFuncs {
			ctx = f(ctx)
		}
		ctx = client.Endpoint.NewContext(ctx)
		return ctx
	}

	// Server identity keypair
	idKey, createdIDKey, err := c.generateOrReadIDKey()
	if err != nil {
		return err
	}
	log15.Debug("Sourcegraph server", "ID", idKey.ID)
	// Uncomment to add ID key prefix to log messages.
	// log.SetPrefix(bold(idKey.ID[:4] + ": "))

	sharedCtxFuncs = append(sharedCtxFuncs, func(ctx context.Context) context.Context {
		ctx = idkey.NewContext(ctx, idKey)
		return ctx
	})
	sharedSecretToken := oauth2.ReuseTokenSource(nil, sharedsecret.TokenSource(idKey))
	clientCtxFuncs = append(clientCtxFuncs, func(ctx context.Context) context.Context {
		return sourcegraph.WithCredentials(ctx, sharedSecretToken)
	})

	clientCtx := clientCtxFunc(context.Background())

	// Listen for events and periodically push them to analytics gateway.
	eventsutil.StartEventLogger(clientCtx, idKey.ID, 10*4096, 256, 10*time.Minute)

	c.runGitServer()

	sm := http.NewServeMux()
	for _, f := range cli.ServeMuxFuncs {
		f(sm)
	}
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
	sm.Handle("/", gziphandler.GzipHandler(app.NewHandlerWithCSRFProtection(app.NewHandler(app_router.New(newRouter())))))
	sm.Handle(assets.URLPathPrefix+"/", http.StripPrefix(assets.URLPathPrefix, assets.NewHandler(newRouter())))

	if (c.CertFile != "" || c.KeyFile != "") && c.HTTPSAddr == "" {
		return errors.New("HTTPS listen address (--https-addr) must be specified if TLS cert and key are set")
	}
	useTLS := c.CertFile != "" || c.KeyFile != ""

	if useTLS && appURL.Scheme == "http" {
		log15.Warn("TLS is enabled but app url scheme is http", "appURL", appURL)
	}

	if !useTLS && appURL.Scheme == "https" {
		log15.Warn("TLS is disabled but app url scheme is https", "appURL", appURL)
	}

	mw := []handlerutil.Middleware{httpctx.Base(clientCtx), middleware.HealthCheck, middleware.RealIP}
	if c.RedirectToHTTPS {
		mw = append(mw, middleware.RedirectToHTTPS)
	}
	if v, _ := strconv.ParseBool(os.Getenv("SG_ENABLE_HSTS")); v {
		mw = append(mw, middleware.StrictTransportSecurity)
	}
	mw = append(mw, middleware.SecureHeader)
	if v, _ := strconv.ParseBool(os.Getenv("SG_STRICT_HOSTNAME")); v {
		mw = append(mw, ensureHostnameHandler)
	}
	mw = append(mw, metricutil.HTTPMiddleware)
	if traceMiddleware := traceutil.HTTPMiddleware(); traceMiddleware != nil {
		mw = append(mw, traceMiddleware)
	}
	mw = append(mw, sourcegraphComGoGetHandler)
	if v, _ := strconv.ParseBool(os.Getenv("SG_ENABLE_GITHUB_CLONE_PROXY")); v {
		mw = append(mw, gitCloneHandler)
	}

	h := handlerutil.WithMiddleware(sm, mw...)

	if err := c.authenticateCLIContext(idKey); err != nil {
		return err
	}

	// Start background workers that receive input from main app.
	//
	// It's safe (and better) to start this before starting the
	// HTTP(S) web server to avoid a brief moment where the web server
	// is started, but the listeners haven't started yet.
	//
	//
	// Start event listeners.
	c.initializeEventListeners(client.Ctx, idKey, appURL)

	serveHTTPS := func(l net.Listener, srv *http.Server, addr string) {
		grpcSrv := server.NewServer(server.Config(serverCtxFunc))

		// Handler that sends traffic to either Web or gRPC depending
		// on content-type
		srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcSrv.ServeHTTP(w, r)
			} else {
				h.ServeHTTP(w, r)
			}
		})

		log15.Debug("HTTPS running", "on", addr)
		srv.Addr = addr
		go func() { log.Fatal(srv.Serve(l)) }()
	}

	serveHTTP := func(l net.Listener, srv *http.Server, addr string) {
		// We need to use cmux since go's built in http server won't
		// allow http/2 on non TLS connections.
		lmux := cmux.New(l)
		grpcListener := lmux.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
		anyListener := lmux.Match(cmux.Any())

		// Web
		log15.Debug("HTTP running", "on", addr)
		srv.Addr = addr
		srv.Handler = h
		go func() { log.Fatal(srv.Serve(anyListener)) }()

		// gRPC
		grpcSrv := server.NewServer(server.Config(serverCtxFunc))
		go func() { log.Fatal(grpcSrv.Serve(grpcListener)) }()

		go func() {
			if err := lmux.Serve(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
				log.Fatalf("Error serving: %s.", err)
			}
		}()
	}

	// Start HTTP server.
	if c.HTTPAddr != "" {
		l, err := net.Listen("tcp", c.HTTPAddr)
		if err != nil {
			return err
		}
		l = tcpKeepAliveListener{l.(*net.TCPListener)}
		serveHTTP(l, &http.Server{}, c.HTTPAddr)
	}

	// Start HTTPS server.
	if useTLS && c.HTTPSAddr != "" {
		l, err := net.Listen("tcp", c.HTTPSAddr)
		if err != nil {
			return err
		}
		l = tcpKeepAliveListener{l.(*net.TCPListener)}

		cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			return err
		}

		var srv http.Server
		config := srv.TLSConfig
		if config == nil {
			config = &tls.Config{}
		}

		config.Certificates = []tls.Certificate{cert}
		srv.TLSConfig = config
		l = tls.NewListener(l, srv.TLSConfig)

		serveHTTPS(l, &srv, c.HTTPSAddr)
	}

	// Connection test
	c.checkReachability()
	log15.Info(fmt.Sprintf("✱ Sourcegraph running at %s", c.AppURL))

	// Start background repo updater worker.
	repoUpdaterCtx, err := c.authenticateScopedContext(client.Ctx, idKey, []string{"internal:repoupdater"})
	if err != nil {
		return err
	}
	repoupdater.RepoUpdater.Start(repoUpdaterCtx)

	idKeyText, _ := idKey.MarshalText()
	if c.NoWorker {
		log15.Info("Skip starting worker process.")
	} else {
		go func() {
			if err := c.WorkCmd.Execute([]string{string(idKeyText)}); err != nil {
				log.Fatal("Worker exited with error:", err)
			}
		}()
	}

	// Occasionally compute instance usage stats for uplink, but don't do
	// it too often
	go statsutil.ComputeUsageStats(client.Ctx, 10*time.Minute)

	// Prepare for initial onboarding.
	if createdIDKey && !c.NoInitialOnboarding {
		if err := c.prepareInitialOnboarding(client.Ctx); err != nil {
			log15.Warn("Error preparing initial onboarding", "err", err)
		}
	}

	select {}
	return nil
}

// generateOrReadIDKey reads the server's ID key (or creates one on-demand).
func (c *ServeCmd) generateOrReadIDKey() (k *idkey.IDKey, created bool, err error) {
	if s := c.IDKeyData; s != "" {
		log15.Debug("Reading ID key from environment (or CLI flag).")
		k, err = idkey.FromString(s)
		return k, false, err
	}

	c.IDKeyFile = os.ExpandEnv(c.IDKeyFile)

	if data, err := ioutil.ReadFile(c.IDKeyFile); err == nil {
		// File exists.
		k, err = idkey.New(data)
		if err != nil {
			return nil, false, err
		}
	} else if os.IsNotExist(err) {
		log15.Debug("Generating new Sourcegraph ID key", "path", c.IDKeyFile)
		k, err = idkey.Generate()
		if err != nil {
			return nil, false, err
		}
		data, err := k.MarshalText()
		if err != nil {
			return nil, false, err
		}
		if err := os.MkdirAll(filepath.Dir(c.IDKeyFile), 0700); err != nil {
			return nil, false, err
		}
		if err := ioutil.WriteFile(c.IDKeyFile, data, 0600); err != nil {
			return nil, false, err
		}
		created = true
	}
	return
}

// authenticateCLIContext adds a "service account" access token to
// client.Ctx and to the global CLI flags (which are effectively
// inherited by subcommands run with cmdWithClientArgs). The server
// uses this to run privileged in-process workers.
//
// In general, when running the worker or other CLI commands, if the
// server requires auth, then the user needs to specify auth on the
// command line (or have it stored previously using, e.g., `src
// login`). This is because those operations don't necessarily know
// the server's secrets, so they are treated as external, untrusted
// clients. But for operations spawned IN the server process, we
// obviously know the server's secrets, so we can use them to
// authenticate the in-process worker and other CLI commands.
func (c *ServeCmd) authenticateCLIContext(k *idkey.IDKey) error {
	src := client.UpdateGlobalTokenSource{TokenSource: sharedsecret.ShortTokenSource(k, "internal:cli")}

	// Call it once to set Credentials.AccessToken immediately.
	tok, err := src.Token()
	if err != nil {
		return err
	}

	client.Ctx = sourcegraph.WithCredentials(client.Ctx, sharedsecret.DefensiveReuseTokenSource(tok, src))
	return nil
}

// authenticateScopedContext adds a token with the specified scope to the given
// context. This context can only make gRPC calls that are permitted for the given
// scope. See the accesscontrol package for information about different scopes.
func (c *ServeCmd) authenticateScopedContext(ctx context.Context, k *idkey.IDKey, scopes []string) (context.Context, error) {
	src := sharedsecret.TokenSource(k, scopes...)
	tok, err := src.Token()
	if err != nil {
		return nil, err
	}

	ctx = sourcegraph.WithCredentials(ctx, oauth2.ReuseTokenSource(tok, src))
	return ctx, nil
}

// initializeEventListeners creates special scoped contexts and passes them to
// event listeners.
func (c *ServeCmd) initializeEventListeners(parent context.Context, k *idkey.IDKey, appURL *url.URL) {
	ctx := conf.WithURL(parent, appURL)
	ctx = authpkg.WithActor(ctx, authpkg.Actor{})
	// Mask out the server's private key from the context passed to the listener
	ctx = idkey.NewContext(ctx, nil)

	for _, l := range events.GetRegisteredListeners() {
		listenerCtx, err := c.authenticateScopedContext(ctx, k, l.Scopes())
		if err != nil {
			log.Fatalf("Could not initialize listener context: %v", err)
		} else {
			l.Start(listenerCtx)
		}
	}
}

// checkReachability attempts to contact the gRPC server on both the
// internally and externally published URL. It calls log.Fatal if the
// server can't contact itself, which indicates a likely configuration
// problem (wrong gRPC URL or bind address?).
func (c *ServeCmd) checkReachability() {
	var timeout time.Duration
	if os.Getenv("CI") == "" {
		timeout = 5 * time.Second
	} else {
		timeout = 15 * time.Second // CI is slower
	}

	doCheck := func(ctx context.Context, errorIsFatal bool) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cl, err := sourcegraph.NewClientFromContext(ctx)
		if err != nil {
			if errorIsFatal {
				log.Fatalf("Fatal: could not create client: %s", err)
			} else {
				log.Printf("Warning: could not create client: %s", err)
			}
			return
		}
		if _, err := cl.Meta.Status(ctx, &pbtypes.Void{}); err != nil && grpc.Code(err) != codes.Unauthenticated {
			msg := fmt.Sprintf("Reachability check to server at %s failed (%s). Clients (including the web app) would be unable to connect to the server.", sourcegraph.GRPCEndpoint(ctx), err)
			if errorIsFatal {
				log.Fatalf(msg)
			} else {
				log.Println("Warning:", msg)
			}
		}
	}

	// Check internal gRPC endpoint.
	doCheck(client.Ctx, true)

	// Check external gRPC endpoint if it differs from the internal
	// endpoint.
	extEndpoint, err := url.Parse(c.AppURL)
	if err != nil {
		log.Fatal(err)
	}
	if extEndpoint != nil && extEndpoint.String() != sourcegraph.GRPCEndpoint(client.Ctx).String() {
		doCheck(sourcegraph.WithGRPCEndpoint(client.Ctx, extEndpoint), false)
	}
}

func gitCloneHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if ua := r.UserAgent(); !strings.HasPrefix(ua, "git/") && !strings.HasPrefix(ua, "JGit/") {
		next(w, r)
		return
	}

	// handle `git clone`
	h := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "https", Host: "github.com", Path: "/"})
	origDirector := h.Director
	h.Director = func(r *http.Request) {
		origDirector(r)
		r.Host = "github.com"
		if strings.HasPrefix(r.URL.Path, "/github.com/") {
			r.URL.Path = r.URL.Path[len("/github.com"):]
		}
	}
	h.ServeHTTP(w, r)
}

// ensureHostnameHandler ensures that the URL hostname is whatever is in SG_URL.
func ensureHostnameHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := httpctx.FromRequest(r)

	wantHost := conf.AppURL(ctx).Host
	if strings.Split(wantHost, ":")[0] == "localhost" {
		// if localhost, don't enforce redirect, so the site is easier to share with others
		next(w, r)
		return
	}

	if r.Host == wantHost || r.Host == "" || r.URL.Path == middleware.StatusEndpoint {
		next(w, r)
		return
	}

	// redirect to desired host
	newURL := *r.URL
	newURL.User = nil
	newURL.Host = wantHost
	newURL.Scheme = conf.AppURL(ctx).Scheme
	log.Printf("ensureHostnameHandler: Permanently redirecting from requested host %q to %q.", r.Host, newURL.String())
	http.Redirect(w, r, newURL.String(), http.StatusMovedPermanently)
}

// safeConfigFlags returns the commandline flag data for the `src serve` command,
// by filtering out secrets in the ServeCmdPrivate flag struct.
func (c *ServeCmd) safeConfigFlags() string {
	serveFlagData := cli.Serve.GetData()
	if len(serveFlagData) > 0 {
		// The first element is the data of the top level group under `src serve`,
		// i.e. the data of the struct ServeCmd. Since this struct contains private
		// flags (ServeCmdPrivate), we discard this struct from the returned slice,
		// and instead append a safe version of this struct.
		serveFlagData = serveFlagData[1:]
	}
	serveCmdSafe := *c
	serveCmdSafe.ServeCmdPrivate = ServeCmdPrivate{}
	configStr := fmt.Sprintf("%+v", serveCmdSafe)
	for _, data := range serveFlagData {
		if data != nil {
			configStr = fmt.Sprintf("%s %+v", configStr, data)
		}
	}
	return configStr
}

func (c *ServeCmd) runGitServer() {
	gitServers := strings.Fields(c.GitServers)
	if len(c.GitServers) != 0 {
		for _, addr := range gitServers {
			gitserver.Connect(addr)
		}
		return
	}

	stdoutReader, stdoutWriter := io.Pipe()
	go func() {
		cmd := exec.Command(srccmd.Path, "git-server", "--auto-terminate", "--repos-dir="+os.ExpandEnv(c.ReposDir))
		cmd.StdinPipe() // keep stdin from closing
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

	gitserver.Connect(addr)
}

// runSnapshotProfiler starts up the snapshotprof in a goroutine
func runSnapshotProfiler(path string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("snapshot-profiler failed to open %s: %s", path, err)
	}
	go func() {
		err := snapshotprof.Run(f, 5*time.Minute)
		log15.Error("snapshot-profiler failed writing to log file", "error", err, "path", path)
	}()
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func noiseyLogFilter(r *log15.Record) bool {
	if r.Lvl != log15.LvlDebug {
		return true
	}
	noiseyPrefixes := []string{"repoUpdater: RefreshVCS"}
	for _, prefix := range noiseyPrefixes {
		if strings.HasPrefix(r.Msg, prefix) {
			return false
		}
	}
	if !strings.HasPrefix(r.Msg, "gRPC ") || len(r.Ctx) < 2 {
		return true
	}
	rpc, ok := r.Ctx[1].(string)
	if !ok {
		return true
	}
	noisyRpc := []string{"Builds.DequeueNext", "MirrorRepos.RefreshVCS"}
	for _, n := range noisyRpc {
		if rpc == n {
			return false
		}
	}
	return true
}
