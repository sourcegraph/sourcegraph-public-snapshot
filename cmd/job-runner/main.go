// Command job-runner runs arbitrary/third-party jobs in a containerized (but)
// not entirely sandboxed) fashion.
//
// Note: job-runner does not import the standard Sourcegraph debugserver or tracing
// packages, as this service is intended to be as isolated as possible for hightened
// security and they depend on the frontend internal API.
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sourcegraph/sourcegraph/cmd/job-runner/runner"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

const port = "3190"

func main() {
	env.Lock()
	env.HandleHelpFlag()
	log.SetFlags(0)

	service := &runner.Service{
		JobFinish: func() {
			log.Fatal("job finished: exiting to restart ephemeral container (this is normal)")
		},
		Log: log15.Root(),
	}

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)

	// We rely on the liveness probe not responding in the event a job has gone
	// rogue and consumed all available container resources.
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
		return
	})
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", ot.Middleware(service))
	server := &http.Server{Addr: addr, Handler: mux}
	go shutdownOnSIGINT(server)

	log15.Info("job-runner: waiting for a job on", "addr", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func shutdownOnSIGINT(s *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.Shutdown(ctx)
	if err != nil {
		log.Fatal("graceful server shutdown failed, will exit:", err)
	}
}
