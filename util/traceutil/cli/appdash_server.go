package cli

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
	sgxcli "sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil"
)

func init() {
	_, err := sgxcli.Internal.AddCommand("serve-appdash",
		"start appdash server",
		"The serve-appdash command starts a standalone Appdash server.",
		&cmdFlags,
	)
	if err != nil {
		log.Fatal(err)
	}

	sgxcli.PostInit = append(sgxcli.PostInit, func() {
		if _, err := sgxcli.Serve.AddGroup("Appdash server", "Appdash server", &groupFlags); err != nil {
			log.Fatal(err)
		}
	})

	sgxcli.ServeInit = append(sgxcli.ServeInit, func() {
		if err := groupFlags.configureAndStart(true); err != nil {
			log.Fatal("Error configuring and starting appdash server:", err)
		}
	})

	// Server config is used to initialize client flags, so ensure
	// server configureAndStart runs before clientFlags.configure.
	initClient()
}

var cmdFlags, groupFlags ServerFlags

type ServerFlags struct {
	Disable bool `long:"appdash.disable-server" description:"don't run an appdash server (neither collector nor web UI)"`

	HTTPAddr    string `long:"appdash.http-addr" description:"http bind address for background appdash" default:":7800"`
	TLSCertFile string `long:"appdash.tls-cert" description:"certificate file for HTTP and collector TLS (if not set, TLS is disabled)"`
	TLSKeyFile  string `long:"appdash.tls-key" description:"key file for HTTP and collector TLS (if not set, TLS is disabled)"`

	HTTPBasicAuthUser     string `long:"appdash.http-basic-auth-user" description:"username required for basic auth (only used if set)"`
	HTTPBasicAuthPassword string `long:"appdash.http-basic-auth-password" description:"password required for basic auth"`

	CollectorAddr   string `long:"appdash.collector-addr" description:"TCP collector bind address for background appdash ('127.0.0.1:0' for randomly chosen)" default:"127.0.0.1:0"`
	CollectorTLS    bool   `long:"appdash.collector-tls" description:"whether or not the collector should use TLS"`
	NSlowest        int    `long:"appdash.n-slowest" description:"number of slowest traces to keep for a URL route (before deleting oldest)" default:"5"`
	MaxRate         int    `long:"appdash.max-rate" description:"maximum expected rate of concurrent requests (slowest traces will be missed otherwise)" default:"4096"`
	KeepMax         int    `long:"appdash.keep-max" description:"max number of recent traces to keep (before deleting oldest)" default:"2000"`
	MinRecentTraces int    `long:"appdash.min-recent-traces" description:"number of minutes of recent traces to keep in storage" default:"5"`
	LogDebug        bool   `long:"appdash.log-debug" description:"enable appdash debug logging"`
	LogTrace        bool   `long:"appdash.log-trace" description:"enable appdash trace logging"`
}

// configureAndStart starts Appdash servers per the configuration and
// updates clientFlags to point to the server
//
// if serveInGoroutine is true then serving the Appdash UI will occur in a
// separate goroutine (rather than blocking).
func (f *ServerFlags) configureAndStart(serveInGoroutine bool) error {
	if f.Disable {
		log15.Debug("Appdash server (collector and web UI) is disabled")
		return nil
	}

	// Create a recent store, writing out recent traces to our own MemoryStore.
	recentStore := &appdash.LimitStore{
		Max:         f.KeepMax,                // up to N recent traces.
		DeleteStore: appdash.NewMemoryStore(), // use our own backing MemoryStore
	}

	// Create an aggregate store, writing out aggregated traces to our own
	// MemoryStore.
	//
	// TODO(slimsag): MinEvictAge of 72/hrs is hard-coded in the UI in some places,
	// so we can't expose it as variable for now.
	aggStore := &appdash.AggregateStore{
		MinEvictAge: 72 * time.Hour,           // up to N hours of aggregated timespans
		MaxRate:     f.MaxRate,                // expected maximum rate of concurrent trace collections
		NSlowest:    f.NSlowest,               // keep the N slowest full traces
		MemoryStore: appdash.NewMemoryStore(), // use our own backing MemoryStore
		Debug:       f.LogDebug,
	}

	// Make it such that collections are sent to both the LimitStore and
	// AggregateStore (so we keep a limited number of recent traces in addition
	// to the aggregated traces).
	store := appdash.MultiStore(recentStore, aggStore)

	useTLS := f.TLSCertFile != "" || f.TLSKeyFile != ""

	var l net.Listener
	var proto string
	const listenNet = "tcp4" // IPv6 causes some issues inside some environments (e.g., Mesos)
	collectorUseTLS := f.CollectorTLS && useTLS && f.CollectorAddr != "127.0.0.1:0"
	if collectorUseTLS {
		certBytes, err := ioutil.ReadFile(f.TLSCertFile)
		if err != nil {
			return err
		}
		keyBytes, err := ioutil.ReadFile(f.TLSKeyFile)
		if err != nil {
			return err
		}

		var tc tls.Config
		cert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return err
		}
		tc.Certificates = []tls.Certificate{cert}
		l, err = tls.Listen(listenNet, f.CollectorAddr, &tc)
		if err != nil {
			return err
		}
		proto = fmt.Sprintf("TLS with cert %s, key %s", f.TLSCertFile, f.TLSKeyFile)
	} else {
		var err error
		l, err = net.Listen(listenNet, f.CollectorAddr)
		if err != nil {
			return err
		}
		proto = "plaintext (non-TLS) TCP"
	}
	cs := appdash.NewServer(l, appdash.NewLocalCollector(store))
	cs.Debug = f.LogDebug
	cs.Trace = f.LogTrace
	go cs.Start()

	// Configure appdash client.
	clientFlags.RemoteAddr = l.Addr().String()
	clientFlags.TLS = collectorUseTLS

	// Create a MultiQueryer that will query both our LimitStore and our
	// AggregateStore for traces.
	app := traceapp.New(nil)
	app.Store = store
	app.Queryer = appdash.MultiQueryer(
		recentStore.DeleteStore.(*appdash.MemoryStore),
		aggStore.MemoryStore,
	)

	// Setup basic authentication if desired.
	h := http.Handler(app)
	if f.HTTPBasicAuthUser != "" {
		h = httputil.BasicAuth(f.HTTPBasicAuthUser, f.HTTPBasicAuthPassword, 0, app)
	}

	log15.Debug("Appdash server running", "web", f.HTTPAddr, "collector", l.Addr(), "proto", proto)
	serve := func() {
		if useTLS {
			log.Fatal(http.ListenAndServeTLS(f.HTTPAddr, f.TLSCertFile, f.TLSKeyFile, h))
		} else {
			log.Fatal(http.ListenAndServe(f.HTTPAddr, h))
		}
	}
	if serveInGoroutine {
		go serve()
	} else {
		serve()
	}
	return nil
}

// Execute treats these flags like a command so that `src internal serve-appdash <flags>`
// works properly (such that Appdash can be ran independently from the rest of
// the code in our binary).
func (f *ServerFlags) Execute(args []string) error {
	return f.configureAndStart(false)
}
