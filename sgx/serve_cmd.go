package sgx

import (
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/resonancelabs/go-pub/instrument"
	tg_client "github.com/resonancelabs/go-pub/instrument/client"
	"github.com/soheilhy/cmux"
	"github.com/sourcegraph/mux"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	app_router "src.sourcegraph.com/sourcegraph/app/router"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/auth/ldap"
	"src.sourcegraph.com/sourcegraph/auth/sharedsecret"
	"src.sourcegraph.com/sourcegraph/client/pkg/oauth2client"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/gitserver/sshgit"
	"src.sourcegraph.com/sourcegraph/httpapi"
	"src.sourcegraph.com/sourcegraph/httpapi/router"
	"src.sourcegraph.com/sourcegraph/server"
	localcli "src.sourcegraph.com/sourcegraph/server/local/cli"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/ui"
	ui_router "src.sourcegraph.com/sourcegraph/ui/router"
	"src.sourcegraph.com/sourcegraph/usercontent"
	"src.sourcegraph.com/sourcegraph/util/cacheutil"
	"src.sourcegraph.com/sourcegraph/util/expvarutil"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	httputil2 "src.sourcegraph.com/sourcegraph/util/httputil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
	"src.sourcegraph.com/sourcegraph/util/metricutil"
	"src.sourcegraph.com/sourcegraph/util/statsutil"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
)

// Stripped down help message presented to most users (full help message can be
// gotten with -a or --help-all).
var shortHelpMessage = `Usage:
  src serve [serve-OPTIONS]

Starts an HTTP server serving the app and API.

[serve command options]
          --http-addr=                           HTTP listen address for app, REST API, and gRPC API (:3080)
          --https-addr=                          HTTPS (TLS) listen address for app, REST API, and gRPC API (:3443)
          --ssh-addr=                            SSH address to listen on, if not blank (:3022)
          --prof-http=BIND-ADDR                  net/http/pprof http bind address (:6060)
          --app-url=                             publicly accessible URL to web app (e.g., what you type into your browser) (http://<http-addr>)
          --reload                               reload templates, blog posts, etc. on each request (dev mode)
          --no-worker                            do not start background worker
          --test-ui                              starts the UI test server which causes all UI endpoints to return mock data
          --tls-cert=                            certificate file (for TLS)
          --tls-key=                             key file (for TLS)
      -i, --id-key=                              identity key file ($SGPATH/id.pem) [$SRC_ID_KEY_FILE]
          --id-key-data=                         identity key file data (overrides -i/--id-key) [$SRC_ID_KEY_DATA]
          --dequeue-msec=                        if no builds are dequeued, sleep up to this many msec before trying again (1000)
      -n, --num-workers=                         number of parallel workers (1) [$SG_NUM_WORKERS]
          --build-root=                          root of dir tree in which to perform builds ($SGPATH/builds)
          --clean                                remove temp dirs and build data when the worker starts and after builds complete

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
// instance. These fields will not be sent to the federation root server
// for diagnostic purposes.
type ServeCmdPrivate struct {
	CertFile string `long:"tls-cert" description:"certificate file (for TLS)"`
	KeyFile  string `long:"tls-key" description:"key file (for TLS)"`

	IDKeyFile string `short:"i" long:"id-key" description:"identity key file" default:"$SGPATH/id.pem" env:"SRC_ID_KEY_FILE"`
	IDKeyData string `long:"id-key-data" description:"identity key file data (overrides -i/--id-key)" env:"SRC_ID_KEY_DATA"`
}

var serveCmdInst ServeCmd

type ServeCmd struct {
	HTTPAddr  string `long:"http-addr" default:":3080" description:"HTTP listen address for app, REST API, and gRPC API"`
	HTTPSAddr string `long:"https-addr" default:":3443" description:"HTTPS (TLS) listen address for app, REST API, and gRPC API"`
	SSHAddr   string `long:"ssh-addr" default:":3022" description:"SSH address to listen on, if not blank"`

	ProfBindAddr string `long:"prof-http" default:":6060" description:"net/http/pprof http bind address" value-name:"BIND-ADDR"`

	AppURL string `long:"app-url" default:"http://<http-addr>" description:"publicly accessible URL to web app (e.g., what you type into your browser)"`

	RedirectToHTTPS bool `long:"app.redirect-to-https" description:"redirect HTTP requests to the equivalent HTTPS URL" env:"SG_FORCE_HTTPS"`

	NoWorker          bool          `long:"no-worker" description:"do not start background worker"`
	TestUI            bool          `long:"test-ui" description:"starts the UI test server which causes all UI endpoints to return mock data"`
	GraphUplinkPeriod time.Duration `long:"graphuplink" default:"10m" description:"how often to communicate back to the mothership; if 0, then no periodic communication occurs"`

	// Flags containing sensitive information must be added to this struct.
	ServeCmdPrivate

	Prefetch bool `long:"prefetch" description:"prefetch directory children"`

	WorkCmd

	GraphStoreOpts `group:"Graph data storage (defs, refs, etc.)" namespace:"graphstore"`
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
	if Endpoint.URL == "" {
		Endpoint.URL = appURL.String()

		// Reset cliCtx to use new endpoint.
		cliCtx = WithClientContext(context.Background())
	}

	return appURL, nil
}

func (c *ServeCmd) Execute(args []string) error {
	logHandler := log15.StderrHandler
	if globalOpt.VerbosePkg != "" {
		logHandler = log15.MatchFilterHandler("pkg", globalOpt.VerbosePkg, log15.StderrHandler)
	}

	// Filter log output by level.
	lvl, err := log15.LvlFromString(globalOpt.LogLevel)
	if err != nil {
		return err
	}
	log15.Root().SetHandler(log15.LvlFilterHandler(lvl, logHandler))

	if len(os.Getenv("SG_TRACEGUIDE_ACCESS_TOKEN")) > 0 {
		options := &tg_client.Options{
			AccessToken: os.Getenv("SG_TRACEGUIDE_ACCESS_TOKEN"),
		}
		if len(os.Getenv("SG_TRACEGUIDE_SERVICE_HOST")) > 0 {
			options.ServiceHost = os.Getenv("SG_TRACEGUIDE_SERVICE_HOST")
		}
		instrument.SetDefaultRuntime(tg_client.NewRuntime(options))
		instrument.Log("Initialized Traceguide runtime")
	}

	cacheutil.Precache = c.Prefetch
	log15.Debug("Cache prefetching", "enabled", cacheutil.Precache)

	// Clear auth specified on the CLI. If we didn't do this, then the
	// app, git, and HTTP API would all inherit the process's owner's
	// current auth. This is undesirable, unexpected, and could lead
	// to unintentionally leaking private info.
	Credentials.AccessToken = ""
	Credentials.AuthFile = ""

	c.GraphStoreOpts.expandEnv()
	log15.Debug("GraphStore", "at", c.GraphStoreOpts.Root)

	for _, f := range cli.ServeInit {
		f()
	}

	if c.ProfBindAddr != "" {
		// Starts a pprof server by default, but this is OK, because only
		// whitelisted ports on the web server machines should be publicly
		// accessible anyway.
		go func() {
			pp := http.NewServeMux()
			pp.Handle("/debug/vars", http.HandlerFunc(expvarutil.ExpvarHandler))
			pp.Handle("/debug/gc", http.HandlerFunc(expvarutil.GCHandler))
			pp.Handle("/debug/freeosmemory", http.HandlerFunc(expvarutil.FreeOSMemoryHandler))
			pp.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
			pp.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
			pp.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
			pp.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
			pp.Handle("/metrics", prometheus.Handler())
			log.Println("warning: could not start pprof HTTP server:", http.ListenAndServe(c.ProfBindAddr, pp))
		}()
		log15.Debug("Profiler available", "on", fmt.Sprintf("%s/debug/pprof", c.ProfBindAddr))
	}

	var (
		sharedCtxFuncs []func(context.Context) context.Context
		serverCtxFuncs []func(context.Context) context.Context
		clientCtxFuncs []func(context.Context) context.Context = ClientContextFuncs
	)

	// Server identity keypair
	idKey, err := c.generateOrReadIDKey()
	if err != nil {
		return err
	}
	log15.Debug("Sourcegraph server", "ID", idKey.ID)
	// Uncomment to add ID key prefix to log messages.
	// log.SetPrefix(bold(idKey.ID[:4] + ": "))
	var idKeyToken oauth2.TokenSource
	var rootPubKey *idkey.PubKey
	if !fed.Config.IsRoot {
		tokenURL := fed.Config.RootURL().ResolveReference(app_router.Rel.URLTo(app_router.OAuth2ServerToken))
		idKeyToken = idKey.TokenSource(context.Background(), tokenURL.String())

		// Fetch the root server's public key
		rootPubKey = c.fetchRootPubKey()
	}

	sharedCtxFuncs = append(sharedCtxFuncs, func(ctx context.Context) context.Context {
		if !fed.Config.IsRoot {
			ctx = sourcegraph.WithCredentials(ctx, idKeyToken)
			ctx = idkey.WithRootPubKey(ctx, rootPubKey)
		}
		ctx = idkey.NewContext(ctx, idKey)
		return ctx
	})
	clientCtxFuncs = append(clientCtxFuncs, func(ctx context.Context) context.Context {
		return oauth2client.WithClientID(ctx, idKey.ID)
	})
	sharedSecretToken := oauth2.ReuseTokenSource(nil, sharedsecret.TokenSource(idKey))
	clientCtxFuncs = append(clientCtxFuncs, func(ctx context.Context) context.Context {
		return sourcegraph.WithCredentials(ctx, sharedSecretToken)
	})

	// graphstore
	serverCtxFuncs = append(serverCtxFuncs, c.GraphStoreOpts.context)

	// User Content.
	if !appconf.Flags.DisableUserContent {
		var err error
		usercontent.Store, err = usercontent.LocalStore()
		if err != nil {
			return err
		}
	}

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
		ctx = conf.WithAppURL(ctx, appURL)
		return ctx
	}

	clientCtx := func(ctx context.Context) context.Context {
		ctx = sharedCtxFunc(ctx)
		for _, f := range clientCtxFuncs {
			ctx = f(ctx)
		}
		ctx = WithClientContext(ctx)
		return ctx
	}(context.Background())

	serverCtxFunc := func(ctx context.Context) context.Context {
		ctx = sharedCtxFunc(ctx)
		for _, f := range serverCtxFuncs {
			ctx = f(ctx)
		}

		ctx = Endpoint.NewContext(ctx)

		for _, f := range ServerContextFuncs {
			var err error
			ctx, err = f(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}

		return ctx
	}

	if fed.Config.IsRoot {
		// Listen for events and flush them to elasticsearch
		metricutil.StartEventForwarder(clientCtx)
		metricutil.StartEventLogger(clientCtx, 10*4096, 1024, 5*time.Minute)
	} else if c.GraphUplinkPeriod != 0 {
		// Listen for events and periodically push them upstream
		metricutil.StartEventLogger(clientCtx, 10*4096, 256, 10*time.Minute)
	}

	metricutil.LogEvent(clientCtx, &sourcegraph.UserEvent{
		Type:     "notif",
		ClientID: idKey.ID,
		Service:  "serve_cmd",
		Method:   "start",
		Result:   "success",
	})
	metricutil.LogConfig(clientCtx, idKey.ID, c.safeConfigFlags())

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
	sm.Handle("/.api/", httpapi.NewHandler(router.New(subRouter(newRouter().PathPrefix("/.api/")))))
	sm.Handle("/.ui/", ui.NewHandler(ui_router.New(subRouter(newRouter().PathPrefix("/.ui/")), c.TestUI), c.TestUI))
	sm.Handle("/", app.NewHandlerWithCSRFProtection(app_router.New(newRouter())))

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
	if app.UseWebpackDevServer {
		mw = append(mw, webpackDevServerHandler)
	}
	// TODO: if we keep this old behavior of `go get` cloning GH repos, do it
	// under a better-named environment variable.
	if v, _ := strconv.ParseBool(os.Getenv("SG_ENABLE_GO_GET")); v {
		mw = append(mw, goGetGitHubHandler)
	} else {
		mw = append(mw, goGetHandler)
	}
	if v, _ := strconv.ParseBool(os.Getenv("SG_ENABLE_GITHUB_CLONE_PROXY")); v {
		mw = append(mw, gitCloneHandler)
	}

	h := handlerutil.WithMiddleware(sm, mw...)

	serveHTTP := func(l net.Listener, srv http.Server, addr string, tls bool) {
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
		serveHTTP(l, http.Server{}, c.HTTPAddr, false)
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

		serveHTTP(l, srv, c.HTTPSAddr, true)
	}

	log15.Info(fmt.Sprintf("✱ Sourcegraph running at %s", c.AppURL))

	if err := c.authenticateCLIContext(idKey); err != nil {
		return err
	}

	cacheutil.HTTPAddr = c.AppURL // TODO: HACK

	if !c.NoWorker {
		go func() {
			if err := c.WorkCmd.Execute(nil); err != nil {
				log.Fatal("Worker exited with error:", err)
			}
		}()
	}

	if authutil.ActiveFlags.IsLDAP() {
		if err := ldap.VerifyConfig(); err != nil {
			log.Fatalf("Could not connect to LDAP server: %v", err)
		} else {
			log15.Info("Connection to LDAP server successful")
		}
	}

	// Start SSH git server.
	if c.SSHAddr != "" {
		privateSigner, err := ssh.NewSignerFromKey(idKey.Private())
		if err != nil {
			return err
		}
		// create a context with regular (non self-signed) tokens that
		// are valid on the federation root server.
		ctx, err := c.authenticateScopedContext(cliCtx, idKey, []string{"internal:sshgit"})
		if err != nil {
			return err
		}
		err = (&sshgit.Server{}).ListenAndStart(ctx, c.SSHAddr, privateSigner, idKey.ID)
		if err != nil {
			return err
		}
	}

	// Start background repo updater worker.
	app.RepoUpdater.Start(cliCtx)

	// Refresh commit list periodically
	go c.repoStatusCommitLogCacheRefresher()

	// Start event listeners
	c.initializeEventListeners(cliCtx, idKey, appURL)

	// Occasionally compute instance usage stats for uplink, but don't do
	// it too often
	statsInterval := c.GraphUplinkPeriod
	if statsInterval < 10*time.Minute {
		statsInterval = 10 * time.Minute
	}

	go statsutil.ComputeUsageStats(cliCtx, statsInterval)

	// Occasionally send metrics and usage stats upstream via GraphUplink
	go c.graphUplink(clientCtx)

	// Connection test
	c.checkReachability()

	// Wait for signal to exit.
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	return nil
}

// generateOrReadIDKey reads the server's ID key (or creates one on-demand).
func (c *ServeCmd) generateOrReadIDKey() (*idkey.IDKey, error) {
	if s := c.IDKeyData; s != "" {
		if globalOpt.Verbose {
			log.Println("Reading ID key from environment (or CLI flag).")
		}

		return idkey.FromString(s)
	}

	c.IDKeyFile = os.ExpandEnv(c.IDKeyFile)

	var k *idkey.IDKey
	if data, err := ioutil.ReadFile(c.IDKeyFile); err == nil {
		// File exists.
		k, err = idkey.New(data)
		if err != nil {
			return nil, err
		}
	} else if os.IsNotExist(err) {
		log.Printf("Generating new Sourcegraph ID key at %s...", c.IDKeyFile)
		k, err = idkey.Generate()
		if err != nil {
			return nil, err
		}
		data, err := k.MarshalText()
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Dir(c.IDKeyFile), 0700); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(c.IDKeyFile, data, 0600); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	return k, nil
}

// authenticateCLIContext adds a "service account" access token to
// cliCtx and to the global CLI flags (which are effectively inherited
// by subcommands run with cmdWithClientArgs). The server uses this to
// run privileged in-process workers.
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
	src := updateGlobalTokenSource{sharedsecret.ShortTokenSource(k, "internal:cli")}

	// Call it once to set Credentials.AccessToken immediately.
	tok, err := src.Token()
	if err != nil {
		return err
	}

	cliCtx = sourcegraph.WithCredentials(cliCtx, oauth2.ReuseTokenSource(tok, src))
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

// updateGlobalTokenSource updates Credentials.AccessToken with the
// newest access token each time it is refreshed.
//
// TODO(sqs): synchronize access to Credentials.AccessToken.
type updateGlobalTokenSource struct{ oauth2.TokenSource }

func (ts updateGlobalTokenSource) Token() (*oauth2.Token, error) {
	tok, err := ts.TokenSource.Token()
	if tok != nil && tok.AccessToken != Credentials.AccessToken {
		Credentials.AccessToken = tok.AccessToken
	}
	return tok, err
}

// initializeEventListeners creates special scoped contexts and passes them to
// event listeners.
func (c *ServeCmd) initializeEventListeners(parent context.Context, k *idkey.IDKey, appURL *url.URL) {
	ctx := conf.WithAppURL(parent, appURL)
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
		timeout = time.Second
	} else {
		timeout = 15 * time.Second // CI is slower
	}

	doCheck := func(ctx context.Context, errorIsFatal bool) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cl := sourcegraph.NewClientFromContext(ctx)
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
	doCheck(cliCtx, true)

	// Check external gRPC endpoint if it differs from the internal
	// endpoint.
	extEndpoint, err := url.Parse(c.AppURL)
	if err != nil {
		log.Fatal(err)
	}
	if extEndpoint != nil && extEndpoint.String() != sourcegraph.GRPCEndpoint(cliCtx).String() {
		doCheck(sourcegraph.WithGRPCEndpoint(cliCtx, extEndpoint), false)
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

func goGetHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.URL.Query().Get("go-get") != "1" {
		next(w, r)
		return
	}

	// Escape the repo URI so we don't open ourselves up to XSS. We just assume
	// that the user requested a valid repo URI, go get handles the rest for us if
	// they didn't.
	if !strings.HasPrefix(r.URL.Path, "/") {
		r.URL.Path = "/" + r.URL.Path
	}
	repoURI := template.HTMLEscapeString(r.URL.Path)

	// Build git repository clone URL.
	appURL := conf.AppURL(httpctx.FromRequest(r))
	gitRepo := &url.URL{
		Scheme: appURL.Scheme,
		Host:   appURL.Host,
		Path:   repoURI,
	}

	// Build the Go package path (e.g. "src.example.com/my/go/pkg").
	pkgPath := strings.TrimPrefix(appURL.Host, "www.") + repoURI // for when AppURL has a leading www

	fmt.Fprintf(w, `<html><head><meta name="go-import" content="%s git %s"></head><body>go get %s</body></html>`, pkgPath, gitRepo, pkgPath)
	log.Println("go get", pkgPath)
}

func goGetGitHubHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.URL.Query().Get("go-get") != "1" {
		next(w, r)
		return
	}
	// handle `go get`
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "import paths must have at least 2 URL path components", http.StatusNotFound)
		return
	}
	repo := template.HTMLEscapeString(strings.Join(parts[:2], "/"))
	// Determine the host (without protocol) prefix.
	host := conf.AppURL(httpctx.FromRequest(r)).Host
	host = strings.TrimPrefix(host, "www.") // for when AppURL has a leading www
	fmt.Fprintf(w, `<html><head><meta name="go-import" content="%s/%s git https://github.com/%s"></head><body>go-get</body></html>`, host, repo, repo)
	log.Println("go-get", strings.Join(parts, "/"))
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
func webpackDevServerHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if w.Header().Get("access-control-allow-origin") == "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	next(w, r)
}

// repoStatusCommitLogCacheRefresher periodically refreshes the commit log cache
// for each repository's default branch
func (c *ServeCmd) repoStatusCommitLogCacheRefresher() {
	if localcli.Flags.CommitLogCachePeriod == 0 {
		return
	}
	log15.Debug("commit-log-cache", "refresh-period", localcli.Flags.CommitLogCachePeriod)

	cl := Client()
Outer:
	for {
		var allRepos []*sourcegraph.Repo
		for page := int32(1); ; page++ {
			repos, err := cl.Repos.List(cliCtx, &sourcegraph.RepoListOptions{
				ListOptions: sourcegraph.ListOptions{Page: page},
			})
			if err != nil {
				log.Printf("RepoStatusCommits: Repos.List error: %s", err)
				time.Sleep(localcli.Flags.CommitLogCachePeriod)
				continue Outer
			}
			if len(repos.Repos) == 0 {
				break
			}
			allRepos = append(allRepos, repos.Repos...)
		}

		for _, repo := range allRepos {
			_, err := cl.Repos.ListCommits(cliCtx, &sourcegraph.ReposListCommitsOp{
				Repo: sourcegraph.RepoSpec{URI: repo.URI},
				Opt: &sourcegraph.RepoListCommitsOptions{
					Head:         repo.DefaultBranch,
					RefreshCache: true,
					ListOptions: sourcegraph.ListOptions{
						// Note: this number should be greater than or equal to the number of
						// commits that builds.getRepoBuildInfoInexact (in svc/local/builds_repo.go)
						// looks back for a successful build.
						PerPage: 250,
					},
				},
			})
			log15.Debug("commit-log-cache", "refreshed-repo", repo.URI)
			if err != nil {
				log.Printf("RepoStatusCommits: Repos.ListCommits error: %s", err)
				continue
			}

			// pre-cache root dir
			cacheutil.PrecacheRoot(repo.URI)
		}
		time.Sleep(localcli.Flags.CommitLogCachePeriod)
	}
}

func (c *ServeCmd) graphUplink(ctx context.Context) {
	if c.GraphUplinkPeriod == 0 || fed.Config.IsRoot {
		return
	}

	rctx := fed.Config.NewRemoteContext(ctx)

	for {
		time.Sleep(c.GraphUplinkPeriod)
		cl := sourcegraph.NewClientFromContext(rctx)
		buf := &bytes.Buffer{}
		mfs := metricutil.SnapshotMetricFamilies()
		mfs.Marshal(buf)

		log15.Debug("GraphUplink sending metrics snapshot", "mothership", fed.Config.RootURL(), "numMetrics", len(mfs))
		snapshot := sourcegraph.MetricsSnapshot{
			Type:          sourcegraph.TelemetryType_PrometheusDelimited0dot0dot4,
			TelemetryData: buf.Bytes(),
		}
		_, err := cl.GraphUplink.Push(rctx, &snapshot)
		if err != nil {
			log15.Error("GraphUplink push failed", "error", err)
		}
	}
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

// fetchRootPubKey will fetch the root server's published public key
// and return the public key and fingerprint.
func (c *ServeCmd) fetchRootPubKey() *idkey.PubKey {
	if fed.Config.IsRoot {
		return nil
	}

	ctx := fed.Config.NewRemoteContext(context.Background())
	rootKey, err := sourcegraph.NewClientFromContext(ctx).Meta.PubKey(ctx, &pbtypes.Void{})
	if err != nil {
		log15.Error("fetchRootPubKey could not fetch public key", "error", err)
		return nil
	}

	rootPubKey, err := x509.ParsePKIXPublicKey([]byte(rootKey.Key))
	if err != nil {
		log15.Error("fetchRootPubKey could not parse the public key", "error", err)
		return nil
	}

	rootPubKeyRSA, ok := rootPubKey.(crypto.PublicKey)
	if !ok {
		log15.Error("fetchRootPubKey got unknown public key format")
		return nil
	}

	rootID, err := idkey.Fingerprint(rootPubKeyRSA)
	if err != nil {
		log15.Error("fetchRootPubKey could not compute public key fingerprint", err)
		return nil
	}

	return &idkey.PubKey{Key: rootPubKeyRSA, ID: rootID}
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
