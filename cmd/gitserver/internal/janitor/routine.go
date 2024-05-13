package janitor

import (
	"context"
	"time"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/janitor/stats"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// New creates a new janitor. This is a temporary routine that lives in gitserver
// that somewhat mimics the behavior of the old janitor by simply iterating over
// all the repositories and optimizing them. If we build coordinator, the per-repo
// optimization will be invoked by the coordinator via gRPC instead, and it will
// manage the queue of repos to optimize.
func New(logger log.Logger, concurrency int, fs gitserverfs.FS, getBackendFunc func(dir common.GitDir, repoName api.RepoName) git.GitBackend) *janitor {
	if concurrency < 1 {
		panic("concurrency must be at least 1")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &janitor{
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger.Scoped("janitor"),
		fs:             fs,
		getBackendFunc: getBackendFunc,
		concurrency:    concurrency,
		stopped:        make(chan struct{}),
	}
}

var _ goroutine.BackgroundRoutine = (*janitor)(nil)

type janitor struct {
	logger         log.Logger
	fs             gitserverfs.FS
	ctx            context.Context
	cancel         context.CancelFunc
	getBackendFunc func(dir common.GitDir, repoName api.RepoName) git.GitBackend
	concurrency    int
	stopped        chan struct{}
}

func (j *janitor) Start() {
	for {
		start := time.Now()
		p := pool.New().WithMaxGoroutines(j.concurrency)
		err := j.fs.ForEachRepo(func(name api.RepoName, dir common.GitDir) (done bool) {
			p.Go(func() {
				backend := j.getBackendFunc(dir, name)

				logger := j.logger.With(log.String("repo", string(name)))

				logger.Info("optimizing repository")

				ctx, cancel := context.WithTimeout(j.ctx, 2*time.Hour)
				defer cancel()

				// We always attempt to bring the repo into a good initial state.
				err := repairRepo(ctx, logger, backend, dir)
				if err != nil {
					logger.Error("failed to repair repository", log.String("repo", string(name)), log.Error(err))
					return
				}

				s, err := stats.RepositoryInfoForRepository(ctx, dir)
				if err != nil {
					logger.Error("failed to get repository stats, won't attempt to optimize repository", log.String("repo", string(name)), log.Error(err))
					return
				}

				now := time.Now()

				shouldRepack, repackCfg := shouldRepackObjects(logger, s, now)
				if shouldRepack {
					logger.Info("will repack objects")
					if err := RepackObjects(ctx, backend, dir, repackCfg); err != nil {
						logger.Error("failed to repack objects", log.String("repo", string(name)), log.Error(err))
						return
					}
				}

				shouldPrune, expiry := shouldPruneObjects(s, now)
				if shouldPrune {
					logger.Info("will prune objects")
					if err := backend.Maintenance().PruneObjects(ctx, expiry); err != nil {
						logger.Error("failed to prune objects", log.String("repo", string(name)), log.Error(err))
						return
					}
				}

				if shouldRepackReferences(s) {
					logger.Info("will repack references")
					if err := backend.Maintenance().PackRefs(ctx); err != nil {
						logger.Error("failed to pack refs", log.String("repo", string(name)), log.Error(err))
						return
					}
				}

				shouldWrite, replaceChain := shouldWriteCommitGraph(s, now)
				if shouldWrite {
					logger.Info("will write commit graph")
					err := backend.Maintenance().WriteCommitGraph(ctx, replaceChain)
					if err != nil {
						logger.Error("failed to write commit graph", log.String("repo", string(name)), log.Error(err))
						return
					}
				}

				// TODO: If we did any of the above, we might want to recalculate the repo size
				// and store the new size in the DB.

				return
			})

			return false
		})
		p.Wait()
		if err != nil {
			j.logger.Error("failed to iterate over repositories", log.Error(err))
		}

		j.logger.Info("Janitor run finished", log.Duration("duration", time.Since(start)))
		time.Sleep(1 * time.Minute)
	}
}

func (j *janitor) Stop() {
	close(j.stopped)
	j.cancel()
}
