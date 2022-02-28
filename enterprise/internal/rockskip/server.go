package rockskip

import (
	"context"
	"database/sql"
	"sync"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Symbol struct {
	Name   string `json:"name"`
	Parent string `json:"parent"`
	Kind   string `json:"kind"`
	Line   int    `json:"line"`
}

type ParseSymbolsFunc func(path string, bytes []byte) (symbols []Symbol, err error)

const NULL = 0

type Server struct {
	db                   *sql.DB
	git                  Git
	createParser         func() ParseSymbolsFunc
	status               *ServerStatus
	repoUpdates          chan struct{}
	maxRepos             int
	logQueries           bool
	repoCommitToDone     map[string]chan struct{}
	repoCommitToDoneMu   sync.Mutex
	indexRequestQueues   []chan indexRequest
	symbolsCacheSize     int
	pathSymbolsCacheSize int
}

func NewServer(
	db *sql.DB,
	git Git,
	createParser func() ParseSymbolsFunc,
	maxConcurrentlyIndexing int,
	maxRepos int,
	logQueries bool,
	indexRequestsQueueSize int,
	symbolsCacheSize int,
	pathSymbolsCacheSize int,
) (*Server, error) {
	indexRequestQueues := make([]chan indexRequest, maxConcurrentlyIndexing)
	for i := 0; i < maxConcurrentlyIndexing; i++ {
		indexRequestQueues[i] = make(chan indexRequest, indexRequestsQueueSize)
	}

	server := &Server{
		db:                   db,
		git:                  git,
		createParser:         createParser,
		status:               NewStatus(),
		repoUpdates:          make(chan struct{}, 1),
		maxRepos:             maxRepos,
		logQueries:           logQueries,
		repoCommitToDone:     map[string]chan struct{}{},
		repoCommitToDoneMu:   sync.Mutex{},
		indexRequestQueues:   indexRequestQueues,
		symbolsCacheSize:     symbolsCacheSize,
		pathSymbolsCacheSize: pathSymbolsCacheSize,
	}

	err := server.startCleanupThread()
	if err != nil {
		return nil, err
	}

	for i := 0; i < maxConcurrentlyIndexing; i++ {
		err = server.startIndexingThread(server.indexRequestQueues[i])
		if err != nil {
			return nil, err
		}
	}

	return server, nil
}

func (s *Server) startIndexingThread(indexRequestQueue chan indexRequest) (err error) {
	go func() {
		for indexRequest := range indexRequestQueue {
			// Get a fresh connection from the DB pool to get deterministic "lock stacking" behavior.
			// https://www.postgresql.org/docs/9.1/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
			conn, err := s.db.Conn(context.Background())
			if err != nil {
				log15.Error("Failed to get connection for indexing thread", "error", err)
				continue
			}

			err = s.Index(context.Background(), conn, indexRequest.repo, indexRequest.commit, s.createParser())
			close(indexRequest.done)
			if err != nil {
				log15.Error("indexing error", "repo", indexRequest.repo, "commit", indexRequest.commit, "err", err)
			}

			conn.Close()
		}
	}()

	return nil
}

func (s *Server) startCleanupThread() error {
	go func() {
		for range s.repoUpdates {
			// Get a fresh connection from the DB pool to get deterministic "lock stacking" behavior.
			// https://www.postgresql.org/docs/9.1/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
			conn, err := s.db.Conn(context.Background())
			if err != nil {
				log15.Error("Failed to get connection for deleting old repos", "error", err)
				continue
			}

			threadStatus := s.status.NewThreadStatus("cleanup")
			err = DeleteOldRepos(context.Background(), conn, s.maxRepos, threadStatus)
			threadStatus.End()
			if err != nil {
				log15.Error("Failed to delete old repos", "error", err)
			}

			conn.Close()
		}
	}()

	return nil
}

func getHops(ctx context.Context, tx Queryable, commit int, tasklog *TaskLog) ([]int, error) {
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

func DeleteOldRepos(ctx context.Context, db *sql.Conn, maxRepos int, threadStatus *ThreadStatus) error {
	// Keep deleting repos until we're back to at most maxRepos.
	for {
		more, err := tryDeleteOldestRepo(ctx, db, maxRepos, threadStatus)
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
	if n == 0 {
		return 0
	}
	if n%2 != 0 {
		return 0
	}
	return 1 + ruler(n/2)
}

// combineErrors is like errors.Append, but it returns nil when there are no errors or all errors are
// nil so that success is propagated.
func combineErrors(errOrNils ...error) error {
	errs := []error{}
	for _, err := range errOrNils {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return errors.Append(errs[0], errs[1:]...)
}
