//docker:user sourcegraph

// Postgres defaults for cluster deployments.
//docker:env PGDATABASE=sg
//docker:env PGHOST=pgsql
//docker:env PGPORT=5432
//docker:env PGSSLMODE=disable
//docker:env PGUSER=sg

// The management console provides a failsafe editor for the core configuration
// options for the Sourcegraph instance.
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
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
)

const port = "6075"

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

	// TODO: register HTTP handlers here!

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	log15.Info("management-console: listening", "addr", addr)
	log.Fatalf("Fatal error serving: %s", http.ListenAndServe(addr, nil))
}
