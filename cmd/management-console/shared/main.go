// Package shared contains the shared management console implementation.
//
// The management console provides a failsafe editor for critical Sourcegraph
// configuration which, if changed correctly, could prevent access to the
// Sourcegraph instance.
package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/management-console/assets"
	"github.com/sourcegraph/sourcegraph/cmd/management-console/shared/internal/tlscertgen"
	"github.com/sourcegraph/sourcegraph/pkg/db/confdb"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/db/globalstatedb"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

const port = "2633"

var (
	tlsCert   = env.Get("TLS_CERT", "/etc/sourcegraph/management/cert.pem", "TLS certificate (automatically generated if file does not exist)")
	tlsKey    = env.Get("TLS_KEY", "/etc/sourcegraph/management/key.pem", "TLS key (automatically generated if file does not exist)")
	customTLS = env.Get("CUSTOM_TLS", "false", "When true, disable TLS cert/key generation to prevent accidents.")
)

func configureTLS() error {
	customTLS, _ := strconv.ParseBool(customTLS)

	_, err := os.Stat(tlsCert)
	if os.IsNotExist(err) {
		if customTLS {
			return err
		}
	} else if err != nil {
		return err
	}

	_, err = os.Stat(tlsKey)
	if os.IsNotExist(err) {
		if customTLS {
			return err
		}
	} else if err != nil {
		return err
	}

	if customTLS {
		return nil // cert files exist
	}

	if err := os.MkdirAll(filepath.Dir(tlsCert), 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(tlsKey), 0700); err != nil {
		return err
	}

	// Generate a TLS cert.
	certOut, err := os.Create(tlsCert)
	if err != nil {
		return errors.Wrap(err, "failed to open cert.pem for writing")
	}
	defer certOut.Close()

	keyOut, err := os.OpenFile(tlsKey, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to open key.pem for writing")
	}
	defer keyOut.Close()

	return tlscertgen.Generate(tlscertgen.Options{
		Cert:         certOut,
		Key:          keyOut,
		Hosts:        []string{"management-console.sourcegraph.com"},
		ValidFor:     100 * 365 * 24 * time.Hour,
		ECDSACurve:   "P256",
		Organization: "Sourcegraph",
	})
}

func Main() {
	env.Lock()
	env.HandleHelpFlag()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
		<-c
		os.Exit(0)
	}()

	go debugserver.Start()

	// The management console connects directly to the DB (e.g. in case the
	// frontend is down due to a bad config).
	err := dbconn.ConnectToDB("")
	if err != nil {
		log.Fatalf("Fatal error connecting to Postgres DB: %s", err)
	}

	// 🚨 SECURITY: ALL management console API routes MUST go through the
	// authentication middleware, otherwise they would be exposed to the public
	// internet.
	protectedRoutes := http.NewServeMux()
	protectedRoutes.HandleFunc("/api/get", serveGet)
	protectedRoutes.HandleFunc("/api/update", serveUpdate)

	// Static assets are excluded from the authentication middleware because
	// they are the same for all Sourcegraph users AND because the
	// authentication middleware is intentionally slow/costly (i.e. it would
	// add ~1s to the load time of each asset).
	unprotectedRoutes := http.NewServeMux()
	unprotectedRoutes.Handle("/", http.FileServer(assets.Assets))
	unprotectedRoutes.Handle("/api/", AuthMiddleware(protectedRoutes))

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	log15.Info("management-console: listening", "addr", addr)

	if err := configureTLS(); err != nil {
		log.Fatal("failed to configure TLS: error:", err)
	}

	s := &http.Server{
		Addr:           addr,
		Handler:        HSTSMiddleware(unprotectedRoutes),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatalf("Fatal error serving: %s", s.ListenAndServeTLS(tlsCert, tlsKey))
}

type jsonConfiguration struct {
	ID       string
	Contents string
}

func serveGet(w http.ResponseWriter, r *http.Request) {
	logger := log15.New("route", "get")

	critical, err := confdb.CriticalGetLatest(r.Context())
	if err != nil {
		logger.Error("confdb.CriticalGetLatest failed", "error", err)
		http.Error(w, "Error retrieving latest critical configuration.", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(&jsonConfiguration{
		ID:       strconv.Itoa(int(critical.ID)),
		Contents: critical.Contents,
	})
	if err != nil {
		logger.Error("json response encoding failed", "error", err)
		http.Error(w, "Error encoding json response.", http.StatusInternalServerError)
	}
}

func serveUpdate(w http.ResponseWriter, r *http.Request) {
	logger := log15.New("route", "update")

	httpError := func(error_ error, code string) {
		logger.Error("Error updating critical configuration", "error", error_)
		json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}{
			Error: error_.Error(),
			Code:  code,
		})
	}

	httpSuccess := func(payload interface{}) {
		err := json.NewEncoder(w).Encode(payload)
		if err != nil {
			httpError(errors.Wrap(err, "Error encoding JSON response"), "internal_error")
		}
	}

	var args struct {
		LastID   string
		Contents string
	}
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		httpError(errors.Wrap(err, "Unexpected error when decoding arguments"), "bad_request")
		return
	}

	lastID, err := strconv.Atoi(args.LastID)
	lastIDInt32 := int32(lastID)
	if err != nil {
		httpError(errors.Wrap(err, "Unexpected error when decoding LastID argument"), "bad_request")
		return
	}

	critical, err := confdb.CriticalCreateIfUpToDate(r.Context(), &lastIDInt32, args.Contents)
	if err != nil {
		if err == confdb.ErrNewerEdit {
			httpError(confdb.ErrNewerEdit, "newer_edit")
			return
		}
		httpError(errors.Wrap(err, "Error updating latest critical configuration"), "database_error")
		return
	}

	httpSuccess(&jsonConfiguration{
		ID:       strconv.Itoa(int(critical.ID)),
		Contents: critical.Contents,
	})
}

// HSTSMiddleware effectively instructs browsers to change all HTTP requests to
// HTTPS.
func HSTSMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		h.ServeHTTP(w, r)
	})
}

// AuthMiddleware wraps h and performs authentication for ALL management
// console routes.
//
// 🚨 SECURITY: This function handles all authentication for the management
// console. Any regression here would be of EXTREME concern.
func AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, haveUserPass := r.BasicAuth()
		if !haveUserPass {
			// User has not yet been prompted for auth.
			w.Header().Set("WWW-Authenticate", `Basic realm="Sourcegraph management console (enter any username)."`)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "%s\n", http.StatusText(http.StatusUnauthorized))
			return
		}

		// Attempt authentication.
		err := globalstatedb.AuthenticateManagementConsole(r.Context(), pass)
		if err != nil {
			log15.Warn("Rejecting request with failed authentication", "username", user)
			w.Header().Set("WWW-Authenticate", `Basic realm="Sourcegraph management console (enter any username)."`)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "%s\n", http.StatusText(http.StatusUnauthorized))
			return
		}

		// Successfully authenticated.
		h.ServeHTTP(w, r)
	})
}
