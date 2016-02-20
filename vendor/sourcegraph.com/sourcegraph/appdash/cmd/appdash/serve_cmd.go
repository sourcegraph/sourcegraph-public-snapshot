package main

import (
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"strings"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
)

func init() {
	_, err := CLI.AddCommand("serve",
		"start an appdash server",
		"The serve command starts an appdash server.",
		&serveCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

// ServeCmd is the command for running Appdash in server mode, where a
// collector server and the web UI are hosted.
type ServeCmd struct {
	CollectorAddr string `long:"collector" description:"collector listen address" default:":7701"`
	HTTPAddr      string `long:"http" description:"HTTP listen address" default:":7700"`
	SampleData    bool   `long:"sample-data" description:"add sample data"`

	StoreFile       string        `short:"f" long:"store-file" description:"persisted store file" default:"/tmp/appdash.gob"`
	PersistInterval time.Duration `short:"p" long:"persist-interval" description:"interval between persisting store to file" default:"2s"`

	Debug bool `short:"d" long:"debug" description:"debug log"`
	Trace bool `long:"trace" description:"trace log"`

	DeleteAfter time.Duration `long:"delete-after" description:"delete traces after a certain age (0 to disable)" default:"30m"`

	TLSCert string `long:"tls-cert" description:"TLS certificate file (if set, enables TLS)"`
	TLSKey  string `long:"tls-key" description:"TLS key file (if set, enables TLS)"`

	BasicAuth string `long:"basic-auth" description:"if set to 'user:passwd', require HTTP Basic Auth for web app"`
}

var serveCmd ServeCmd

// Execute execudes the commands with the given arguments and returns an error,
// if any.
func (c *ServeCmd) Execute(args []string) error {
	var (
		memStore = appdash.NewMemoryStore()
		Store    = appdash.Store(memStore)
		Queryer  = memStore
	)

	if c.StoreFile != "" {
		f, err := os.Open(c.StoreFile)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if f != nil {
			if n, err := memStore.ReadFrom(f); err == nil {
				log.Printf("Read %d traces from file %s", n, c.StoreFile)
			} else if err != nil {
				f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		}
		if c.PersistInterval != 0 {
			go func() {
				if err := appdash.PersistEvery(memStore, c.PersistInterval, c.StoreFile); err != nil {
					log.Fatal(err)
				}
			}()
		}
	}

	if c.DeleteAfter > 0 {
		Store = &appdash.RecentStore{
			MinEvictAge: c.DeleteAfter,
			DeleteStore: memStore,
			Debug:       true,
		}
	}

	app := traceapp.New(nil)
	app.Store = Store
	app.Queryer = Queryer

	var h http.Handler
	if c.BasicAuth != "" {
		parts := strings.SplitN(c.BasicAuth, ":", 2)
		if len(parts) != 2 {
			log.Fatalf("Basic auth must be specified as 'user:passwd'.")
		}
		user, passwd := parts[0], parts[1]
		if user == "" || passwd == "" {
			log.Fatalf("Basic auth user and passwd must both be nonempty.")
		}
		log.Printf("Requiring HTTP Basic auth")
		h = newBasicAuthHandler(user, passwd, app)
	} else {
		h = app
	}

	if c.SampleData {
		sampleData(Store)
	}

	var l net.Listener
	var proto string
	if c.TLSCert != "" || c.TLSKey != "" {
		certBytes, err := ioutil.ReadFile(c.TLSCert)
		if err != nil {
			log.Fatal(err)
		}
		keyBytes, err := ioutil.ReadFile(c.TLSKey)
		if err != nil {
			log.Fatal(err)
		}

		var tc tls.Config
		cert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			log.Fatal(err)
		}
		tc.Certificates = []tls.Certificate{cert}
		l, err = tls.Listen("tcp", c.CollectorAddr, &tc)
		if err != nil {
			log.Fatal(err)
		}
		proto = fmt.Sprintf("TLS cert %s, key %s", c.TLSCert, c.TLSKey)
	} else {
		var err error
		l, err = net.Listen("tcp", c.CollectorAddr)
		if err != nil {
			log.Fatal(err)
		}
		proto = "plaintext TCP (no security)"
	}
	log.Printf("appdash collector listening on %s (%s)", c.CollectorAddr, proto)
	cs := appdash.NewServer(l, appdash.NewLocalCollector(Store))
	cs.Debug = c.Debug
	cs.Trace = c.Trace
	go cs.Start()

	if c.TLSCert != "" || c.TLSKey != "" {
		log.Printf("appdash HTTPS server listening on %s (TLS cert %s, key %s)", c.HTTPAddr, c.TLSCert, c.TLSKey)
		return http.ListenAndServeTLS(c.HTTPAddr, c.TLSCert, c.TLSKey, h)
	}

	log.Printf("appdash HTTP server listening on %s", c.HTTPAddr)
	return http.ListenAndServe(c.HTTPAddr, h)
}

func newBasicAuthHandler(user, passwd string, h http.Handler) http.Handler {
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, passwd)))
	return &basicAuthHandler{h, []byte(want)}
}

type basicAuthHandler struct {
	http.Handler
	want []byte // = "Basic " base64(user ":" passwd) [precomputed]
}

func (h *basicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Constant time comparison to avoid timing attack.
	authHdr := r.Header.Get("authorization")
	if len(h.want) == len(authHdr) && subtle.ConstantTimeCompare(h.want, []byte(authHdr)) == 1 {
		h.Handler.ServeHTTP(w, r)
		return
	}
	w.Header().Set("WWW-Authenticate", `Basic realm="appdash"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
