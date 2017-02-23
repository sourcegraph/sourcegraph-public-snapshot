package zap

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/experimental_level"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap/internal/pkg/mutexmap"
	"github.com/sourcegraph/zap/server/refdb"
)

type serverRepo struct {
	refdb refdb.RefDB // the repo's refdb (safe for concurrent access)

	mu sync.Mutex
	// workspace:
	workspace       Workspace // set for non-bare repos added via workspace/add
	workspaceCancel func()    // tear down workspace
	config          RepoConfiguration

	refLocks mutexmap.MutexMap
}

// acquireRef acquires an exclusive lock that allows the holder to
// perform operations on the named ref. The release func must be
// called to release the lock.
//
// Callers can use the following to hold the lock for the remaining
// execution of the current function:
//
//   defer repo.acquireRef(refName)()
func (s *serverRepo) acquireRef(refName string) (release func()) {
	lock := s.refLocks.Get(refName)
	lock.Lock()
	return lock.Unlock
}

func (s *Server) getRepo(ctx context.Context, log *log.Context, repoName string) (*serverRepo, error) {
	repo, err := s.getRepoIfExists(ctx, log, repoName)
	if err != nil {
		return nil, err
	}
	if repo != nil {
		return repo, nil
	}

	// Only perform implicit repository creation below if the backend desires
	// it. For example, a local server will not desire this because all
	// repositories should be created via workspace/add (e.g. zap init) not
	// implicitly.
	if !s.backend.CanAutoCreate() {
		return nil, &jsonrpc2.Error{
			Code:    int64(ErrorCodeRepoNotExists),
			Message: fmt.Sprintf("repo not found: %s (add it with 'zap init')", repoName),
		}
	}

	s.reposMu.Lock()
	defer s.reposMu.Unlock()
	repo, exists := s.repos[repoName]
	if !exists {
		repo = &serverRepo{
			refdb: refdb.NewMemoryRefDB(),
		}
		s.repos[repoName] = repo
	}
	return repo, nil
}

func (s *Server) getRepoIfExists(ctx context.Context, log *log.Context, repoName string) (*serverRepo, error) {
	ok, err := s.backend.CanAccess(ctx, log, repoName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, &jsonrpc2.Error{
			Code:    int64(ErrorCodeRepoNotExists),
			Message: fmt.Sprintf("access forbidden to repo: %s", repoName),
		}
	}

	s.reposMu.Lock()
	defer s.reposMu.Unlock()
	return s.repos[repoName], nil
}

func (c *serverConn) handleRepoWatch(ctx context.Context, log *log.Context, repo *serverRepo, params RepoWatchParams) error {
	if params.Repo == "" {
		return &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: "repo is required"}
	}

	if err := params.validate(); err != nil {
		return err
	}

	{
		c.mu.Lock()
		if c.watchingRepos == nil {
			c.watchingRepos = map[string][]string{}
		}
		level.Info(log).Log("set-watch-refspec", fmt.Sprint(params.Refspecs), "old", fmt.Sprint(c.watchingRepos[params.Repo]))
		c.watchingRepos[params.Repo] = params.Refspecs
		c.mu.Unlock()
	}

	// Send over current state of all matched: for each non-symbolic
	// ref, send a ref/update; for each symbolic ref, send a
	// ref/updateSymbolic.
	//
	// From here on, clients will receive all future updates, so this
	// means they always have the full state of the repository.
	refs := refsMatchingRefspecs(repo.refdb, params.Refspecs)
	if len(refs) > 0 {
		for _, ref := range refs {
			if ref.IsSymbolic() {
				// Send all symbolic refs last, so that when the
				// client receives them, it has already received their
				// target refs. This makes client implementation
				// easier.
				continue
			}

			log := log.With("update-ref-downstream-with-initial-state", ref.Name)
			level.Debug(log).Log()
			refObj := ref.Object.(serverRef)
			if err := c.conn.Notify(ctx, "ref/update", RefUpdateDownstreamParams{
				RefIdentifier: RefIdentifier{Repo: params.Repo, Ref: ref.Name},
				State: &RefState{
					RefBaseInfo: RefBaseInfo{GitBase: refObj.gitBase, GitBranch: refObj.gitBranch},
					History:     refObj.history(),
				},
			}); err != nil {
				return err
			}
		}

		// Now send symbolic refs (see above for why we send them last).
		for _, ref := range refs {
			if !ref.IsSymbolic() {
				continue
			}

			log := log.With("update-symbolic-ref-with-initial-state", ref.Name)
			level.Debug(log).Log()
			if err := c.conn.Notify(ctx, "ref/updateSymbolic", RefUpdateSymbolicParams{
				RefIdentifier: RefIdentifier{Repo: params.Repo, Ref: ref.Name},
				Target:        ref.Target,
			}); err != nil {
				return err
			}
		}
	} else {
		level.Debug(log).Log("no-matching-refs", "")
	}

	return nil
}

func excludeSymbolicRefs(refs []refdb.Ref) []refdb.Ref {
	refs2 := make([]refdb.Ref, 0, len(refs))
	for _, ref := range refs {
		if !ref.IsSymbolic() {
			refs2 = append(refs2, ref)
		}
	}
	return refs2
}

func refsMatchingRefspecs(db refdb.RefDB, refspecs []string) []refdb.Ref {
	refs := map[string]refdb.Ref{}
	for _, refspec := range refspecs {
		for _, ref := range db.List(refspec) {
			refs[ref.Name] = ref
		}
	}

	refList := make([]refdb.Ref, 0, len(refs))
	for _, ref := range refs {
		refList = append(refList, ref)
	}
	sort.Sort(sortableRefs(refList))
	return refList
}

func matchAnyRefspec(refspecs []string, ref string) bool {
	for _, refspec := range refspecs {
		if refdb.MatchPattern(refspec, ref) {
			return true
		}
	}
	return false
}
