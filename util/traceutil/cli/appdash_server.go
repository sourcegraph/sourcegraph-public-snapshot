package cli

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
	sgxcli "sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil"
)

func init() {
	_, err := sgxcli.Internal.AddCommand("serve-appdash",
		"start appdash server",
		"The serve-appdash command starts a standalone Appdash server.",
		&serverCmdFlags,
	)
	if err != nil {
		log.Fatal(err)
	}

	sgxcli.PostInit = append(sgxcli.PostInit, func() {
		if _, err := sgxcli.Serve.AddGroup("Appdash server", "Appdash server", &serverGroupFlags); err != nil {
			log.Fatal(err)
		}
	})

	sgxcli.ServeInit = append(sgxcli.ServeInit, func() {
		if err := serverGroupFlags.configureAndStart(true); err != nil {
			log.Fatal("Error configuring and starting appdash server:", err)
		}
	})

	// Server config is used to initialize client flags, so ensure
	// server configureAndStart runs before clientFlags.configure.
	initClient()
}

var serverCmdFlags, serverGroupFlags ServerConfig

type ServerConfig struct {
	Disable bool `long:"appdash.disable-server" description:"don't run an appdash server (neither collector nor web UI)" env:"SRC_APPDASH_DISABLE_SERVER"`

	URL         string `long:"appdash.url" description:"externally accessible URL for Appdash's web UI" default:"http://localhost:7800" env:"SRC_APPDASH_URL"`
	HTTPAddr    string `long:"appdash.http-addr" description:"http bind address for background appdash" default:":7800" env:"SRC_APPDASH_HTTP_ADDR"`
	TLSCertFile string `long:"appdash.tls-cert" description:"certificate file for HTTP and collector TLS (if not set, TLS is disabled)" env:"SRC_APPDASH_TLS_CERT"`
	TLSKeyFile  string `long:"appdash.tls-key" description:"key file for HTTP and collector TLS (if not set, TLS is disabled)" env:"SRC_APPDASH_TLS_KEY"`

	HTTPBasicAuthUser     string `long:"appdash.http-basic-auth-user" description:"username required for basic auth (only used if set)" env:"SRC_APPDASH_HTTP_BASIC_AUTH_USER"`
	HTTPBasicAuthPassword string `long:"appdash.http-basic-auth-password" description:"password required for basic auth" env:"SRC_APPDASH_HTTP_BASIC_AUTH_PASSWORD"`

	CollectorAddr string `long:"appdash.collector-addr" description:"TCP collector bind address for background appdash ('127.0.0.1:0' for randomly chosen)" default:"127.0.0.1:0" env:"SRC_APPDASH_COLLECTOR_ADDR"`
	CollectorTLS  bool   `long:"appdash.collector-tls" description:"whether or not the collector should use TLS" env:"SRC_APPDASH_COLLECTOR_TLS"`
	LogDebug      bool   `long:"appdash.log-debug" description:"enable appdash debug logging" env:"SRC_APPDASH_LOG_DEBUG"`
	LogTrace      bool   `long:"appdash.log-trace" description:"enable appdash trace logging" env:"SRC_APPDASH_LOG_TRACE"`

	// InfluxDB-specific configuration settings.
	InfluxLogDir       string `long:"appdash.influx-log-dir" description:"InfluxDB log directory" default:"$SGPATH" ENV:"SRC_APPDASH_INFLUX_LOG_DIR"`
	InfluxAddr         string `long:"appdash.influx-addr" description:"InfluxDB cluster-wide communication service bind address" default:":8088" env:"SRC_APPDASH_INFLUX_ADDR"`
	InfluxAdminAddr    string `long:"appdash.influx-admin-addr" description:"InfluxDB admin service bind address" default:":8083" env:"SRC_APPDASH_INFLUX_ADMIN_ADDR"`
	InfluxCollectdAddr string `long:"appdash.influx-collectd-addr" description:"InfluxDB collectd service bind address" default:"" env:"SRC_APPDASH_INFLUX_COLLECTD_ADDR"`
	InfluxGraphiteAddr string `long:"appdash.influx-graphite-addr" description:"InfluxDB graphite service bind address" default:":2003" env:"SRC_APPDASH_INFLUX_GRAPHITE_ADDR"`
	InfluxHTTPDAddr    string `long:"appdash.influx-httpd-addr" description:"InfluxDB HTTPD service bind address" default:":8086" env:"SRC_APPDASH_INFLUX_HTTPD_ADDR"`
	InfluxOpenTSDBAddr string `long:"appdash.influx-opentsdb-addr" description:"InfluxDB OpenTSDB service bind address" default:":4242" env:"SRC_APPDASH_INFLUX_OPENTSDB_ADDR"`
	InfluxUDPAddr      string `long:"appdash.influx-udp-addr" description:"InfluxDB UDP service bind address" default:"" env:"SRC_APPDASH_INFLUX_UDP_ADDR"`

	// Deprecated flags, to be removed soon!
	//
	// TODO(slimsag): Remove these after May 19, 2016.
	NSlowest        int `long:"appdash.n-slowest" description:"deprecated; has no effect" default:"5" env:"SRC_APPDASH_N_SLOWEST"`
	MaxRate         int `long:"appdash.max-rate" description:"deprecated; has no effect" default:"4096" env:"SRC_APPDASH_MAX_RATE"`
	KeepMax         int `long:"appdash.keep-max" description:"deprecated; has no effect" default:"2000" env:"SRC_APPDASH_KEEP_MAX"`
	MinRecentTraces int `long:"appdash.min-recent-traces" description:"deprecated; has no effect" default:"5" env:"SRC_APPDASH_MIN_RECENT_TRACES"`
}

// configureAndStart starts Appdash servers per the configuration and
// updates clientFlags to point to the server
//
// if serveInGoroutine is true then serving the Appdash UI will occur in a
// separate goroutine (rather than blocking).
func (f *ServerConfig) configureAndStart(serveInGoroutine bool) error {
	if f.Disable {
		log15.Debug("Appdash server (collector and web UI) is disabled")
		return nil
	}

	// Create a default InfluxDB configuration.
	conf, err := appdash.NewInfluxDBConfig()
	if err != nil {
		return fmt.Errorf("failed to create influxdb config, error: %v", err)
	}

	// Usage metrics are non-invasive; but we disable them anyway to get rid of
	// an annoying log message about it.
	conf.Server.ReportingDisabled = true

	// Setup address configurations.
	conf.Server.BindAddress = f.InfluxAddr
	conf.Server.Admin.BindAddress = f.InfluxAdminAddr
	conf.Server.CollectdInputs[0].BindAddress = f.InfluxCollectdAddr
	conf.Server.GraphiteInputs[0].BindAddress = f.InfluxGraphiteAddr
	conf.Server.HTTPD.BindAddress = f.InfluxHTTPDAddr
	conf.Server.OpenTSDBInputs[0].BindAddress = f.InfluxOpenTSDBAddr
	conf.Server.UDPInputs[0].BindAddress = f.InfluxUDPAddr

	// InfluxDB needs an admin user in the config, even if not using HTTP basic
	// auth.
	conf.Server.HTTPD.AuthEnabled = true
	if f.HTTPBasicAuthUser != "" {
		conf.AdminUser.Username = f.HTTPBasicAuthUser
		conf.AdminUser.Password = f.HTTPBasicAuthPassword
	} else {
		conf.AdminUser.Username = "sourcegraph"
		conf.AdminUser.Password = "sourcegraph"
	}

	logOutputName := ""
	if f.LogDebug {
		logOutputName = "(stderr)"
		conf.LogOutput = os.Stderr
	} else if f.InfluxLogDir != "" {
		logOutputName = os.ExpandEnv(filepath.Join(f.InfluxLogDir, "influxdb.log"))
		logFile, err := os.OpenFile(logOutputName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open influxdb log file, error: %v", err)
		}
		conf.LogOutput = logFile
	}

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

	// Configure appdash client.
	clientFlags.RemoteAddr = l.Addr().String()
	clientFlags.TLS = collectorUseTLS
	clientFlags.url = f.URL

	serve := func() {
		log15.Info("InfluxDB server starting", "logoutput", logOutputName)
		store, err := appdash.NewInfluxDBStore(conf)
		if err != nil {
			log.Fatalf("failed to create influxdb store, error: %v", err)
		}
		log15.Info("InfluxDB server started")

		cs := appdash.NewServer(l, appdash.NewLocalCollector(store))
		cs.Debug = f.LogDebug
		cs.Trace = f.LogTrace
		go cs.Start()

		// Create a MultiQueryer that will query both our LimitStore and our
		// AggregateStore for traces.
		appdashURL, err := url.Parse(f.URL)
		if err != nil {
			log.Fatalln("failed to parse --appdash.url, error:", err)
		}
		app, err := traceapp.New(nil, appdashURL)
		if err != nil {
			log.Fatalln(err)
		}
		app.Store = store
		app.Queryer = store
		app.Aggregator = store

		// Setup basic authentication if desired.
		h := http.Handler(app)
		if f.HTTPBasicAuthUser != "" {
			h = httputil.BasicAuth(f.HTTPBasicAuthUser, f.HTTPBasicAuthPassword, 0, app)
		}

		log15.Debug("Appdash server running", "web", f.HTTPAddr, "collector", l.Addr(), "proto", proto)

		if useTLS {
			log.Fatal(http.ListenAndServeTLS(f.HTTPAddr, f.TLSCertFile, f.TLSKeyFile, h))
		} else {
			log.Fatal(http.ListenAndServe(f.HTTPAddr, h))
		}
		if err := store.Close(); err != nil {
			log.Fatal(err)
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
func (f *ServerConfig) Execute(args []string) error {
	return f.configureAndStart(false)
}
