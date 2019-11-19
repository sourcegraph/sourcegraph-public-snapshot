// Package shared contains the shared management console implementation.
//
// The management console provides a failsafe editor for critical Sourcegraph
// configuration which, if changed correctly, could prevent access to the
// Sourcegraph instance.
package shared

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/management-console/assets"
	"github.com/sourcegraph/sourcegraph/internal/db/confdb"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"gopkg.in/inconshreveable/log15.v2"
)

const port = "2633"

var (
	disableConfigUpdates = env.Get("DISABLE_CONFIG_UPDATES", "false", "When true, disables updating the configuration. Useful when using CRITICAL_CONFIG_FILE on the frontend service.")
)

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

	routes := http.NewServeMux()
	routes.Handle("/", http.FileServer(assets.Assets))
	routes.HandleFunc("/api/get", serveGet)
	routes.HandleFunc("/api/update", serveUpdate)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	log15.Info("management-console: listening", "addr", addr)

	s := &http.Server{
		Addr:           addr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.Handler = routes
	log.Fatalf("Fatal error serving: %s", s.ListenAndServe())
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
		httpError(w, "Error retrieving latest critical configuration.", "internal_error")
		return
	}

	err = json.NewEncoder(w).Encode(&jsonConfiguration{
		ID:       strconv.Itoa(int(critical.ID)),
		Contents: critical.Contents,
	})
	if err != nil {
		logger.Error("json response encoding failed", "error", err)
		httpError(w, "Error encoding json response.", "internal_error")
	}
}

func httpError(w http.ResponseWriter, message string, code string) {
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}{
		Error: message,
		Code:  code,
	})
}

// serveUpdate updates the critical site configuration. It is eventually consistent--there are no
// guarantees that the configuration has propagated to all services on return, but this typically
// happens within a few seconds.
func serveUpdate(w http.ResponseWriter, r *http.Request) {
	logger := log15.New("route", "update")

	disableConfigUpdates, _ := strconv.ParseBool(disableConfigUpdates)
	if disableConfigUpdates {
		httpError(w, errors.New("Updating configuration was disabled via DISABLE_CONFIG_UPDATES").Error(), "config_updates_disabled")
		return
	}

	var args struct {
		LastID   string
		Contents string
	}
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		logger.Error("json argument decoding failed", "error", err)
		httpError(w, errors.Wrap(err, "Unexpected error when decoding arguments").Error(), "bad_request")
		return
	}

	lastID, err := strconv.Atoi(args.LastID)
	lastIDInt32 := int32(lastID)
	if err != nil {
		logger.Error("argument LastID decoding failed", "error", err)
		httpError(w, errors.Wrap(err, "Unexpected error when decoding LastID argument").Error(), "bad_request")
		return
	}

	err = validateConfig(args.Contents,
		validateExternalURL,
	)
	if err != nil {
		httpError(w, errors.Wrap(err, "Invalid critical configuration found").Error(), "bad_request")
		return
	}

	critical, err := confdb.CriticalCreateIfUpToDate(r.Context(), &lastIDInt32, args.Contents)
	if err != nil {
		if err == confdb.ErrNewerEdit {
			httpError(w, confdb.ErrNewerEdit.Error(), "newer_edit")
			return
		}
		logger.Error("confdb.CriticalCreateIfUpToDate failed", "error", err)
		httpError(w, errors.Wrap(err, "Error updating latest critical configuration").Error(), "internal_error")
		return
	}

	err = json.NewEncoder(w).Encode(&jsonConfiguration{
		ID:       strconv.Itoa(int(critical.ID)),
		Contents: critical.Contents,
	})
	if err != nil {
		logger.Error("json response encoding failed", "error", err)
		httpError(w, errors.Wrap(err, "Error encoding JSON response").Error(), "internal_error")
	}
}
