package idx

import (
	"context"
	"fmt"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type workQueue struct {
	enqueue chan<- string          // channel of inputs
	dequeue chan<- (chan<- string) // channel of task executors
}

func NewQueue(lengthGauge prometheus.Gauge) *workQueue {
	enqueue, dequeue := queueWithoutDuplicates(lengthGauge)
	return &workQueue{enqueue: enqueue, dequeue: dequeue}
}

func (w *workQueue) Enqueue(repo string) {
	w.enqueue <- repo
}

// queueWithoutDuplicates provides a queue that ignores a new entry if it is already enqueued.
// Sending to the dequeue channel blocks if no entry is available.
func queueWithoutDuplicates(lengthGauge prometheus.Gauge) (enqueue chan<- string, dequeue chan<- (chan<- string)) {
	var queue []string
	set := make(map[string]struct{})
	enqueueChan := make(chan string)
	dequeueChan := make(chan (chan<- string))

	go func() {
		for {
			if len(queue) == 0 {
				repo := <-enqueueChan
				queue = append(queue, repo)
				set[repo] = struct{}{}
				if lengthGauge != nil {
					lengthGauge.Set(float64(len(queue)))
				}
			}

			select {
			case repo := <-enqueueChan:
				if _, ok := set[repo]; ok {
					continue // duplicate, discard
				}
				queue = append(queue, repo)
				set[repo] = struct{}{}
				if lengthGauge != nil {
					lengthGauge.Set(float64(len(queue)))
				}
			case c := <-dequeueChan:
				repo := queue[0]
				queue = queue[1:]
				delete(set, repo)
				if lengthGauge != nil {
					lengthGauge.Set(float64(len(queue)))
				}
				c <- repo
			}
		}
	}()

	return enqueueChan, dequeueChan
}

var (
	// Global state shared by all work threads
	currentJobs   = make(map[string]struct{})
	currentJobsMu sync.Mutex
)

func Work(ctx context.Context, w *workQueue, s svc) {
	if s == nil {
		s = DefaultSvc
	}
	for {
		c := make(chan string)
		w.dequeue <- c
		repo := <-c

		{
			currentJobsMu.Lock()
			if _, ok := currentJobs[repo]; ok {
				currentJobsMu.Unlock()
				return // in progress, discard
			}
			currentJobs[repo] = struct{}{}
			currentJobsMu.Unlock()
		}

		if err := index(ctx, w, s, repo); err != nil {
			log15.Error("Indexing failed", "repo", repo, "error", err)
		}

		{
			currentJobsMu.Lock()
			delete(currentJobs, repo)
			currentJobsMu.Unlock()
		}
	}
}

func index(ctx context.Context, wq *workQueue, s svc, repoName string) error {
	repo, err := s.GetByURI(ctx, repoName)
	if err != nil {
		return fmt.Errorf("Repos.GetByURI failed: %s", err)
	}

	headCommit, err := s.ResolveRevision(ctx, repo, "HEAD")
	if err != nil {
		// If clone is in progress, re-enqueue after 5 seconds
		if _, ok := err.(vcs.RepoNotExistError); ok && err.(vcs.RepoNotExistError).CloneInProgress {
			go func() {
				time.Sleep(5 * time.Second)
				wq.Enqueue(repoName)
			}()
			return nil
		}
		return fmt.Errorf("ResolveRevision failed: %s", err)
	}

	inv, err := s.GetInventoryUncached(ctx, &sourcegraph.RepoRevSpec{
		Repo:     repo.ID,
		CommitID: string(headCommit),
	})
	if err != nil {
		return fmt.Errorf("Repos.GetInventory failed: %s", err)
	}

	if repo.IndexedRevision != nil && *repo.IndexedRevision == string(headCommit) {
		return nil // index is up-to-date
	}

	log15.Info("Indexing started", "repo", repoName, "headCommit", headCommit)
	defer log15.Info("Indexing finished", "repo", repoName, "headCommit", headCommit)

	// Global refs & packages indexing. Neither index forks.
	if !repo.Fork {
		// Global refs stores and queries private repository data separately,
		// so it is fine to index private repositories.
		err = s.DefsRefreshIndex(ctx, repo.URI, string(headCommit))
		if err != nil {
			return fmt.Errorf("Defs.RefreshIndex failed: %s", err)
		}

		// As part of package indexing, it's fine to index private repositories
		// because backend.Pkgs.ListPackages is responsible for authentication
		// checks.
		err = s.PkgsRefreshIndex(ctx, repo.URI, string(headCommit))
		if err != nil {
			return fmt.Errorf("Pkgs.RefreshIndex failed: %s", err)
		}

		// Spider out to index dependencies
		if err := enqueueDependencies(ctx, wq, s, inv.PrimaryProgrammingLanguage(), repo.ID); err != nil {
			return err
		}
	}

	if err := s.Update(ctx, &sourcegraph.ReposUpdateOp{
		Repo:            repo.ID,
		IndexedRevision: string(headCommit),
		Language:        inv.PrimaryProgrammingLanguage(),
	}); err != nil {
		return fmt.Errorf("Repos.Update failed: %s", err)
	}

	return nil
}

// enqueueDependencies makes a best effort to enqueue dependencies of the specified repository
// (repoID) for certain languages. The languages covered are languages where the language server
// itself cannot resolve dependencies to source repository URLs. For those languages, dependency
// repositories must be indexed before cross-repo jump-to-def can work. enqueueDependencies tries to
// best-effort determine what those dependencies are and enqueue them.
func enqueueDependencies(ctx context.Context, wq *workQueue, s svc, lang string, repoID int32) error {
	// do nothing if this is not a language that requires heuristic dependency resolution
	if lang != "Java" {
		return nil
	}

	log15.Info("Enqueuing dependencies for repo", "repo", repoID, "lang", lang)

	deps, err := s.Dependencies(ctx, repoID, true) // exclude private dependencies, because we don't have GitHub creds in the indexer
	if err != nil {
		return fmt.Errorf("Pkgs.DependencyReferences failed: %s", err)
	}

	// Filter out already-indexed dependencies
	var unfetchedDeps []*sourcegraph.DependencyReference
	for _, dep := range deps {
		pkgs, err := s.ListPackages(ctx, &sourcegraph.ListPackagesOp{Lang: lang, PkgQuery: depReferenceToPkgQuery(lang, dep), Limit: 1})
		if err != nil {
			return err
		}
		if len(pkgs) == 0 {
			unfetchedDeps = append(unfetchedDeps, dep)
		}
	}

	// Resolve and enqueue unindexed dependencies for indexing
	resolvedDeps := resolveDependencies(ctx, s, lang, unfetchedDeps)
	resolvedDepsList := make([]string, 0, len(resolvedDeps))
	for rawDepRepo, _ := range resolvedDeps {
		repo, err := s.GetByURI(ctx, rawDepRepo)
		if err != nil {
			log15.Warn("Could not resolve repository, skipping", "repo", rawDepRepo, "error", err)
			continue
		}
		wq.Enqueue(repo.URI)
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
func resolveDependencies(ctx context.Context, s svc, lang string, deps []*sourcegraph.DependencyReference) map[string]struct{} {
	switch lang {
	case "Java":
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
			depRepoURI, err := s.GoogleGitHub(depQuery)
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
