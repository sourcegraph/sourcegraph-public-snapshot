package internal

import (
	"bufio"
	"bytes"
	"container/list"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/perforce"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/urlredactor"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/vcssyncer"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	repoClonedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_repo_cloned",
		Help: "number of successful git clones run",
	})
	repoCloneFailedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_repo_cloned_failed",
		Help: "number of failed git clones",
	})
)

// NewClonePipeline creates a new pipeline that clones repos asynchronously. It
// creates a producer-consumer pipeline that handles clone requests asychronously.
func (s *Server) NewClonePipeline(logger log.Logger, cloneQueue *common.Queue[*cloneJob]) goroutine.BackgroundRoutine {
	return &clonePipelineRoutine{
		tasks:  make(chan *cloneTask),
		logger: logger,
		s:      s,
		queue:  cloneQueue,
	}
}

type clonePipelineRoutine struct {
	logger log.Logger

	tasks chan *cloneTask
	// TODO: Get rid of this dependency.
	s      *Server
	queue  *common.Queue[*cloneJob]
	cancel context.CancelFunc
}

func (p *clonePipelineRoutine) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	// Start a go routine for each the producer and the consumer.
	go p.cloneJobConsumer(ctx, p.tasks)
	go p.cloneJobProducer(ctx, p.tasks)
}

func (p *clonePipelineRoutine) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *clonePipelineRoutine) cloneJobProducer(ctx context.Context, tasks chan<- *cloneTask) {
	defer close(tasks)

	for {
		// Acquire the cond mutex lock and wait for a signal if the queue is empty.
		p.queue.Mutex.Lock()
		if p.queue.Empty() {
			// TODO: This should only wait if ctx is not canceled.
			p.queue.Cond.Wait()
		}

		// The queue is not empty and we have a job to process! But don't forget to unlock the cond
		// mutex here as we don't need to hold the lock beyond this point for now.
		p.queue.Mutex.Unlock()

		// Keep popping from the queue until the queue is empty again, in which case we start all
		// over again from the top.
		for {
			job, doneFunc := p.queue.Pop()
			if job == nil {
				break
			}

			select {
			case tasks <- &cloneTask{
				cloneJob: *job,
				done:     doneFunc,
			}:
			case <-ctx.Done():
				p.logger.Error("cloneJobProducer", log.Error(ctx.Err()))
				return
			}
		}
	}
}

func (p *clonePipelineRoutine) cloneJobConsumer(ctx context.Context, tasks <-chan *cloneTask) {
	logger := p.s.Logger.Scoped("cloneJobConsumer")

	for task := range tasks {
		logger := logger.With(log.String("job.repo", string(task.repo)))

		select {
		case <-ctx.Done():
			logger.Error("context done", log.Error(ctx.Err()))
			return
		default:
		}

		ctx, cancel, err := p.s.acquireCloneLimiter(ctx)
		if err != nil {
			logger.Error("acquireCloneLimiter", log.Error(err))
			continue
		}

		go func(task *cloneTask) {
			defer cancel()

			err := p.s.doClone(ctx, task.repo, task.dir, task.syncer, task.lock, task.remoteURL, task.options)
			if err != nil {
				logger.Error("failed to clone repo", log.Error(err))
			}
			// Use a different context in case we failed because the original context failed.
			setLastErrorNonFatal(p.s.ctx, p.logger, p.s.DB, p.s.Hostname, task.repo, err)
			_ = task.done()
		}(task)
	}
}

// cloneOptions specify optional behaviour for the cloneRepo function.
type CloneOptions struct {
	// Block will wait for the clone to finish before returning. If the clone
	// fails, the error will be returned. The passed in context is
	// respected. When not blocking the clone is done with a server background
	// context.
	Block bool

	// Overwrite will overwrite the existing clone.
	Overwrite bool
}

// cloneJob abstracts away a repo and necessary metadata to clone it. In the future it may be
// possible to simplify this, but to do that, doClone will need to do a lot less than it does at the
// moment.
type cloneJob struct {
	repo   api.RepoName
	dir    common.GitDir
	syncer vcssyncer.VCSSyncer

	// TODO: cloneJobConsumer should acquire a new lock. We are trying to keep the changes simple
	// for the time being. When we start using the new approach of using long lived goroutines for
	// cloning we will refactor doClone to acquire a new lock.
	lock RepositoryLock

	remoteURL *vcs.URL
	options   CloneOptions
}

// cloneTask is a thin wrapper around a cloneJob to associate the doneFunc with each job.
type cloneTask struct {
	*cloneJob
	done func() time.Duration
}

// NewCloneQueue initializes a new cloneQueue.
func NewCloneQueue(obctx *observation.Context, jobs *list.List) *common.Queue[*cloneJob] {
	return common.NewQueue[*cloneJob](obctx, "clone-queue", jobs)
}

// CloneRepo performs a clone operation for the given repository. It is
// non-blocking by default.
func (s *Server) CloneRepo(ctx context.Context, repo api.RepoName, opts CloneOptions) (cloneProgress string, err error) {
	if isAlwaysCloningTest(repo) {
		return "This will never finish cloning", nil
	}

	dir := gitserverfs.RepoDirFromName(s.ReposDir, repo)

	// PERF: Before doing the network request to check if isCloneable, lets
	// ensure we are not already cloning.
	if progress, cloneInProgress := s.Locker.Status(dir); cloneInProgress {
		return progress, nil
	}

	// We always want to store whether there was an error cloning the repo, but only
	// after we checked if a clone is already in progress, otherwise we would race with
	// the actual running clone for the DB state of last_error.
	defer func() {
		// Use a different context in case we failed because the original context failed.
		setLastErrorNonFatal(s.ctx, s.Logger, s.DB, s.Hostname, repo, err)
	}()

	syncer, err := s.GetVCSSyncer(ctx, repo)
	if err != nil {
		return "", errors.Wrap(err, "get VCS syncer")
	}

	// We may be attempting to clone a private repo so we need an internal actor.
	remoteURL, err := s.getRemoteURL(actor.WithInternalActor(ctx), repo)
	if err != nil {
		return "", err
	}

	// isCloneable causes a network request, so we limit the number that can
	// run at one time. We use a separate semaphore to cloning since these
	// checks being blocked by a few slow clones will lead to poor feedback to
	// users. We can defer since the rest of the function does not block this
	// goroutine.
	ctx, cancel, err := s.acquireCloneableLimiter(ctx)
	if err != nil {
		return "", err // err will be a context error
	}
	defer cancel()

	if err = s.RPSLimiter.Wait(ctx); err != nil {
		return "", err
	}

	if err := syncer.IsCloneable(ctx, repo, remoteURL); err != nil {
		redactedErr := urlredactor.New(remoteURL).Redact(err.Error())
		return "", errors.Errorf("error cloning repo: repo %s not cloneable: %s", repo, redactedErr)
	}

	// Mark this repo as currently being cloned. We have to check again if someone else isn't already
	// cloning since we released the lock. We released the lock since isCloneable is a potentially
	// slow operation.
	lock, ok := s.Locker.TryAcquire(dir, "starting clone")
	if !ok {
		// Someone else beat us to it
		status, _ := s.Locker.Status(dir)
		return status, nil
	}

	if s.skipCloneForTests {
		lock.Release()
		return "", nil
	}

	if opts.Block {
		ctx, cancel, err := s.acquireCloneLimiter(ctx)
		if err != nil {
			return "", err
		}
		defer cancel()

		// We are blocking, so use the passed in context.
		err = s.doClone(ctx, repo, dir, syncer, lock, remoteURL, opts)
		err = errors.Wrapf(err, "failed to clone %s", repo)
		return "", err
	}

	// We push the cloneJob to a queue and let the producer-consumer pipeline take over from this
	// point. See definitions of cloneJobProducer and cloneJobConsumer to understand how these jobs
	// are processed.
	s.CloneQueue.Push(&cloneJob{
		repo:      repo,
		dir:       dir,
		syncer:    syncer,
		lock:      lock,
		remoteURL: remoteURL,
		options:   opts,
	})

	return "", nil
}

func (s *Server) doClone(
	ctx context.Context,
	repo api.RepoName,
	dir common.GitDir,
	syncer vcssyncer.VCSSyncer,
	lock RepositoryLock,
	remoteURL *vcs.URL,
	opts CloneOptions,
) (err error) {
	logger := s.Logger.Scoped("doClone").With(log.String("repo", string(repo)))

	defer lock.Release()
	defer func() {
		if err != nil {
			repoCloneFailedCounter.Inc()
		}
	}()
	if err := s.RPSLimiter.Wait(ctx); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, conf.GitLongCommandTimeout())
	defer cancel()

	dstPath := string(dir)
	if !opts.Overwrite {
		// We clone to a temporary directory first, so avoid wasting resources
		// if the directory already exists.
		if _, err := os.Stat(dstPath); err == nil {
			return &os.PathError{
				Op:   "cloneRepo",
				Path: dstPath,
				Err:  os.ErrExist,
			}
		}
	}

	// We clone to a temporary location first to avoid having incomplete
	// clones in the repo tree. This also avoids leaving behind corrupt clones
	// if the clone is interrupted.
	tmpDir, err := gitserverfs.TempDir(s.ReposDir, "clone-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	tmpPath := filepath.Join(tmpDir, ".git")

	// It may already be cloned
	if !repoCloned(dir) {
		if err := s.DB.GitserverRepos().SetCloneStatus(ctx, repo, types.CloneStatusCloning, s.Hostname); err != nil {
			s.Logger.Error("Setting clone status in DB", log.Error(err))
		}
	}
	defer func() {
		// Use a background context to ensure we still update the DB even if we time out
		if err := s.DB.GitserverRepos().SetCloneStatus(context.Background(), repo, cloneStatus(repoCloned(dir), false), s.Hostname); err != nil {
			s.Logger.Error("Setting clone status in DB", log.Error(err))
		}
	}()

	logger.Info("cloning repo", log.String("tmp", tmpDir), log.String("dst", dstPath))

	progressReader, progressWriter := io.Pipe()
	// We also capture the entire output in memory for the call to SetLastOutput
	// further down.
	// TODO: This might require a lot of memory depending on the amount of logs
	// produced, the ideal solution would be that readCloneProgress stores it in
	// chunks.
	output := &linebasedBufferedWriter{}
	eg := readCloneProgress(s.DB, logger, lock, io.TeeReader(progressReader, output), repo)

	cloneErr := syncer.Clone(ctx, repo, remoteURL, dir, tmpPath, progressWriter)
	progressWriter.Close()

	if err := eg.Wait(); err != nil {
		s.Logger.Error("reading clone progress", log.Error(err))
	}

	// best-effort update the output of the clone
	if err := s.DB.GitserverRepos().SetLastOutput(context.Background(), repo, output.String()); err != nil {
		s.Logger.Error("Setting last output in DB", log.Error(err))
	}

	if cloneErr != nil {
		// TODO: Should we really return the entire output here in an error?
		// It could be a super big error string.
		return errors.Wrapf(cloneErr, "clone failed. Output: %s", output.String())
	}

	if testRepoCorrupter != nil {
		testRepoCorrupter(ctx, common.GitDir(tmpPath))
	}

	if err := postRepoFetchActions(ctx, logger, s.DB, s.Hostname, s.RecordingCommandFactory, s.ReposDir, repo, common.GitDir(tmpPath), remoteURL, syncer); err != nil {
		return err
	}

	if opts.Overwrite {
		// remove the current repo by putting it into our temporary directory, outside of the git repo.
		err := fileutil.RenameAndSync(dstPath, filepath.Join(tmpDir, "old"))
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to remove old clone")
		}
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return err
	}
	if err := fileutil.RenameAndSync(tmpPath, dstPath); err != nil {
		return err
	}

	logger.Info("repo cloned")
	repoClonedCounter.Inc()

	s.Perforce.EnqueueChangelistMappingJob(perforce.NewChangelistMappingJob(repo, dir))

	return nil
}

// linebasedBufferedWriter is an io.Writer that writes to a buffer.
// '\r' resets the write offset to the index after last '\n' in the buffer,
// or the beginning of the buffer if a '\n' has not been written yet.
//
// This exists to remove intermediate progress reports from "git clone
// --progress".
type linebasedBufferedWriter struct {
	// writeOffset is the offset in buf where the next write should begin.
	writeOffset int

	// afterLastNewline is the index after the last '\n' in buf
	// or 0 if there is no '\n' in buf.
	afterLastNewline int

	buf []byte
}

func (w *linebasedBufferedWriter) Write(p []byte) (n int, err error) {
	l := len(p)
	for {
		if len(p) == 0 {
			// If p ends in a '\r' we still want to include that in the buffer until it is overwritten.
			break
		}
		idx := bytes.IndexAny(p, "\r\n")
		if idx == -1 {
			w.buf = append(w.buf[:w.writeOffset], p...)
			w.writeOffset = len(w.buf)
			break
		}
		w.buf = append(w.buf[:w.writeOffset], p[:idx+1]...)
		switch p[idx] {
		case '\n':
			w.writeOffset = len(w.buf)
			w.afterLastNewline = len(w.buf)
			p = p[idx+1:]
		case '\r':
			// Record that our next write should overwrite the data after the most recent newline.
			// Don't slice it off immediately here, because we want to be able to return that output
			// until it is overwritten.
			w.writeOffset = w.afterLastNewline
			p = p[idx+1:]
		default:
			panic(fmt.Sprintf("unexpected char %q", p[idx]))
		}
	}
	return l, nil
}

// String returns the contents of the buffer as a string.
func (w *linebasedBufferedWriter) String() string {
	return string(w.buf)
}

// Bytes returns the contents of the buffer.
func (w *linebasedBufferedWriter) Bytes() []byte {
	return w.buf
}

// readCloneProgress scans the reader and saves the most recent line of output
// as the lock status, writes to a log file if siteConfig.cloneProgressLog is
// enabled, and optionally to the database when the feature flag `clone-progress-logging`
// is enabled.
func readCloneProgress(db database.DB, logger log.Logger, lock RepositoryLock, pr io.Reader, repo api.RepoName) *errgroup.Group {
	// Use a background context to ensure we still update the DB even if we
	// time out. IE we intentionally don't take an input ctx.
	ctx := featureflag.WithFlags(context.Background(), db.FeatureFlags())
	enableExperimentalDBCloneProgress := featureflag.FromContext(ctx).GetBoolOr("clone-progress-logging", false)

	var logFile *os.File

	if conf.Get().CloneProgressLog {
		var err error
		logFile, err = os.CreateTemp("", "")
		if err != nil {
			logger.Warn("failed to create temporary clone log file", log.Error(err), log.String("repo", string(repo)))
		} else {
			logger.Info("logging clone output", log.String("file", logFile.Name()), log.String("repo", string(repo)))
			defer logFile.Close()
		}
	}

	dbWritesLimiter := rate.NewLimiter(rate.Limit(1.0), 1)
	scan := bufio.NewScanner(pr)
	scan.Split(scanCRLF)
	store := db.GitserverRepos()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		for scan.Scan() {
			progress := scan.Text()
			lock.SetStatus(progress)

			if logFile != nil {
				// Failing to write here is non-fatal and we don't want to spam our logs if there
				// are issues
				_, _ = fmt.Fprintln(logFile, progress)
			}
			// Only write to the database persisted status if line indicates progress
			// which is recognized by presence of a '%'. We filter these writes not to waste
			// rate-limit tokens on log lines that would not be relevant to the user.
			if enableExperimentalDBCloneProgress {
				if strings.Contains(progress, "%") && dbWritesLimiter.Allow() {
					if err := store.SetCloningProgress(ctx, repo, progress); err != nil {
						logger.Error("error updating cloning progress in the db", log.Error(err))
					}
				}
			}
		}
		if err := scan.Err(); err != nil {
			return err
		}

		return nil
	})

	return eg
}

// scanCRLF is similar to bufio.ScanLines except it splits on both '\r' and '\n'
// and it does not return tokens that contain only whitespace.
func scanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	trim := func(data []byte) []byte {
		data = bytes.TrimSpace(data)
		if len(data) == 0 {
			// Don't pass back a token that is all whitespace.
			return nil
		}
		return data
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, trim(data[:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), trim(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// maybeStartClone checks if a given repository is cloned on disk. If not, it starts
// cloning the repository in the background and returns a NotFound error, if no current
// clone operation is running for that repo yet. If it is already cloning, a NotFound
// error with CloneInProgress: true is returned.
// Note: If disableAutoGitUpdates is set in the site config, no operation is taken and
// a NotFound error is returned.
func (s *Server) maybeStartClone(ctx context.Context, logger log.Logger, repo api.RepoName) (notFound *protocol.NotFoundPayload, cloned bool) {
	dir := gitserverfs.RepoDirFromName(s.ReposDir, repo)
	if repoCloned(dir) {
		return nil, true
	}

	if conf.Get().DisableAutoGitUpdates {
		logger.Debug("not cloning on demand as DisableAutoGitUpdates is set")
		return &protocol.NotFoundPayload{}, false
	}

	cloneProgress, cloneInProgress := s.Locker.Status(dir)
	if cloneInProgress {
		return &protocol.NotFoundPayload{
			CloneInProgress: true,
			CloneProgress:   cloneProgress,
		}, false
	}

	cloneProgress, err := s.CloneRepo(ctx, repo, CloneOptions{})
	if err != nil {
		logger.Debug("error starting repo clone", log.String("repo", string(repo)), log.Error(err))
		return &protocol.NotFoundPayload{CloneInProgress: false}, false
	}

	return &protocol.NotFoundPayload{
		CloneInProgress: true,
		CloneProgress:   cloneProgress,
	}, false
}
