//docker:user sourcegraph

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/indexer/idx"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
)

var numWorkers = env.Get("NUM_WORKERS", "4", "The maximum number of indexing done in parallel.")
var googleAPIKey = env.Get("GOOGLE_CSE_API_TOKEN", "", "API token for issuing Google:github.com search queries")

var queueLength = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "indexer",
	Name:      "queue_length",
	Help:      "Lengh of the indexer's queue of repos to check/index.",
})

func init() {
	prometheus.MustRegister(queueLength)
}

func main() {
	env.Lock()
	env.HandleHelpFlag()

	tracer.Init()

	if googleAPIKey != "" {
		if err := idx.Google.SetAPIKey(googleAPIKey); err != nil {
			log.Println("Could not initialize Google API client: ", err)
			os.Exit(1)
		}
	}

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

	log15.Info("indexer: listening", "addr", ":3179")
	log.Fatalf("Fatal error serving: %s", http.ListenAndServe(":3179", nil))
}
