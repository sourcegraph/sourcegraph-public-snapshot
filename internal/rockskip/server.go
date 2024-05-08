package rockskip

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/go-ctags"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Symbol struct {
	Name   string `json:"name"`
	Parent string `json:"parent"`
	Kind   string `json:"kind"`
	Line   int    `json:"line"`
}

const NULL CommitId = 0

type Service struct {
	logger                  log.Logger
	metrics                 *metrics
	db                      *sql.DB
	git                     GitserverClient
	fetcher                 fetcher.RepositoryFetcher
	createParser            func() (ctags.Parser, error)
	status                  *ServiceStatus
	repoUpdates             chan struct{}
	maxRepos                int
	logQueries              bool
	repoCommitToDone        map[string]chan struct{}
	repoCommitToDoneMu      sync.Mutex
	indexRequestQueues      []chan indexRequest
	symbolsCacheSize        int
	pathSymbolsCacheSize    int
	searchLastIndexedCommit bool
}

type metrics struct {
	searchRunning  prometheus.Gauge
	searchFailed   prometheus.Counter
	searchDuration prometheus.Histogram
	indexRunning   prometheus.Gauge
	indexFailed    prometheus.Counter
	indexDuration  prometheus.Histogram
	queueAge       prometheus.Histogram
}

func newMetrics(logger log.Logger, db *sql.DB) *metrics {
	scanCount := func(sql string) (float64, error) {
		row := db.QueryRowContext(context.Background(), sql)
		var count int64
		err := row.Scan(&count)
		if err != nil {
			return 0, err
		}
		return float64(count), nil
	}

	ns := "src_rockskip_service"

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: ns,
		Name:      "repos_indexed",
		Help:      "The number of repositories indexed by rockskip",
	}, func() float64 {
		count, err := scanCount(`SELECT COUNT(*) FROM rockskip_repos`)
		if err != nil {
			logger.Error("failed to get number of index repos", log.Error(err))
			return 0
		}
		return count
	})
	return &metrics{
		searchRunning: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "in_flight_search_requests",
			Help:      "Number of in-flight search requests",
		}),
		searchFailed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      "search_request_errors",
			Help:      "Number of search requests that returned an error",
		}),
		searchDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Name:      "search_request_duration_seconds",
			Help:      "Search request duration in seconds.",
			Buckets:   []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 15, 20, 30},
		}),
		indexRunning: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "in_flight_index_jobs",
			Help:      "Number of in-flight index jobs",
		}),
		indexFailed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      "index_job_errors",
			Help:      "Number of index jobs that returned an error",
		}),
		indexDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Name:      "index_job_duration_seconds",
			Help:      "Search request duration in seconds.",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 22),
		}),
		queueAge: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Name:      "index_queue_age_seconds",
			Help:      "A histogram of the amount of time a popped index request spent sitting in the queue beforehand.",
			Buckets: []float64{
				60,     // 1m
				300,    // 5m
				1200,   // 20m
				2400,   // 40m
				3600,   // 1h
				10800,  // 3h
				18000,  // 5h
				36000,  // 10h
				43200,  // 12h
				54000,  // 15h
				72000,  // 20h
				86400,  // 24h
				108000, // 30h
				126000, // 35h
				172800, // 48h
			},
		}),
	}
}

func NewService(
	db *sql.DB,
	git GitserverClient,
	fetcher fetcher.RepositoryFetcher,
	createParser func() (ctags.Parser, error),
	maxConcurrentlyIndexing int,
	maxRepos int,
	logQueries bool,
	indexRequestsQueueSize int,
	symbolsCacheSize int,
	pathSymbolsCacheSize int,
	searchLastIndexedCommit bool,
) (*Service, error) {
	indexRequestQueues := make([]chan indexRequest, maxConcurrentlyIndexing)
	for i := range maxConcurrentlyIndexing {
		indexRequestQueues[i] = make(chan indexRequest, indexRequestsQueueSize)
	}

	logger := log.Scoped("service")

	service := &Service{
		logger:                  logger,
		metrics:                 newMetrics(logger, db),
		db:                      db,
		git:                     git,
		fetcher:                 fetcher,
		createParser:            createParser,
		status:                  NewStatus(logger),
		repoUpdates:             make(chan struct{}, 1),
		maxRepos:                maxRepos,
		logQueries:              logQueries,
		repoCommitToDone:        map[string]chan struct{}{},
		repoCommitToDoneMu:      sync.Mutex{},
		indexRequestQueues:      indexRequestQueues,
		symbolsCacheSize:        symbolsCacheSize,
		pathSymbolsCacheSize:    pathSymbolsCacheSize,
		searchLastIndexedCommit: searchLastIndexedCommit,
	}

	go service.startCleanupLoop()

	for i := range maxConcurrentlyIndexing {
		go service.startIndexingLoop(service.indexRequestQueues[i])
	}

	return service, nil
}

func (s *Service) startIndexingLoop(indexRequestQueue chan indexRequest) {
	// We should use an internal actor when doing cross service calls.
	ctx := actor.WithInternalActor(context.Background())
	for indexRequest := range indexRequestQueue {
		s.metrics.queueAge.Observe(time.Since(indexRequest.dateAddedToQueue).Seconds())
		err := s.Index(ctx, indexRequest.repo, indexRequest.commit)
		close(indexRequest.done)
		if err != nil {
			s.logger.Error("indexing error",
				log.String("repo", indexRequest.repo),
				log.String("commit", indexRequest.commit),
				log.Error(err),
			)
		}
	}
}

func (s *Service) startCleanupLoop() {
	for range s.repoUpdates {
		threadStatus := s.status.NewThreadStatus("cleanup")
		err := DeleteOldRepos(context.Background(), s.db, s.maxRepos, threadStatus)
		threadStatus.End()
		if err != nil {
			s.logger.Error("failed to delete old repos", log.Error(err))
		}
	}
}

func getHops(ctx context.Context, tx dbutil.DB, commit int, tasklog *TaskLog) ([]int, error) {
	tasklog.Start("get hops")

	current := commit
	spine := []int{current}

	for {
		_, ancestor, _, present, err := GetCommitById(ctx, tx, current)
		if err != nil {
			return nil, errors.Wrap(err, "GetCommitById")
		} else if !present {
			break
		} else {
			if current == NULL {
				break
			}
			current = ancestor
			spine = append(spine, current)
		}
	}

	return spine, nil
}

func DeleteOldRepos(ctx context.Context, db *sql.DB, maxRepos int, threadStatus *ThreadStatus) error {
	// Get a fresh connection from the DB pool to get deterministic "lock stacking" behavior.
	// See doc/dev/background-information/sql/locking_behavior.md for more details.
	conn, err := db.Conn(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to get connection for deleting old repos")
	}
	defer conn.Close()

	// Keep deleting repos until we're back to at most maxRepos.
	for {
		more, err := tryDeleteOldestRepo(ctx, conn, maxRepos, threadStatus)
		if err != nil {
			return err
		}
		if !more {
			return nil
		}
	}
}

// Ruler sequence
//
// input : 0, 1, 2, 3, 4, 5, 6, 7, 8, ...
// output: 0, 0, 1, 0, 2, 0, 1, 0, 3, ...
//
// https://oeis.org/A007814
func ruler(n int) int {
	height := 0
	for n > 0 && n%2 == 0 {
		height++
		n = n / 2
	}
	return height
}
