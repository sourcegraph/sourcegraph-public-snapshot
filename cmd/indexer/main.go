package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/indexer/idx"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"
)

var numWorkers = env.Get("NUM_WORKERS", "4", "The maximum number of indexing done in parallel.")
var profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")
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
	gitserver.DefaultClient.NoCreds = true
	if err := idx.Google.SetAPIKey(googleAPIKey); err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize Google API client: %s\n", err)
		os.Exit(1)
	}

	// SECURITY: This is only safe because the indexer runs in isolation and does not expose any data to the outside world
	ctx := accesscontrol.WithInsecureSkip(context.Background(), true)
	// Prefer raw git over GitHub API
	github.PreferRawGit = true

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
		<-c
		os.Exit(0)
	}()

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
		log15.Info(fmt.Sprintf("Profiler available on %s/pprof", profBindAddr))
	}

	wq := idx.NewQueue(queueLength)
	n, _ := strconv.Atoi(numWorkers)
	for i := 0; i < n; i++ {
		go idx.Work(ctx, wq)
	}

	http.HandleFunc("/refresh", func(resp http.ResponseWriter, req *http.Request) {
		repo := req.URL.Query().Get("repo")
		rev := req.URL.Query().Get("rev")
		if repo == "" {
			http.Error(resp, "missing repo parameter", http.StatusBadRequest)
			return
		}
		wq.Enqueue(repo, rev)
		resp.Write([]byte("OK"))
	})

	fmt.Println("indexer: listening on :3179")
	log.Fatal(http.ListenAndServe(":3179", nil))
}
