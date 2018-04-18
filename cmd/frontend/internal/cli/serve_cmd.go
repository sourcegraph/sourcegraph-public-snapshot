package cli

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"context"

	"github.com/NYTimes/gziphandler"
	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/keegancsmith/tmpfriend"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/pkg/updatecheck"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/bg"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli/loghandlers"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli/middleware"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/useractivity"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/processrestart"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/sysreq"
	tracepkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/trace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"
)

var (
	trace          = env.Get("SRC_LOG_TRACE", "HTTP", "space separated list of trace logs to show. Options: all, HTTP, build, github")
	traceThreshold = env.Get("SRC_LOG_TRACE_THRESHOLD", "", "show traces that take longer than this")

	printLogo, _ = strconv.ParseBool(env.Get("LOGO", "false", "print Sourcegraph logo upon startup"))

	httpAddr         = env.Get("SRC_HTTP_ADDR", ":3080", "HTTP listen address for app and HTTP API")
	httpsAddr        = env.Get("SRC_HTTPS_ADDR", ":3443", "HTTPS (TLS) listen address for app and HTTP API. Only used if manual tls cert and key are specified.")
	httpAddrInternal = env.Get("SRC_HTTP_ADDR_INTERNAL", ":3090", "HTTP listen address for internal HTTP API. This should never be exposed externally, as it lacks certain authz checks.")

	appURL                  = conf.GetTODO().AppURL
	disableBrowserExtension = conf.GetTODO().DisableBrowserExtension

	enableHSTS = env.Get("SG_ENABLE_HSTS", "false", "enable HTTP Strict Transport Security")

	tlsCert = conf.GetTODO().TlsCert
	tlsKey  = conf.GetTODO().TlsKey

	httpToHttpsRedirect = conf.GetTODO().HttpToHttpsRedirect

	// dev browser browser extension ID. You can find this by going to chrome://extensions
	devExtension = "chrome-extension://bmfbcejdknlknpncfpeloejonjoledha"
	// production browser extension ID. This is found by viewing our extension in the chrome store.
	prodExtension = "chrome-extension://dgjhfomjieaadpoljlnidmbgkdffpack"
)

func configureAppURL() (*url.URL, error) {
	var hostPort string
	if strings.HasPrefix(httpAddr, ":") {
		// Prepend localhost if HTTP listen addr is just a port.
		hostPort = "127.0.0.1" + httpAddr
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

// Main is the main entrypoint for the frontend server program.
func Main() error {
	log.SetFlags(0)
	log.SetPrefix("")

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			log.Printf("Version: %s", env.Version)
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

	conf.Validate()

	cleanup := tmpfriend.SetupOrNOOP()
	defer cleanup()

	logHandler := log15.StderrHandler

	// Filter trace logs
	d, _ := time.ParseDuration(traceThreshold)
	logHandler = log15.FilterHandler(loghandlers.Trace(strings.Fields(trace), d), logHandler)

	// Filter log output by level.
	lvl, err := log15.LvlFromString(env.LogLevel)
	if err != nil {
		return err
	}
	log15.Root().SetHandler(log15.LvlFilterHandler(lvl, logHandler))

	tracer.Init("frontend")

	// Don't proceed if system requirements are missing, to avoid
	// presenting users with a half-working experience.
	if err := checkSysReqs(context.Background(), os.Stderr); err != nil {
		return err
	}

	go debugserver.Start()

	db.ConnectToDB("")

	siteid.Init()

	go bg.ApplyUserOrgMap(context.Background())
	go bg.MigrateAdminUsernames(context.Background())
	goroutine.Go(func() {
		bg.MigrateOrgSlackWebhookURLs(context.Background())
	})
	goroutine.Go(func() {
		bg.StartLangServers(context.Background())
	})
	go updatecheck.Start()
	go useractivity.MigrateUserActivityData(context.Background())

	globals.AppURL, err = configureAppURL()
	if err != nil {
		return err
	}
	db.AppURL = globals.AppURL

	tlsCertAndKey := tlsCert != "" && tlsKey != ""
	useTLS := httpsAddr != "" && (tlsCertAndKey || (globals.AppURL.Scheme == "https" && conf.GetTODO().TlsLetsencrypt != "off"))
	if useTLS && globals.AppURL.Scheme == "http" {
		log15.Warn("TLS is enabled but app url scheme is http", "appURL", globals.AppURL)
	}

	sm := http.NewServeMux()
	sm.Handle("/.api/", gziphandler.GzipHandler(httpapi.NewHandler(router.New(mux.NewRouter().PathPrefix("/.api/").Subrouter()))))
	sm.Handle("/", handlerutil.NewHandlerWithCSRFProtection(app.NewHandler(), globals.AppURL.Scheme == "https"))
	assets.Mount(sm)

	var h http.Handler = sm
	h = middleware.SourcegraphComGoGetHandler(h)
	h = middleware.BlackHole(h)
	h = tracepkg.Middleware(h)
	h = secureHeadersMiddleware(h)
	// 🚨 SECURITY: Verify user identity if required
	h, err = auth.NewAuthHandler(context.Background(), h, appURL)
	if err != nil {
		return err
	}
	// Don't leak memory through gorilla/session items stored in context
	h = gcontext.ClearHandler(h)

	// The internal HTTP handler does not include the auth handlers.
	var internalHandler http.Handler
	{
		internalMux := http.NewServeMux()
		internalMux.Handle("/.internal/", gziphandler.GzipHandler(httpapi.NewInternalHandler(router.NewInternal(mux.NewRouter().PathPrefix("/.internal/").Subrouter()))))
		internalMux.Handle("/.api/", gziphandler.GzipHandler(httpapi.NewHandler(router.New(mux.NewRouter().PathPrefix("/.api/").Subrouter()))))
		internalHandler = gcontext.ClearHandler(internalMux)
	}

	// serve will serve h on l. It additionally handles graceful restarts.
	srv := &httpServers{}

	// Start HTTPS server.
	if useTLS {
		tlsConf := &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
		}

		// Configure tlsConf
		if tlsCertAndKey {
			// Manual
			cert, err := tls.X509KeyPair([]byte(tlsCert), []byte(tlsKey))
			if err != nil {
				return err
			}
			tlsConf.Certificates = []tls.Certificate{cert}
		} else {
			// LetsEncrypt
			m := &autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(globals.AppURL.Host),
				Cache:      db.CertCache,
			}
			tlsConf.GetCertificate = m.GetCertificate
			// We register paths on our HTTP handler so that we can do ACME
			// "http-01" challenges. We are required to run the port 80
			// handler since that is the only challenge ACME will issue us
			// that we can accept.
			srv.SetWrapper(m.HTTPHandler)
			if httpAddr == "" {
				log.Fatal("HTTP is disabled but is required to serve HTTPS with Lets Encrypt")
			}
		}

		l, err := net.Listen("tcp", httpsAddr)
		if err != nil {
			// Fatal if we manually specified TLS or enforce lets encrypt
			if tlsCertAndKey || conf.GetTODO().TlsLetsencrypt == "on" {
				log.Fatalf("Could not bind to address %s: %v", httpsAddr, err)
			} else {
				log15.Warn("Failed to bind to HTTPS port, TLS disabled", "address", httpsAddr, "error", err)
			}
		}

		if l != nil {
			l = tls.NewListener(l, tlsConf)
			log15.Debug("HTTPS running", "on", l.Addr())
			srv.GoServe(l, &http.Server{
				Handler:      h,
				ReadTimeout:  75 * time.Second,
				WriteTimeout: 60 * time.Second,
			})
		}
	}

	// Start HTTP server.
	if httpAddr != "" {
		l, err := net.Listen("tcp", httpAddr)
		if err != nil {
			return err
		}

		if httpToHttpsRedirect {
			// Use JS for the redirect because this is the most reliable solution if reverse proxies are involved.
			h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`<script>window.location.protocol = "https:";</script>`))
			})
		}

		log15.Debug("HTTP running", "on", httpAddr)
		srv.GoServe(l, &http.Server{
			Handler:      h,
			ReadTimeout:  75 * time.Second,
			WriteTimeout: 60 * time.Second,
		})
	}

	if httpAddrInternal != "" {
		l, err := net.Listen("tcp", httpAddrInternal)
		if err != nil {
			return err
		}

		log15.Debug("HTTP (internal) running", "on", httpAddrInternal)
		srv.GoServe(l, &http.Server{
			Handler:     internalHandler,
			ReadTimeout: 75 * time.Second,
			// Higher since for internal RPCs which can have large responses (eg git archive)
			WriteTimeout: 10 * time.Minute,
		})
	}

	go func() {
		<-processrestart.WillRestart
		// Block forever so we don't return from main func and exit this process. Package processrestart takes care
		// of killing and restarting this process externally.
		srv.wg.Add(1)

		log15.Debug("Stopping HTTP server due to imminent restart")
		srv.Close()
	}()

	if printLogo {
		fmt.Println(" ")
		fmt.Println(logoColor)
		fmt.Println(" ")
	}
	fmt.Printf("✱ Sourcegraph is ready at: %s\n", appURL)

	srv.Wait()
	return nil
}

type httpServers struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	servers []*http.Server
	wrapper func(http.Handler) http.Handler
}

// SetWrapper will set the wrapper for serve. All handlers served by are
// passed through w.
func (s *httpServers) SetWrapper(w func(http.Handler) http.Handler) {
	s.mu.Lock()
	s.wrapper = w
	s.mu.Unlock()
}

// GoServe serves srv in a new goroutine. If serve returns an error other than
// http.ErrServerClosed it will fatal.
func (s *httpServers) GoServe(l net.Listener, srv *http.Server) {
	s.addServer(srv)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := srv.Serve(l); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
}

func (s *httpServers) addServer(srv *http.Server) *http.Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.wrapper != nil {
		srv.Handler = s.wrapper(srv.Handler)
	}
	s.servers = append(s.servers, srv)
	return srv
}

// Close closes all servers added
func (s *httpServers) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, srv := range s.servers {
		srv.Close()
	}
	s.servers = nil
}

// Wait waits until all servers are closed.
func (s *httpServers) Wait() {
	s.wg.Wait()
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, o := range allowedOrigins {
		if o == origin {
			return true
		}
	}
	return false
}
