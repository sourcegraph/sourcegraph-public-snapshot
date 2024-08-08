package rockskip

import (
	"context"
	"database/sql"
	"math/bits"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	symbolparser "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/fetcher"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
	symbolParserPool        *symbolparser.ParserPool
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

func NewService(
	observationCtx *observation.Context,
	db *sql.DB,
	git GitserverClient,
	fetcher fetcher.RepositoryFetcher,
	symbolParserPool *symbolparser.ParserPool,
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

	service := &Service{
		logger:                  observationCtx.Logger,
		metrics:                 newMetrics(observationCtx, db),
		db:                      db,
		git:                     git,
		fetcher:                 fetcher,
		symbolParserPool:        symbolParserPool,
		status:                  NewStatus(),
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

func getHops(ctx context.Context, tx dbutil.DB, commit int, tasklog *TaskLog) ([]int, error) {
	tasklog.Start("get hops")

	current := commit
	spine := []int{current}

	for {
		_, ancestor, _, present, err := GetCommitById(ctx, tx, current)
		if err != nil {
			return nil, errors.Wrap(err, "GetCommitById")
		}

		if !present || current == NULL {
			return spine, nil
		}

		current = ancestor
		spine = append(spine, current)
	}
}

// Ruler sequence
//
// input : 0, 1, 2, 3, 4, 5, 6, 7, 8, ...
// output: 0, 0, 1, 0, 2, 0, 1, 0, 3, ...
//
// https://oeis.org/A007814
func ruler(n int) int {
	if n <= 0 {
		return 0
	}
	// ruler(n) is equivalent to asking how many times can you divide n by 2
	// before you get an odd number. That is the number of 0's at the end of n
	// when n is written in base 2.
	return bits.TrailingZeros(uint(n))
}
