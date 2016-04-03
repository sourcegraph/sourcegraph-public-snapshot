package sgx

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
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/events"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/httpapi"
	"sourcegraph.com/sourcegraph/sourcegraph/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/snapshotprof"
	"sourcegraph.com/sourcegraph/sourcegraph/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/server"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/client"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/sgxcmd"
	"sourcegraph.com/sourcegraph/sourcegraph/ui"
	ui_router "sourcegraph.com/sourcegraph/sourcegraph/ui/router"
	"sourcegraph.com/sourcegraph/sourcegraph/util/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	httputil2 "sourcegraph.com/sourcegraph/sourcegraph/util/httputil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/util/metricutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/statsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/worker"
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

    Authentication:
          --auth.allow-anon-readers              allow unauthenticated users to perform read operations (viewing repos, etc.)
          --auth.source=                         source of authentication to use (none|local|oauth) (none)

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
	GitServers string `long:"new-git-servers" description:"addresses of the remote git servers; a local git server process is used by default" env:"SRC_NEW_GIT_SERVERS"`
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

	// Periodically forward and/or save metrics.
	metricutil.Start(clientCtx, 10*4096, 256, 5*time.Minute)
	metricutil.LogEvent(clientCtx, &sourcegraph.UserEvent{
		Type:     "notif",
		ClientID: idKey.ID,
		Service:  "serve_cmd",
		Method:   "start",
		Result:   "success",
	})
	metricutil.LogConfig(clientCtx, idKey.ID, c.safeConfigFlags())

	// Listen for events and periodically push them to analytics gateway.
	eventsutil.StartEventLogger(clientCtx, idKey.ID, 10*4096, 256, 10*time.Minute)
	eventsutil.LogStartServer()

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
	sm.Handle("/.ui/", gziphandler.GzipHandler(app.NewHandlerWithCSRFProtection(ui.NewHandler(ui_router.New(subRouter(newRouter().PathPrefix("/.ui/")))))))
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

	mw := []handlerutil.Middleware{httpctx.Base(clientCtx), healthCheckMiddleware, realIPHandler}
	if c.RedirectToHTTPS {
		mw = append(mw, redirectToHTTPSMiddleware)
	}
	if v, _ := strconv.ParseBool(os.Getenv("SG_ENABLE_HSTS")); v {
		mw = append(mw, strictTransportSecurityMiddleware)
	}
	mw = append(mw, secureHeaderMiddleware)
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

	serveHTTP := func(l net.Listener, srv *http.Server, addr string, tls bool) {
		lmux := cmux.New(l)
		grpcListener := lmux.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
		anyListener := lmux.Match(cmux.Any())

		// Web
		log15.Debug("HTTP running", "on", addr, "TLS", tls)
		srv.Addr = addr
		srv.Handler = h
		if tls {
			srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				setTLSMiddleware(w, r, h.ServeHTTP)
			})
		}
		go func() { log.Fatal(srv.Serve(anyListener)) }()

		// gRPC
		log15.Debug("gRPC API running", "on", addr, "TLS", tls)
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
		serveHTTP(l, &http.Server{}, c.HTTPAddr, false)
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

		if config.NextProtos == nil {
			config.NextProtos = []string{"http/1.1"}
		}
		config.Certificates = []tls.Certificate{cert}
		srv.TLSConfig = config
		l = tls.NewListener(l, srv.TLSConfig)

		serveHTTP(l, &srv, c.HTTPSAddr, true)
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

	// Register client.
	go func() {
		if err := c.registerClient(appURL, idKey); err != nil {
			log15.Warn("Failed to register (or check registration) with server", "error", err, "registerURL", c.RegisterURL)
		}
	}()

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
//
// This should be used for authenticating platform apps that will run in-process with
// the server, but which should have limited access to gRPC operations.
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
	ctx = authpkg.WithActor(ctx, authpkg.Actor{ClientID: k.ID})
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

// setTLSMiddleware causes downstream handlers to treat this HTTP
// request as having come via TLS. It is necessary because connection
// muxing (which enables a single port to serve both Web and gRPC)
// does not set the http.Request TLS field (since TLS occurs before
// muxing).
func setTLSMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Header.Get("x-forwarded-proto") == "" {
		r.Header.Set("x-forwarded-proto", "https")
	}
	next(w, r)
}

func redirectToHTTPSMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	isHTTPS := r.TLS != nil || r.Header.Get("x-forwarded-proto") == "https"
	if !isHTTPS {
		url := *r.URL
		url.Scheme = "https"
		url.Host = r.Host
		http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
		return
	}
	next(w, r)
}

func strictTransportSecurityMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Omit subdomains for blogs like gophercon.sourcegraph.com, which need to be HTTP.
	w.Header().Set("strict-transport-security", "max-age=8640000")
	next(w, r)
}

func secureHeaderMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("x-content-type-options", "nosniff")
	w.Header().Set("x-xss-protection", "1; mode=block")
	w.Header().Set("x-frame-options", "DENY")
	next(w, r)
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

	if r.Host == wantHost || r.Host == "" || r.URL.Path == statusEndpoint {
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

// realIPHandler sets req.RemoteAddr from the X-Real-Ip header if it exists.
func realIPHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if s := r.Header.Get("X-Real-Ip"); s != "" && httputil2.StripPort(r.RemoteAddr) == "127.0.0.1" {
		r.RemoteAddr = s
	}
	next(w, r)
}

// webpackDevServerHandler sets a CORS header if you are running with Webpack in local dev.

// registerClient registers this Sourcegraph server as a client
// of another server.
func (c *ServeCmd) registerClient(appURL *url.URL, idKey *idkey.IDKey) error {
	if c.RegisterURL == "" {
		return nil
	}

	// Short-circuit if RegisterURL is set to this server's own app URL.
	if c.RegisterURL == c.AppURL {
		log15.Warn("RegisterURL is set to this server's own app URL, skipping client registration.")
		return nil
	}

	registerURL, err := url.Parse(c.RegisterURL)
	if err != nil {
		return err
	}

	shortClientID := idKey.ID
	if len(shortClientID) > 5 {
		shortClientID = shortClientID[:5]
	}

	jwks, err := idKey.MarshalJWKSPublicKey()
	if err != nil {
		return err
	}

	rctx := sourcegraph.WithGRPCEndpoint(context.Background(), registerURL)

	cl, err := sourcegraph.NewClientFromContext(rctx)
	if err != nil {
		return err
	}

	regClient, err := cl.RegisteredClients.Create(rctx, &sourcegraph.RegisteredClient{
		ID:         idKey.ID,
		ClientName: fmt.Sprintf("Client #%s", shortClientID),
		ClientURI:  appURL.String(),
		JWKS:       string(jwks),
	})
	if grpc.Code(err) == codes.AlreadyExists {
		log15.Debug("Client is already registered", "registerURL", registerURL, "client", idKey.ID)
		return nil
	} else if err != nil {
		return fmt.Errorf("registering client with %s: %s", registerURL, err)
	}
	eventsutil.LogRegisterServer(regClient.ClientName)
	log15.Debug("Registered as client", "registerURL", registerURL, "client", idKey.ID)
	return nil
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
		cmd := exec.Command(sgxcmd.Path, "git-server", "--auto-terminate", "--repos-dir="+os.ExpandEnv(c.ReposDir))
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
