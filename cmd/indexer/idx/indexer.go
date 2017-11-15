package idx

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/prometheus/client_golang/prometheus"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type qitem struct {
	repo string
	rev  string
}

type workQueue struct {
	enqueue chan<- qitem          // channel of inputs
	dequeue chan<- (chan<- qitem) // channel of task executors
}

func NewQueue(lengthGauge prometheus.Gauge) *workQueue {
	enqueue, dequeue := queueWithoutDuplicates(lengthGauge)
	return &workQueue{enqueue: enqueue, dequeue: dequeue}
}

func (w *workQueue) Enqueue(repo string, rev string) {
	w.enqueue <- qitem{repo: repo, rev: rev}
}

// queueWithoutDuplicates provides a queue that ignores a new entry if it is already enqueued.
// Sending to the dequeue channel blocks if no entry is available.
func queueWithoutDuplicates(lengthGauge prometheus.Gauge) (enqueue chan<- qitem, dequeue chan<- (chan<- qitem)) {
	var queue []qitem
	set := make(map[qitem]struct{})
	enqueueChan := make(chan qitem)
	dequeueChan := make(chan (chan<- qitem))

	go func() {
		for {
			if len(queue) == 0 {
				repoRev := <-enqueueChan
				queue = append(queue, repoRev)
				set[repoRev] = struct{}{}
				if lengthGauge != nil {
					lengthGauge.Set(float64(len(queue)))
				}
			}

			select {
			case repoRev := <-enqueueChan:
				if _, ok := set[repoRev]; ok {
					continue // duplicate, discard
				}
				queue = append(queue, repoRev)
				set[repoRev] = struct{}{}
				if lengthGauge != nil {
					lengthGauge.Set(float64(len(queue)))
				}
			case c := <-dequeueChan:
				repoRev := queue[0]
				queue = queue[1:]
				delete(set, repoRev)
				if lengthGauge != nil {
					lengthGauge.Set(float64(len(queue)))
				}
				c <- repoRev
			}
		}
	}()

	return enqueueChan, dequeueChan
}

var (
	// Global state shared by all work threads
	currentJobs   = make(map[qitem]struct{})
	currentJobsMu sync.Mutex
)

func Work(ctx context.Context, w *workQueue) {
	for {
		c := make(chan qitem)
		w.dequeue <- c
		repoRev := <-c

		{
			currentJobsMu.Lock()
			if _, ok := currentJobs[repoRev]; ok {
				currentJobsMu.Unlock()
				return // in progress, discard
			}
			currentJobs[repoRev] = struct{}{}
			currentJobsMu.Unlock()
		}

		if err := index(ctx, w, repoRev.repo, repoRev.rev); err != nil {
			log15.Error("Indexing failed", "repoRev", repoRev, "error", err)
		}

		{
			currentJobsMu.Lock()
			delete(currentJobs, repoRev)
			currentJobsMu.Unlock()
		}
	}
}

func index(ctx context.Context, wq *workQueue, repoName string, rev string) error {
	if rev == "" {
		rev = "HEAD"
	}
	repo, err := backend.Repos.GetByURI(ctx, repoName)
	if err != nil {
		return fmt.Errorf("Repos.GetByURI failed: %s", err)
	}

	var headCommit vcs.CommitID
	if len(rev) == 40 {
		headCommit = vcs.CommitID(rev)
	} else {
		headCommit, err = ResolveRevision(ctx, repo, rev)
		if err != nil {
			if repo.URI == "github.com/sourcegraphtest/AlwaysCloningTest" {
				// Avoid infinite loop for always cloning test.
				return nil
			}
			// If clone is in progress, re-enqueue after 5 seconds
			if _, ok := err.(vcs.RepoNotExistError); ok && err.(vcs.RepoNotExistError).CloneInProgress {
				go func() {
					time.Sleep(5 * time.Second)
					wq.Enqueue(repoName, rev)
				}()
				return nil
			}
			return fmt.Errorf("ResolveRevision failed: %s", err)
		}
		if len(headCommit) != 40 {
			log15.Error("ResolveRevision returned bad commit", "commit", headCommit)
		}
	}

	inv, err := backend.Repos.GetInventoryUncached(ctx, &sourcegraph.RepoRevSpec{
		Repo:     repo.ID,
		CommitID: string(headCommit),
	})

	if err != nil {
		return fmt.Errorf("Repos.GetInventory failed: %s", err)
	}

	if repo.IndexedRevision != nil && (repo.FreezeIndexedRevision || *repo.IndexedRevision == string(headCommit)) {
		return nil // index is up-to-date
	}

	log15.Info("Indexing started", "repo", repoName, "headCommit", headCommit)
	defer log15.Info("Indexing finished", "repo", repoName, "headCommit", headCommit)

	// Global refs & packages indexing. Neither index forks.
	if !repo.Fork {
		// Global refs stores and queries private repository data separately,
		// so it is fine to index private repositories.
		defErr := backend.Defs.RefreshIndex(ctx, repo.URI, string(headCommit))
		if err != nil {
			defErr = fmt.Errorf("Defs.RefreshIndex failed: %s", err)
		}

		// As part of package indexing, it's fine to index private repositories
		// because backend.Pkgs.ListPackages is responsible for authentication
		// checks.
		pkgErr := backend.Pkgs.RefreshIndex(ctx, repo.URI, string(headCommit))
		if err != nil {
			pkgErr = fmt.Errorf("Pkgs.RefreshIndex failed: %s", err)
		}

		// Spider out to index dependencies
		spidErr := enqueueDependencies(ctx, wq, inv.PrimaryProgrammingLanguage(), repo.ID)

		if err := makeMultiErr(defErr, pkgErr, spidErr); err != nil {
			return err
		}
	}

	if err := localstore.Repos.UpdateIndexedRevision(ctx, repo.ID, string(headCommit)); err != nil {
		return fmt.Errorf("Repos.UpdateIndexedRevision failed: %s", err)
	}

	if err := localstore.Repos.UpdateLanguage(ctx, repo.ID, inv.PrimaryProgrammingLanguage()); err != nil {
		return fmt.Errorf("Repos.UpdateLanguage failed: %s", err)
	}

	return nil
}

// enqueueDependencies makes a best effort to enqueue dependencies of the specified repository
// (repoID) for certain languages. The languages covered are languages where the language server
// itself cannot resolve dependencies to source repository URLs. For those languages, dependency
// repositories must be indexed before cross-repo jump-to-def can work. enqueueDependencies tries to
// best-effort determine what those dependencies are and enqueue them.
func enqueueDependencies(ctx context.Context, wq *workQueue, lang string, repoID int32) error {
	// do nothing if this is not a language that requires heuristic dependency resolution
	if lang != "Java" {
		return nil
	}

	log15.Info("Enqueuing dependencies for repo", "repo", repoID, "lang", lang)

	excludePrivate := !feature.Features.Sep20Auth
	deps, err := backend.Defs.Dependencies(ctx, repoID, excludePrivate)
	if err != nil {
		return fmt.Errorf("Defs.DependencyReferences failed: %s", err)
	}

	// Filter out already-indexed dependencies
	var unfetchedDeps []*sourcegraph.DependencyReference
	for _, dep := range deps {
		pkgs, err := backend.Pkgs.ListPackages(ctx, &sourcegraph.ListPackagesOp{Lang: lang, PkgQuery: depReferenceToPkgQuery(lang, dep), Limit: 1})
		if err != nil {
			return err
		}
		if len(pkgs) == 0 {
			unfetchedDeps = append(unfetchedDeps, dep)
		}
	}

	// Resolve and enqueue unindexed dependencies for indexing
	resolvedDeps := resolveDependencies(ctx, lang, unfetchedDeps)
	resolvedDepsList := make([]string, 0, len(resolvedDeps))
	for rawDepRepo, _ := range resolvedDeps {
		repo, err := backend.Repos.GetByURI(ctx, rawDepRepo)
		if err != nil {
			log15.Warn("Could not resolve repository, skipping", "repo", rawDepRepo, "error", err)
			continue
		}
		wq.Enqueue(repo.URI, "")
		resolvedDepsList = append(resolvedDepsList, repo.URI)
	}
	log15.Info("Enqueued dependencies for repo", "repo", repoID, "lang", lang, "num", len(resolvedDeps), "dependencies", resolvedDepsList)
	return nil
}

// depReferenceToPkgQuery maps from a DependencyReference to a package descriptor query that
// uniquely identifies the dependency package (typically discarding version information).  The
// mapping can be different for different languages, so languages are handled case-by-case.
func depReferenceToPkgQuery(lang string, dep *sourcegraph.DependencyReference) map[string]interface{} {
	switch lang {
	case "Java":
		return map[string]interface{}{"id": dep.DepData["id"]}
	default:
		return nil
	}
}

// resolveDependencies resolves a list of DependencyReferences to a set of source repository URLs.
// This mapping is different from language to language and is often heuristic, so different
// languages are handled case-by-case.
func resolveDependencies(ctx context.Context, lang string, deps []*sourcegraph.DependencyReference) map[string]struct{} {
	switch lang {
	case "Java":
		if !Google.Enabled() {
			return nil
		}

		// Best-effort fetch from GitHub via Google Search. Equivalent to searching for "site:github.com $groupId:$artifactId".
		depQueries := make(map[string]struct{})
		for _, dep := range deps {
			if dep.DepData == nil {
				continue
			}
			id, ok := dep.DepData["id"].(string)
			if !ok {
				continue
			}
			depQueries[id] = struct{}{}
		}
		resolvedDeps := map[string]struct{}{}
		for depQuery, _ := range depQueries {
			depRepoURI, err := Google.Search(depQuery)
			if err != nil {
				log15.Warn("Could not resolve dependency to repository via Google, skipping", "query", depQuery, "error", err)
				continue
			}
			resolvedDeps[depRepoURI] = struct{}{}
		}
		return resolvedDeps
	default:
		return nil
	}
}

type multiError []error

func makeMultiErr(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	var nonnil []error
	for _, err := range errs {
		if err != nil {
			nonnil = append(nonnil, err)
		}
	}
	if len(nonnil) == 0 {
		return nil
	}
	if len(nonnil) == 1 {
		return nonnil[0]
	}
	return multiError(nonnil)
}

func (e multiError) Error() string {
	errs := make([]string, len(e))
	for i := 0; i < len(e); i++ {
		errs[i] = e[i].Error()
	}
	return fmt.Sprintf("multiple errors:\n\t%s", strings.Join(errs, "\n\t"))
}
