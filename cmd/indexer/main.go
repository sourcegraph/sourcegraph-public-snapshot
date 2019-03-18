//docker:user sourcegraph

package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/indexer/idx"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
)

var numWorkers = env.Get("NUM_WORKERS", "4", "The maximum number of indexing done in parallel.")

var queueLength = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "indexer",
	Name:      "queue_length",
	Help:      "Length of the indexer's queue of repos to check/index.",
})

func init() {
	prometheus.MustRegister(queueLength)
}

const port = "3179"

func main() {
	env.Lock()
	env.HandleHelpFlag()

	tracer.Init()

	ctx := context.Background()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
		<-c
		os.Exit(0)
	}()

	go debugserver.Start()

	// Secondary queue relies on access to frontend, so wait until it has started up.
	api.WaitForFrontend(ctx)

	worker := idx.NewWorker(ctx, idx.NewQueue(queueLength), idx.SecondaryQueue(ctx))
	n, _ := strconv.Atoi(numWorkers)
	for i := 0; i < n; i++ {
		go worker.Work()
	}

	http.HandleFunc("/refresh", func(resp http.ResponseWriter, req *http.Request) {
		repo := api.RepoURI(req.URL.Query().Get("repo"))
		rev := req.URL.Query().Get("rev")
		if repo == "" {
			http.Error(resp, "missing repo parameter", http.StatusBadRequest)
			return
		}
		worker.Enqueue(repo, rev)
		resp.Write([]byte("OK"))
	})

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	log15.Info("indexer: listening", "addr", addr)
	log.Fatalf("Fatal error serving: %s", http.ListenAndServe(addr, nil))
}
