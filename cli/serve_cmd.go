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
	"net/url"
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/keegancsmith/tmpfriend"
	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	app_router "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/loghandlers"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/middleware"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/srccmd"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/serverctx"
	"sourcegraph.com/sourcegraph/sourcegraph/services/events"
	"sourcegraph.com/sourcegraph/sourcegraph/services/httpapi"
	"sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/services/repoupdater"
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

	IDKeyData string `long:"id-key-data" description:"identity key file data (overrides -i/--id-key)" env:"SRC_ID_KEY_DATA"`
}

var serveCmdInst ServeCmd

type ServeCmd struct {
	HTTPAddr  string `long:"http-addr" default:":3080" description:"HTTP listen address for app, REST API, and gRPC API" env:"SRC_HTTP_ADDR"`
	HTTPSAddr string `long:"https-addr" default:":3443" description:"HTTPS (TLS) listen address for app, REST API, and gRPC API" env:"SRC_HTTPS_ADDR"`

	ProfBindAddr string `long:"prof-http" default:":6060" description:"net/http/pprof http bind address" value-name:"BIND-ADDR" env:"SRC_PROF_HTTP"`

	AppURL string `long:"app-url" default:"http://<http-addr>" description:"publicly accessible URL to web app (e.g., what you type into your browser)" env:"SRC_APP_URL"`

	NoWorker bool `long:"no-worker" description:"deprecated"`

	// Flags containing sensitive information must be added to this struct.
	ServeCmdPrivate

	GraphStoreOpts `group:"Graph data storage (defs, refs, etc.)" namespace:"graphstore"`

	ReposDir   string `long:"fs.repos-dir" description:"root dir containing repos" default:"$SGPATH/repos" env:"SRC_REPOS_DIR"`
	Gitservers string `long:"git-servers" description:"addresses of the remote gitservers; a local gitserver process is used by default" env:"SRC_GIT_SERVERS"`
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
	if endpoint.URL == "" {
		endpoint.URL = appURL.String()
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
		logHandler = log15.FilterHandler(loghandlers.NotNoisey, logHandler)
	}

	// Filter trace logs
	logHandler = log15.FilterHandler(loghandlers.Trace(globalOpt.Trace, globalOpt.TraceThreshold), logHandler)

	// Filter log output by level.
	lvl, err := log15.LvlFromString(globalOpt.LogLevel)
	if err != nil {
		return err
	}
	log15.Root().SetHandler(log15.LvlFilterHandler(lvl, logHandler))

	if t := os.Getenv("LIGHTSTEP_ACCESS_TOKEN"); t != "" {
		opentracing.InitGlobalTracer(lightstep.NewTracer(lightstep.Options{
			AccessToken: t,
		}))
	}

	// Don't proceed if system requirements are missing, to avoid
	// presenting users with a half-working experience.
	if err := checkSysReqs(context.Background(), os.Stderr); err != nil {
		return err
	}

	c.GraphStoreOpts.expandEnv()
	log15.Debug("GraphStore", "at", c.GraphStoreOpts.Root)

	for _, f := range cli.ServeInit {
		f()
	}

	if c.ProfBindAddr != "" {
		go debugserver.Start(c.ProfBindAddr)
		log15.Debug("Profiler available", "on", fmt.Sprintf("%s/pprof", c.ProfBindAddr))
	}

	app.Init()

	conf.AppURL, err = c.configureAppURL()
	if err != nil {
		return err
	}

	c.GraphStoreOpts.apply()

	// Server identity keypair
	if s := c.IDKeyData; s != "" {
		auth.ActiveIDKey, err = auth.FromString(s)
		if err != nil {
			return err
		}
	} else {
		log15.Warn("Using default ID key.")
	}

	// Shared context setup between client and server.
	sharedCtxFunc := func(ctx context.Context) context.Context {
		return ctx
	}
	clientCtxFunc := func(ctx context.Context) context.Context {
		ctx = sharedCtxFunc(ctx)
		for _, f := range cli.ClientContext {
			ctx = f(ctx)
		}
		return ctx
	}

	clientCtx := clientCtxFunc(context.Background())

	c.runGitserver()

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
	sm.Handle("/", gziphandler.GzipHandler(handlerutil.NewHandlerWithCSRFProtection(app.NewHandler(app_router.New(newRouter())))))
	sm.Handle(assets.URLPathPrefix+"/", http.StripPrefix(assets.URLPathPrefix, assets.NewHandler(newRouter())))

	if (c.CertFile != "" || c.KeyFile != "") && c.HTTPSAddr == "" {
		return errors.New("HTTPS listen address (--https-addr) must be specified if TLS cert and key are set")
	}
	useTLS := c.CertFile != "" || c.KeyFile != ""

	if useTLS && conf.AppURL.Scheme == "http" {
		log15.Warn("TLS is enabled but app url scheme is http", "appURL", conf.AppURL)
	}

	if !useTLS && conf.AppURL.Scheme == "https" {
		log15.Warn("TLS is disabled but app url scheme is https", "appURL", conf.AppURL)
	}

	mw := []handlerutil.Middleware{httpctx.Base(clientCtx), middleware.HealthCheck, middleware.RealIP, middleware.NoCacheByDefault}
	if v, _ := strconv.ParseBool(os.Getenv("SG_ENABLE_HSTS")); v {
		mw = append(mw, middleware.StrictTransportSecurity)
	}
	mw = append(mw, middleware.SecureHeader)
	mw = append(mw, httptrace.Middleware)
	mw = append(mw, middleware.BlackHole)
	mw = append(mw, middleware.SourcegraphComGoGetHandler)

	h := handlerutil.WithMiddleware(sm, mw...)

	// Start background workers that receive input from main app.
	//
	// It's safe (and better) to start this before starting the
	// HTTP(S) web server to avoid a brief moment where the web server
	// is started, but the listeners haven't started yet.
	//
	//
	// Start event listeners.
	c.initializeEventListeners(clientCtx)

	serveHTTPS := func(l net.Listener, srv *http.Server, addr string) {
		log15.Debug("HTTPS running", "on", addr)
		srv.Handler = h
		srv.Addr = addr
		go func() { log.Fatal(srv.Serve(l)) }()
	}

	serveHTTP := func(l net.Listener, srv *http.Server, addr string) {
		log15.Debug("HTTP running", "on", addr)
		srv.Addr = addr
		srv.Handler = h
		go func() { log.Fatal(srv.Serve(l)) }()
	}

	// Start HTTP server.
	if c.HTTPAddr != "" {
		l, err := net.Listen("tcp", c.HTTPAddr)
		if err != nil {
			return err
		}
		l = tcpKeepAliveListener{l.(*net.TCPListener)}
		serveHTTP(l, &http.Server{
			ReadTimeout:  75 * time.Second,
			WriteTimeout: 60 * time.Second,
		}, c.HTTPAddr)
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

		srv := &http.Server{
			ReadTimeout:  75 * time.Second,
			WriteTimeout: 60 * time.Second,
		}
		config := srv.TLSConfig
		if config == nil {
			config = &tls.Config{
				NextProtos: []string{"h2"},
			}
		}

		config.Certificates = []tls.Certificate{cert}
		srv.TLSConfig = config
		l = tls.NewListener(l, srv.TLSConfig)

		serveHTTPS(l, srv, c.HTTPSAddr)
	}

	// Connection test
	log15.Info(fmt.Sprintf("✱ Sourcegraph running at %s", c.AppURL))

	// Start background repo updater worker.
	repoUpdaterCtx, err := authenticateScopedContext(clientCtx, []string{"internal:repoupdater"})
	if err != nil {
		return err
	}
	repoupdater.RepoUpdater.Start(repoUpdaterCtx)

	// HACK(sjl) The Golang garbage collector is a bit conservative with memory that should
	// be returned to host.  On larger nodes, this can get to be 10 gigabytes or
	// more held idle.  By periodically advising the GC to return idle ram, the
	// amount of unused memory is kept to a minimum and processes should no longer
	// get OOMkilled due to spikes.  To fix this, we create a goroutine that
	// calls freeOSMemory() on a loop for the life of the application.
	// We should remove this when Golang's GC gets less agressive about holding on to allocated RAM.
	go func() {
		for {
			debug.FreeOSMemory()
			time.Sleep(60 * time.Second)
		}
	}()

	// HACK(keegancsmith) async is the only user of this at the moment,
	// but other background workers that need access to the store will
	// likely pop up in the future. We need to make this less hacky
	internalServerCtx := func(name string) (context.Context, error) {
		ctx := clientCtx
		for _, f := range serverctx.Funcs {
			ctx, err = f(ctx)
			if err != nil {
				return nil, err
			}
		}
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
	tok, err := auth.NewAccessToken(a, nil, 0)
	if err != nil {
		return nil, err
	}
	return auth.WithActor(sourcegraph.WithAccessToken(ctx, tok), a), nil
}

// initializeEventListeners creates special scoped contexts and passes them to
// event listeners.
func (c *ServeCmd) initializeEventListeners(parent context.Context) {
	for _, l := range events.GetRegisteredListeners() {
		listenerCtx, err := authenticateScopedContext(auth.WithActor(parent, &auth.Actor{}), l.Scopes())
		if err != nil {
			log.Fatalf("Could not initialize listener context: %v", err)
		} else {
			l.Start(listenerCtx)
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

// runGitserver either connects to gitservers specified in c.Gitservers, if any.
// Otherwise it starts a single local gitserver and connects to it.
func (c *ServeCmd) runGitserver() {
	gitservers := strings.Fields(c.Gitservers)
	if len(gitservers) != 0 {
		for _, addr := range gitservers {
			gitserver.DefaultClient.Connect(addr)
		}
		return
	}

	stdoutReader, stdoutWriter := io.Pipe()
	go func() {
		cmd := exec.Command(srccmd.Path, "git-server", "--auto-terminate", "--repos-dir="+os.ExpandEnv(c.ReposDir))
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
