package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/cmd/precise-code-intel-tester/util"
)

// uploadCommand runs the "upload" command.
func uploadCommand() error {
	ctx, cleanup := util.SignalSensitiveContext()
	defer cleanup()

	start := time.Now()

	if err := uploadIndexes(ctx); err != nil {
		return err
	}

	fmt.Printf("All uploads completed processing in %s\n", time.Since(start))
	return nil
}

// Upload represents a fully uploaded (but possibly unprocessed) LSIF index.
type Upload struct {
	Name     string
	Index    int
	Rev      string
	UploadID string
}

// uploadIndexes uploads each file in the index directory and blocks until each upload has
// been successfully processed.
func uploadIndexes(ctx context.Context) error {
	revsByRepo, err := readRevsByRepo()
	if err != nil {
		return err
	}

	total := countRevs(revsByRepo)
	uploaded := make(chan Upload, total)
	processedSignals := makeProcessedSignals(revsByRepo)

	// Watch API for changes in state, and inform workers when their upload has been processed
	go watchStateChanges(ctx, uploaded, processedSignals)

	limiter := util.NewLimiter(numConcurrentUploads)
	defer limiter.Close()

	var fns []util.ParallelFn
	for name, revs := range revsByRepo {
		fns = append(fns, makeTestUploadForRepositoryFunction(name, revs, uploaded, processedSignals, limiter))
	}

	return util.RunParallel(ctx, total, fns)
}

// indexFilenamePattern extracts a repo name and rev from the index filename. We assume that the
// index segment here (the non-captured `.\d+.`) occupies [0,n) without gaps for each repository.
var indexFilenamePattern = regexp.MustCompile(`^(.+)\.\d+\.([0-9A-Fa-f]{40})\.dump$`)

// readRevsByRepo returns a list of revisions by repository names for which there is an index file.
func readRevsByRepo() (map[string][]string, error) {
	infos, err := ioutil.ReadDir(indexDir)
	if err != nil {
		return nil, err
	}

	revsByRepo := map[string][]string{}
	for _, info := range infos {
		matches := indexFilenamePattern.FindStringSubmatch(info.Name())
		if len(matches) > 0 {
			revsByRepo[matches[1]] = append(revsByRepo[matches[1]], matches[2])
		}
	}

	return revsByRepo, nil
}

// countRevs returns the total number of revision.
func countRevs(revsByRepo map[string][]string) int {
	total := 0
	for _, revs := range revsByRepo {
		total += len(revs)
	}

	return total
}

// makeProcessedSignals returns a map of error channels for each revision.
func makeProcessedSignals(revsByRepo map[string][]string) map[string]map[string]chan error {
	processedSignals := map[string]map[string]chan error{}
	for repo, revs := range revsByRepo {
		revMap := make(map[string]chan error, len(revs))
		for _, rev := range revs {
			revMap[rev] = make(chan error, 1)
		}

		processedSignals[repo] = revMap
	}

	return processedSignals
}

// watchStateChanges maintains a list of uploaded but nonterminal upload records. This function
// polls the API and signals the worker when their upload has been processed. If an upload fails
// to process, the error will be sent to the worker.
func watchStateChanges(ctx context.Context, uploaded chan Upload, processedSignals map[string]map[string]chan error) {
	var uploads []Upload

	for {
		select {
		case upload := <-uploaded:
			// Upload complete, add to process watch list
			uploads = append(uploads, upload)

		case <-time.After(time.Millisecond * 500):
			// Check states

		case <-ctx.Done():
			// Close all signal channels (avoid race)
			for name, revs := range processedSignals {
				for rev := range revs {
					ch := processedSignals[name][rev]
					delete(processedSignals[name], rev)
					close(ch)
				}
			}

			return
		}

		var ids []string
		for _, upload := range uploads {
			ids = append(ids, upload.UploadID)
		}

		states, err := uploadStates(ctx, ids)
		if err != nil {
			// Send err to everybody and exit
			for name, revs := range processedSignals {
				for rev := range revs {
					ch := processedSignals[name][rev]
					delete(processedSignals[name], rev)
					ch <- err
					close(ch)
				}
			}

			return
		}

		// Remove terminal uploads and send signals to workers
		uploads = filterUploadsByState(uploads, processedSignals, states)
	}
}

// uploadStates returns a map from upload identifier to its current state.
func uploadStates(ctx context.Context, ids []string) (map[string]string, error) {
	var fragments []string
	for i, id := range ids {
		fragments = append(fragments, fmt.Sprintf(`
			u%d: node(id: "%s") {
				... on LSIFUpload {
					state
				}
			}
		`, i, id))
	}
	query := fmt.Sprintf("{%s}", strings.Join(fragments, "\n"))

	payload := struct {
		Data map[string]struct {
			State string `json:"state"`
		} `json:"data"`
	}{}
	if err := util.QueryGraphQL(ctx, endpoint, token, query, nil, &payload); err != nil {
		return nil, err
	}

	states := map[string]string{}
	for i, id := range ids {
		states[id] = payload.Data[fmt.Sprintf("u%d", i)].State
	}

	return states, nil
}

// filterUploadsByState filters all terminal uploads from the input list. For each terminal upload,
// the corresponding channel is closed to inform the worker that it can unblock. If the upload failed
// to process, a meaningful error value is passed to the worker.
func filterUploadsByState(uploads []Upload, processedSignals map[string]map[string]chan error, states map[string]string) []Upload {
	nonterminals := make([]Upload, 0, len(uploads))

	for _, upload := range uploads {
		var err error
		switch states[upload.UploadID] {
		case "ERRORED":
			err = errors.New("processing failed")
			fallthrough

		case "COMPLETED":
			ch := processedSignals[upload.Name][upload.Rev]
			delete(processedSignals[upload.Name], upload.Rev)
			ch <- err
			close(ch)

		default:
			nonterminals = append(nonterminals, upload)
		}
	}

	return nonterminals
}

// makeTestUploadForRepositoryFunction constructs a function for RunParallel that uploads the index files
// for the given repo, then blocks until the upload records enter a terminal state. If any upload fails to
// process, an error is returned.
func makeTestUploadForRepositoryFunction(name string, revs []string, uploaded chan Upload, processedSignals map[string]map[string]chan error, limiter *util.Limiter) util.ParallelFn {
	var numUploaded uint32
	var numProcessed uint32

	return util.ParallelFn{
		Fn: func(ctx context.Context) error {
			var wg sync.WaitGroup
			ch := make(chan error, len(revs))

			for i, rev := range revs {
				id, err := upload(ctx, name, i, rev, limiter)
				if err != nil {
					return err
				}
				atomic.AddUint32(&numUploaded, 1)

				wg.Add(1)
				go func() {
					defer wg.Done()
					ch <- <-processedSignals[name][rev]
				}()

				select {
				// send id to monitor
				case uploaded <- Upload{Name: name, Index: i, Rev: rev, UploadID: id}:

				case <-ctx.Done():
					return ctx.Err()
				}
			}

			go func() {
				wg.Wait()
				close(ch)
			}()

			// wait for all uploads to process
			for {
				select {
				case err, ok := <-ch:
					if err != nil || !ok {
						return err
					}
					atomic.AddUint32(&numProcessed, 1)

				case <-ctx.Done():
					return ctx.Err()
				}
			}
		},
		Description: func() string {
			if n := atomic.LoadUint32(&numUploaded); n < uint32(len(revs)) {
				return fmt.Sprintf("Uploading index %d of %d for %s...", n+1, len(revs), name)
			}

			return fmt.Sprintf("Waiting for uploads to process for %s...", name)
		},
		Total:    func() int { return len(revs) },
		Finished: func() int { return int(atomic.LoadUint32(&numProcessed)) },
	}
}

// uploadIDPattern extracts a GraphQL identifier from the output of the `src lsif upload` command.
var uploadIDPattern = regexp.MustCompile(`/settings/code-intelligence/lsif-uploads/([0-9A-Za-z=]+)`)

// upload invokes the `src lsif upload` command. This requires that src is installed on the
// current user's $PATH and is relatively up to date.
func upload(ctx context.Context, name string, index int, rev string, limiter *util.Limiter) (string, error) {
	if err := limiter.Acquire(ctx); err != nil {
		return "", err
	}
	defer limiter.Release()

	args := []string{
		fmt.Sprintf("-endpoint=%s", endpoint),
		"lsif",
		"upload",
		"-root=/",
		fmt.Sprintf("-repo=%s", fmt.Sprintf("github.com/%s/%s", "sourcegraph-testing", name)),
		fmt.Sprintf("-commit=%s", rev),
		fmt.Sprintf("-file=%s", filepath.Join(fmt.Sprintf("%s.%d.%s.dump", name, index, rev))),
	}

	cmd := exec.CommandContext(ctx, "src", args...)
	cmd.Dir = indexDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("error running 'src %s':\n%s\n", strings.Join(args, " "), output))
	}

	match := uploadIDPattern.FindSubmatch(output)
	if len(match) == 0 {
		return "", fmt.Errorf("failed to extract URL:\n%s", output)
	}

	return string(match[1]), nil
}
