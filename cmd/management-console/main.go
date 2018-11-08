//docker:user sourcegraph
// Postgres defaults for cluster deployments.
//docker:env PGDATABASE=sg
//docker:env PGHOST=pgsql
//docker:env PGPORT=5432
//docker:env PGSSLMODE=disable
//docker:env PGUSER=sg
//
// The management console provides a failsafe editor for the core configuration
// options for the Sourcegraph instance.
//
// 🚨 SECURITY: No authentication is done by the management console.
// It is the user's responsibility to:
//    1) Limit access to the management console by not exposing its port.
//    2) Ensure that the management console's responses are never propagated to
//    unprivileged users.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/conf/confdb"
	"github.com/sourcegraph/sourcegraph/pkg/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	dbHandler = confdb.CoreSiteConfigurationFiles{}
	address   = flag.String("addr", ":6075", "management console TCP listening address")
)

func main() {
	flag.Parse()
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	err := dbconn.ConnectToDB("")
	if err != nil {
		return errors.Wrap(err, "unable to connect to database")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/get", serveCoreConfigurationGetLatest)
	mux.HandleFunc("/create", serveCoreConfigurationCreateIfUpToDate)

	if env.InsecureDev && strings.HasPrefix(*address, ":") {
		*address = net.JoinHostPort("127.0.0.1", (*address)[1:])
	}
	log15.Info("Management console: listening", "address", *address)

	return http.ListenAndServe(*address, mux)
}

func serveCoreConfigurationGetLatest(w http.ResponseWriter, r *http.Request) {
	logger := log15.New("route", "serveCoreConfigurationGetLatest")

	coreFile, err := dbHandler.CoreGetLatest(r.Context())
	if err != nil {
		logger.Error("dbHandler.CoreGetLatest failed", "error", err)
		http.Error(w, "Unexpected error when retrieving latest core configuration file.", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(coreFile)
	if err != nil {
		logger.Error("json response encoding failed", "error", err)
		http.Error(w, "Unexpected error when encoding json response.", http.StatusInternalServerError)
	}
}

func serveCoreConfigurationCreateIfUpToDate(w http.ResponseWriter, r *http.Request) {
	logger := log15.New("route", "serveCoreConfigurationCreateIfUpToDate")

	var args createIfUpToDateArgs
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		logger.Error("json argument decoding failed", "error", err)
		http.Error(w, "Unexpected error when decoding arguments.", http.StatusBadRequest)
		return
	}

	coreFile, err := dbHandler.CoreCreateIfUpToDate(r.Context(), args.LastID, args.Contents)
	if err != nil {
		logger.Error("dbHandler.CoreCreateIfUpToDate failed", "error", err)
		http.Error(w, "Unexpected error when updating latest core configuration file.", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(coreFile)
	if err != nil {
		logger.Error("json response encoding failed", "error", err)
		http.Error(w, "Unexpected error when encoding json response.", http.StatusInternalServerError)
	}
}

type createIfUpToDateArgs struct {
	LastID   *int32 `json:"lastID,omitempty"`
	Contents string `json:"contents"`
}
