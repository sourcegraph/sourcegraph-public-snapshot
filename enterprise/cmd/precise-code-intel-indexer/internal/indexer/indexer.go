package indexer

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/efritz/glock"
	"github.com/google/uuid"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type Indexer struct {
	options  IndexerOptions
	clock    glock.Clock
	indexIDs map[int]struct{}
	m        sync.RWMutex    // protects indexIDs
	ctx      context.Context // root context passed to commands
	cancel   func()          // cancels the root context
	wg       sync.WaitGroup  // tracks active background goroutines
	finished chan struct{}   // signals that Start has finished
}

type IndexerOptions struct {
	FrontendURL           string
	FrontendURLFromDocker string
	AuthToken             string
	PollInterval          time.Duration
	HeartbeatInterval     time.Duration
	Metrics               IndexerMetrics
}

const ImageBinary = "docker" // TODO - configure

func NewIndexer(ctx context.Context, options IndexerOptions) *Indexer {
	return newIndexer(ctx, options, glock.NewRealClock())
}

func newIndexer(ctx context.Context, options IndexerOptions, clock glock.Clock) *Indexer {
	ctx, cancel := context.WithCancel(ctx)

	return &Indexer{
		options:  options,
		clock:    clock,
		indexIDs: map[int]struct{}{},
		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
}

// Start runs the poll and heartbeat loops. This method blocks until all background
// goroutines have exited.
func (i *Indexer) Start() {
	defer close(i.finished)

	client := client.NewClient(uuid.New().String(), i.options.FrontendURL, i.options.AuthToken)

	i.wg.Add(2)
	go i.poll(client)
	go i.heartbeat(client)
	i.wg.Wait()
}

// poll begins polling for work from the API and indexing repositories.
func (i *Indexer) poll(client client.Client) {
	defer i.wg.Done()

loop:
	for {
		index, dequeued, err := client.Dequeue(i.ctx)
		if err != nil {
			for ex := err; ex != nil; ex = errors.Unwrap(ex) {
				if err == i.ctx.Err() {
					break loop
				}
			}

			log15.Error("Failed to dequeue index", "err", err)
		}

		delay := i.options.PollInterval
		if dequeued {
			log15.Info("Dequeued index for processing", "id", index.ID)

			if err := i.processAndComplete(client, index); err != nil {
				for ex := err; ex != nil; ex = errors.Unwrap(ex) {
					if err == i.ctx.Err() {
						break loop
					}
				}

				log15.Error("Failed to finalize index", "id", index.ID, "err", err)
			}

			// If we had a successful dequeue, do not wait the poll interval.
			// Just attempt to dequeue and process the next unit of work while
			// there are indexes to be processed.
			delay = 0
		}

		select {
		case <-i.clock.After(delay):
		case <-i.ctx.Done():
			break loop
		}
	}
}

// heartbeat sends a periodic request to the frontend to keep the database transactions
// locking the indexes dequeued by this instance alive.
func (i *Indexer) heartbeat(client client.Client) {
	defer i.wg.Done()

loop:
	for {
		if err := client.Heartbeat(i.ctx, i.getIDs()); err != nil {
			for ex := err; ex != nil; ex = errors.Unwrap(ex) {
				if err == i.ctx.Err() {
					break loop
				}
			}

			log15.Error("Failed to perform heartbeat", "err", err)
		}

		select {
		case <-time.After(i.options.HeartbeatInterval):
		case <-i.ctx.Done():
			break loop
		}
	}
}

// Stop will cause the indexer loop to exit after the current iteration. This is done by
// canceling the context passed to the subprocess functions (which may cause the currently
// processing unit of work to fail). This method blocks until all background goroutines have
// exited.
func (i *Indexer) Stop() {
	i.cancel()
	<-i.finished
}

// getIDs returns a slice of index identifiers that are currently being processed.
func (i *Indexer) getIDs() (ids []int) {
	i.m.RLock()
	defer i.m.RUnlock()

	for id := range i.indexIDs {
		ids = append(ids, id)
	}

	return ids
}

// addID adds the given index identifier from the set of currently processing indexes.
func (i *Indexer) addID(indexID int) {
	i.m.Lock()
	defer i.m.Unlock()
	i.indexIDs[indexID] = struct{}{}
}

// removeID removes the given index identifier from the set of currently processing indexes.
func (i *Indexer) removeID(indexID int) {
	i.m.Lock()
	defer i.m.Unlock()
	delete(i.indexIDs, indexID)
}

// process clones the target code, invokes the target indexer, uploads the result to the
// external frontend API, then marks the dequeued record as complete (or errored) in the
// index queue.
func (i *Indexer) processAndComplete(client client.Client, index store.Index) error {
	i.addID(index.ID)
	defer i.removeID(index.ID)

	indexErr := i.process(index)

	if err := client.Complete(i.ctx, index.ID, indexErr); err != nil {
		return err
	}

	if indexErr == nil {
		log15.Info("Marked index as complete", "id", index.ID)
	} else {
		log15.Warn("Marked index as errored", "id", index.ID, "err", indexErr)
	}

	return nil
}

// process clones the target code into a temporary directory, invokes the target indexer in
// a fresh docker container, and uploads the results to the external frontend API.
func (i *Indexer) process(index store.Index) error {
	repoDir, err := i.fetchRepository(index.RepositoryName, index.Commit)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(repoDir)
	}()

	indexAndUploadCommand := []string{
		"lsif-go",
		"&&",
		"src", "-endpoint", fmt.Sprintf(i.options.FrontendURLFromDocker), "lsif", "upload", "-repo", index.RepositoryName, "-commit", index.Commit,
	}

	if err := command(
		i.ctx,
		ImageBinary, "run", "--rm",
		"-v", fmt.Sprintf("%s:/data", repoDir),
		"-w", "/data",
		"sourcegraph/lsif-go:latest",
		"bash", "-c", strings.Join(indexAndUploadCommand, " "),
	); err != nil {
		return errors.Wrap(err, "failed to index repository")
	}

	return nil
}

// fetchRepository creates a temporary directory and performs a git checkout with the given repository
// and commit. If there is an error, the temporary directory is removed.
func (i *Indexer) fetchRepository(repositoryName, commit string) (string, error) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tempDir)
		}
	}()

	cloneURL, err := makeCloneURL(i.options.FrontendURL, i.options.AuthToken, repositoryName)
	if err != nil {
		return "", err
	}

	commands := [][]string{
		{"-C", tempDir, "init"},
		{"-C", tempDir, "-c", "protocol.version=2", "fetch", cloneURL.String(), commit},
		{"-C", tempDir, "checkout", commit},
	}

	for _, args := range commands {
		if err := command(i.ctx, "git", args...); err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed `git %s`", strings.Join(args, " ")))
		}
	}

	return tempDir, nil
}

func makeCloneURL(baseURL, authToken, repositoryName string) (*url.URL, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	base.User = url.UserPassword("indexer", authToken)

	return base.ResolveReference(&url.URL{Path: path.Join(".internal-code-intel", "git", repositoryName)}), nil
}
