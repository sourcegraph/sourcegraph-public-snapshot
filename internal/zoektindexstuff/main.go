// Command zoekt-sourcegraph-indexserver periodically reindexes enabled
// repositories on sourcegraph
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/zoekt"
	"github.com/google/zoekt/build"
	wipindexserver "github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver/wip"
	"github.com/google/zoekt/debugserver"

	"cloud.google.com/go/profiler"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/keegancsmith/tmpfriend"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/automaxprocs/maxprocs"
)

var (
	metricResolveRevisionsDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "resolve_revisions_seconds",
		Help:    "A histogram of latencies for resolving all repository revisions.",
		Buckets: prometheus.ExponentialBuckets(1, 10, 6), // 1s -> 27min
	})

	metricResolveRevisionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "resolve_revision_seconds",
		Help:    "A histogram of latencies for resolving a repository revision.",
		Buckets: prometheus.ExponentialBuckets(.25, 2, 4), // 250ms -> 2s
	}, []string{"success"}) // success=true|false

	metricGetIndexOptionsError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_index_options_error_total",
		Help: "The total number of times we failed to get index options for a repository.",
	})

	metricIndexDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "index_repo_seconds",
		Help:    "A histogram of latencies for indexing a repository.",
		Buckets: prometheus.ExponentialBuckets(.1, 10, 7), // 100ms -> 27min
	}, []string{
		"state", // state is an indexState
		"name",  // name of the repository that was indexed
	})

	metricFetchDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "index_fetch_seconds",
		Help:    "A histogram of latencies for fetching a repository.",
		Buckets: []float64{.05, .1, .25, .5, 1, 2.5, 5, 10, 20, 30, 60, 180, 300, 600}, // 50ms -> 10 minutes
	}, []string{
		"success", // true|false
		"name",    // the name of the repository that the commits were fetched from
	})

	metricIndexIncrementalIndexState = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "index_incremental_index_state",
		Help: "A count of the state on disk vs what we want to build. See zoekt/build.IndexState.",
	}, []string{"state"}) // state is build.IndexState

	metricNumIndexed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "index_num_indexed",
		Help: "Number of indexed repos by code host",
	})

	metricNumAssigned = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "index_num_assigned",
		Help: "Number of repos assigned to this indexer by code host",
	})

	metricFailingTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "index_failing_total",
		Help: "Counts failures to index (indexing activity, should be used with rate())",
	})

	metricIndexingTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "index_indexing_total",
		Help: "Counts indexings (indexing activity, should be used with rate())",
	})
)

// set of repositories that we want to capture separate indexing metrics for
var reposWithSeparateIndexingMetrics = make(map[string]struct{})

var debug = log.New(ioutil.Discard, "", log.LstdFlags)

func batched(slice []uint32, size int) <-chan []uint32 {
	c := make(chan []uint32)
	go func() {
		for len(slice) > 0 {
			if size > len(slice) {
				size = len(slice)
			}
			c <- slice[:size]
			slice = slice[size:]
		}
		close(c)
	}()
	return c
}

func listIndexed(indexDir string) []uint32 {
	index := wipindexserver.GetShards(indexDir)
	metricNumIndexed.Set(float64(len(index)))
	repoIDs := make([]uint32, 0, len(index))
	for id := range index {
		repoIDs = append(repoIDs, id)
	}
	sort.Slice(repoIDs, func(i, j int) bool {
		return repoIDs[i] < repoIDs[j]
	})
	return repoIDs
}

func hostnameBestEffort() string {
	if h := os.Getenv("NODE_NAME"); h != "" {
		return h
	}
	if h := os.Getenv("HOSTNAME"); h != "" {
		return h
	}
	hostname, _ := os.Hostname()
	return hostname
}

// setupTmpDir sets up a temporary directory on the same volume as the
// indexes.
//
// If main is true we will delete older temp directories left around. main is
// false when this is a debug command.
func setupTmpDir(index string, main bool) error {
	tmpRoot := filepath.Join(index, ".indexserver.tmp")
	if err := os.MkdirAll(tmpRoot, 0755); err != nil {
		return err
	}
	if !tmpfriend.IsTmpFriendDir(tmpRoot) {
		_, err := tmpfriend.RootTempDir(tmpRoot)
		return err
	}
	return nil
}

func printMetaData(fn string) error {
	repo, indexMeta, err := zoekt.ReadMetadataPath(fn)
	if err != nil {
		return err
	}

	err = json.NewEncoder(os.Stdout).Encode(indexMeta)
	if err != nil {
		return err
	}

	err = json.NewEncoder(os.Stdout).Encode(repo)
	if err != nil {
		return err
	}
	return nil
}

func printShardStats(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}

	iFile, err := zoekt.NewIndexFile(f)
	if err != nil {
		return err
	}

	return zoekt.PrintNgramStats(iFile)
}

func initializeGoogleCloudProfiler() {
	// Google cloud profiler is opt-in since we only want to run it on
	// Sourcegraph.com.
	if os.Getenv("GOOGLE_CLOUD_PROFILER_ENABLED") == "" {
		return
	}

	err := profiler.Start(profiler.Config{
		Service:        "zoekt-sourcegraph-indexserver",
		ServiceVersion: zoekt.Version,
		MutexProfiling: true,
		AllocForceGC:   true,
	})
	if err != nil {
		log.Printf("could not initialize google cloud profiler: %s", err.Error())
	}
}

func srcLogLevelIsDebug() bool {
	lvl := os.Getenv("SRC_LOG_LEVEL")
	return strings.EqualFold(lvl, "dbug") || strings.EqualFold(lvl, "debug")
}

func getEnvWithDefaultInt64(k string, defaultVal int64) int64 {
	v := os.Getenv(k)
	if v == "" {
		return defaultVal
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Fatalf("error parsing ENV %s: %s", k, err)
	}
	return i
}

func setCompoundShardCounter(indexDir string) {
	fns, err := filepath.Glob(filepath.Join(indexDir, "compound-*.zoekt"))
	if err != nil {
		log.Printf("setCompoundShardCounter: %s\n", err)
		return
	}
	// TODO
	_ = fns
	// metricNumberCompoundShards.Set(float64(len(fns)))
}

func main() {
	defaultIndexDir := os.Getenv("DATA_DIR")
	if defaultIndexDir == "" {
		defaultIndexDir = build.DefaultDir
	}

	root := flag.String("sourcegraph_url", os.Getenv("SRC_FRONTEND_INTERNAL"), "http://sourcegraph-frontend-internal or http://localhost:3090. If a path to a directory, we fake the Sourcegraph API and index all repos rooted under path.")
	interval := flag.Duration("interval", time.Minute, "sync with sourcegraph this often")
	vacuumInterval := flag.Duration("vacuum_interval", time.Hour, "run vacuum this often")
	mergeInterval := flag.Duration("merge_interval", time.Hour, "run merge this often")
	targetSize := flag.Int64("merge_target_size", getEnvWithDefaultInt64("SRC_TARGET_SIZE", 2000), "the target size of compound shards in MiB")
	maxSize := flag.Int64("merge_max_size", getEnvWithDefaultInt64("SRC_MAX_SIZE", 1800), "the maximum size in MiB a shard can have to be considered for merging")
	minSize := flag.Int64("merge_min_size", getEnvWithDefaultInt64("SRC_MIN_SIZE", 1800), "the minimum size of a compound shard in MiB")
	index := flag.String("index", defaultIndexDir, "set index directory to use")
	listen := flag.String("listen", ":6072", "listen on this address.")
	hostname := flag.String("hostname", hostnameBestEffort(), "the name we advertise to Sourcegraph when asking for the list of repositories to index. Can also be set via the NODE_NAME environment variable.")
	cpuFraction := flag.Float64("cpu_fraction", 1.0, "use this fraction of the cores for indexing.")
	dbg := flag.Bool("debug", srcLogLevelIsDebug(), "turn on more verbose logging.")

	// non daemon mode for debugging/testing
	debugList := flag.Bool("debug-list", false, "do not start the indexserver, rather list the repositories owned by this indexserver then quit.")
	debugIndex := flag.String("debug-index", "", "do not start the indexserver, rather index the repository ID then quit.")
	debugShard := flag.String("debug-shard", "", "do not start the indexserver, rather print shard stats then quit.")
	debugMeta := flag.String("debug-meta", "", "do not start the indexserver, rather print shard metadata then quit.")
	debugMerge := flag.Bool("debug-merge", false, "do not start the indexserver, rather run merge in the index directory then quit.")
	debugMergeSimulate := flag.Bool("simulate", false, "use in conjuction with debugMerge. If set, merging is simulated.")

	_ = flag.Bool("exp-git-index", true, "DEPRECATED: not read anymore. We always use zoekt-git-index now.")

	flag.Parse()

	if *cpuFraction <= 0.0 || *cpuFraction > 1.0 {
		log.Fatal("cpu_fraction must be between 0.0 and 1.0")
	}
	if *index == "" {
		log.Fatal("must set -index")
	}
	needSourcegraph := !(*debugShard != "" || *debugMeta != "" || *debugMerge)
	if *root == "" && needSourcegraph {
		log.Fatal("must set -sourcegraph_url")
	}
	rootURL, err := url.Parse(*root)
	if err != nil {
		log.Fatalf("url.Parse(%v): %v", *root, err)
	}

	// Tune GOMAXPROCS to match Linux container CPU quota.
	_, _ = maxprocs.Set()

	// Automatically prepend our own path at the front, to minimize
	// required configuration.
	if l, err := os.Readlink("/proc/self/exe"); err == nil {
		os.Setenv("PATH", filepath.Dir(l)+":"+os.Getenv("PATH"))
	}

	if _, err := os.Stat(*index); err != nil {
		if err := os.MkdirAll(*index, 0755); err != nil {
			log.Fatalf("MkdirAll %s: %v", *index, err)
		}
	}

	isDebugCmd := *debugList || *debugIndex != "" || *debugShard != "" || *debugMeta != "" || *debugMerge

	if err := setupTmpDir(*index, !isDebugCmd); err != nil {
		log.Fatalf("failed to setup TMPDIR under %s: %v", *index, err)
	}

	if *dbg || isDebugCmd {
		debug = log.New(os.Stderr, "", log.LstdFlags)
	}

	indexingMetricsReposAllowlist := os.Getenv("INDEXING_METRICS_REPOS_ALLOWLIST")
	if indexingMetricsReposAllowlist != "" {
		var repos []string

		for _, r := range strings.Split(indexingMetricsReposAllowlist, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				repos = append(repos, r)
			}
		}

		for _, r := range repos {
			reposWithSeparateIndexingMetrics[r] = struct{}{}
		}

		debug.Printf("capturing separate indexing metrics for: %s", repos)
	}

	var sg wipindexserver.Sourcegraph
	if rootURL.IsAbs() {
		var batchSize int
		if v := os.Getenv("SRC_REPO_CONFIG_BATCH_SIZE"); v != "" {
			batchSize, err = strconv.Atoi(v)
			if err != nil {
				log.Fatal("Invalid value for SRC_REPO_CONFIG_BATCH_SIZE, must be int")
			}
		}

		client := retryablehttp.NewClient()
		client.Logger = debug
		sg = &sourcegraphClient{
			Root:      rootURL,
			Client:    client,
			Hostname:  *hostname,
			BatchSize: batchSize,
		}
	} else {
		sg = sourcegraphFake{
			RootDir: rootURL.String(),
			Log:     log.New(os.Stderr, "sourcegraph: ", log.LstdFlags),
		}
	}

	cpuCount := int(math.Round(float64(runtime.GOMAXPROCS(0)) * (*cpuFraction)))
	if cpuCount < 1 {
		cpuCount = 1
	}
	s := &wipindexserver.Server{
		Sourcegraph:     sg,
		IndexDir:        *index,
		Interval:        *interval,
		VacuumInterval:  *vacuumInterval,
		MergeInterval:   *mergeInterval,
		CPUCount:        cpuCount,
		TargetSizeBytes: *targetSize * 1024 * 1024,
		MaxSizeBytes:    *maxSize * 1024 * 1024,
		MinSizeBytes:    *minSize * 1024 * 1024,
		ShardMerging:    zoekt.ShardMergingEnabled(),
	}

	if *debugList {
		repos, err := s.Sourcegraph.List(context.Background(), listIndexed(s.IndexDir))
		if err != nil {
			log.Fatal(err)
		}
		for _, r := range repos.IDs {
			fmt.Println(r)
		}
		os.Exit(0)
	}

	if *debugIndex != "" {
		id, err := strconv.Atoi(*debugIndex)
		if err != nil {
			log.Fatal(err)
		}
		msg, err := s.ForceIndex(uint32(id))
		log.Println(msg)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *debugShard != "" {
		err = printShardStats(*debugShard)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if *debugMeta != "" {
		err = printMetaData(*debugMeta)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if *debugMerge {
		err = wipindexserver.DoMerge(*index, *targetSize*1024*1024, *maxSize*1024*1024, *debugMergeSimulate)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	initializeGoogleCloudProfiler()
	setCompoundShardCounter(s.IndexDir)

	if *listen != "" {
		go func() {
			mux := http.NewServeMux()
			debugserver.AddHandlers(mux, true)
			mux.Handle("/", s)
			debug.Printf("serving HTTP on %s", *listen)
			log.Fatal(http.ListenAndServe(*listen, mux))
		}()
	}

	s.Run()
}
