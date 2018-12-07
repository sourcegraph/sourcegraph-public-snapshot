//docker:user sourcegraph

// Postgres defaults for cluster deployments.
//docker:env PGDATABASE=sg
//docker:env PGHOST=pgsql
//docker:env PGPORT=5432
//docker:env PGSSLMODE=disable
//docker:env PGUSER=sg

// The management console provides a failsafe editor for the critical
// configuration options for the Sourcegraph instance.
//
// 🚨 SECURITY: No authentication is done by the management console.
// It is currently the user's responsibility to:
//
// 1. Limit access to the management console by not exposing its port.
// 2. Ensure that the management console's responses are never propagated to
//    unprivileged users.
//
package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/management-console/assets"
	"github.com/sourcegraph/sourcegraph/pkg/db/confdb"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
)

const port = "2633"

func main() {
	env.Lock()
	env.HandleHelpFlag()

	tracer.Init()

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

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(assets.Assets))
	mux.HandleFunc("/get", serveGet)
	mux.HandleFunc("/update", serveUpdate)
	http.Handle("/", mux)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	log15.Info("management-console: listening", "addr", addr)
	log.Fatalf("Fatal error serving: %s", http.ListenAndServe(addr, nil))
}

func serveGet(w http.ResponseWriter, r *http.Request) {
	logger := log15.New("route", "get")

	critical, err := confdb.CriticalGetLatest(r.Context())
	if err != nil {
		logger.Error("confdb.CriticalGetLatest failed", "error", err)
		http.Error(w, "Error retrieving latest critical configuration.", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(&struct {
		ID       string
		Contents string
	}{
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

	var args struct {
		LastID   string `json:"lastID"`
		Contents string `json:"contents"`
	}
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		logger.Error("json argument decoding failed", "error", err)
		http.Error(w, "Unexpected error when decoding arguments.", http.StatusBadRequest)
		return
	}

	lastID, err := strconv.Atoi(args.LastID)
	lastIDInt32 := int32(lastID)
	if err != nil {
		logger.Error("argument LastID decoding failed", "error", err)
		http.Error(w, "Unexpected error when decoding LastID argument.", http.StatusBadRequest)
		return
	}

	critical, err := confdb.CriticalCreateIfUpToDate(r.Context(), &lastIDInt32, args.Contents)
	if err != nil {
		logger.Error("confdb.CriticalCreateIfUpToDate failed", "error", err)
		http.Error(w, "Error updating latest critical configuration.", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(critical)
	if err != nil {
		logger.Error("json response encoding failed", "error", err)
		http.Error(w, "Error encoding json response.", http.StatusInternalServerError)
	}
}
