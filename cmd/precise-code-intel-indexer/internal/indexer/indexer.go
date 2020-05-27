package indexer

import (
	"context"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
)

type Indexer struct {
	db           db.DB
	processor    Processor
	frontendURL  string
	pollInterval time.Duration
	metrics      IndexerMetrics
	done         chan struct{}
	once         sync.Once
}

func NewIndexer(
	db db.DB,
	gitserverClient gitserver.Client,
	frontendURL string,
	pollInterval time.Duration,
	metrics IndexerMetrics,
) *Indexer {
	processor := &processor{
		db:              db,
		gitserverClient: gitserverClient,
		frontendURL:     frontendURL,
	}

	return &Indexer{
		db:           db,
		processor:    processor,
		frontendURL:  frontendURL,
		pollInterval: pollInterval,
		metrics:      metrics,
		done:         make(chan struct{}),
	}
}

func (i *Indexer) Start() {
	for {
		if ok, _ := i.dequeueAndProcess(context.Background()); !ok {
			select {
			case <-time.After(i.pollInterval):
			case <-i.done:
				return
			}
		} else {
			select {
			case <-i.done:
				return
			default:
			}
		}
	}
}

func (i *Indexer) Stop() {
	i.once.Do(func() {
		close(i.done)
	})
}

func (i *Indexer) dequeueAndProcess(ctx context.Context) (_ bool, err error) {
	start := time.Now()

	index, tx, ok, err := i.db.DequeueIndex(ctx)
	if err != nil || !ok {
		return false, errors.Wrap(err, "db.DequeueIndex")
	}
	defer func() {
		err = tx.Done(err)

		// TODO(efritz) - set error if indexing failed
		i.metrics.Processor.Observe(time.Since(start).Seconds(), 1, &err)
	}()

	log15.Info(
		"Dequeued index for processing",
		"id", index.ID,
		"repository_id", index.RepositoryID,
		"commit", index.Commit,
	)

	if processErr := i.processor.Process(ctx, index); processErr == nil {
		log15.Info(
			"Indexed repository",
			"id", index.ID,
			"repository_id", index.RepositoryID,
			"commit", index.Commit,
		)

		if markErr := tx.MarkIndexComplete(ctx, index.ID); markErr != nil {
			return true, errors.Wrap(markErr, "db.MarkIndexComplete")
		}
	} else {
		// TODO(efritz) - distinguish between index and system errors
		log15.Warn(
			"Failed to index repository",
			"id", index.ID,
			"repository_id", index.RepositoryID,
			"commit", index.Commit,
		)

		if markErr := tx.MarkIndexErrored(ctx, index.ID, processErr.Error(), ""); markErr != nil {
			return true, errors.Wrap(markErr, "db.MarkIndexErrored")
		}
	}

	return true, nil
}
