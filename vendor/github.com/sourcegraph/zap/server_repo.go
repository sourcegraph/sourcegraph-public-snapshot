package zap

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap/pkg/fpath"
	"github.com/sourcegraph/zap/server/refdb"
)

type serverRepo struct {
	repoDir string
	refdb   *refdb.SyncRefDB // the repo's refdb (safe for concurrent access)

	mu sync.Mutex
	// workspace:
	workspace       Workspace // set for non-bare repos added via workspace/add
	workspaceCancel func()    // tear down workspace
	config          RepoConfiguration
}

func (s *Server) getRepo(ctx context.Context, logger log.Logger, repoDir string) (*serverRepo, error) {
	repo, err := s.getRepoIfExists(ctx, logger, repoDir)
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
			Message: fmt.Sprintf("repo not found: %s (add it with 'zap init')", repoDir),
		}
	}

	s.reposMu.Lock()
	defer s.reposMu.Unlock()
	repo, exists := s.repos[fpath.Key(repoDir)]
	if !exists {
		repo = &serverRepo{
			repoDir: repoDir,
			refdb:   refdb.Sync(refdb.NewMemoryRefDB()),
		}
		s.repos[fpath.Key(repoDir)] = repo
	}
	return repo, nil
}

func (s *Server) getRepoIfExists(ctx context.Context, logger log.Logger, repoDir string) (*serverRepo, error) {
	ok, err := s.backend.CanAccess(ctx, logger, repoDir)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, &jsonrpc2.Error{
			Code:    int64(ErrorCodeRepoNotExists),
			Message: fmt.Sprintf("access forbidden to repo: %s", repoDir),
		}
	}

	s.reposMu.Lock()
	defer s.reposMu.Unlock()
	return s.repos[fpath.Key(repoDir)], nil
}

// resolveRefShortName resolves a ref fuzzy name to the ref it refers
// to (and if symbolic ref, its target ref as well). For example, a
// fuzzy name of "foo" would resolve to "branch/foo" (assuming no ref
// exists whose full name is "foo").
//
// The behavior is identical to (*refdb.SyncRefDB).Resolve, except
// that the primary ref is looked up fuzzily.
//
// Only the "ref/info" method should resolve fuzzy names; other
// methods should require full ref names to avoid ambiguity.
func (s *serverRepo) resolveRefByFuzzyName(fuzzy string) (ref refdb.OwnedRef, target *refdb.OwnedRef) {
	ref = s.refdb.Lookup(fuzzy)
	if ref.Ref == nil {
		ref.Unlock()
		ref = s.refdb.Lookup("branch/" + fuzzy)
	}
	if ref.Ref != nil {
		CheckRefName(ref.Ref.Name)

		// Resolve target.
		if ref.Ref.Target != "" && ref.Ref.Target != ref.Ref.Name {
			targetRef := s.refdb.Lookup(ref.Ref.Target)
			target = &targetRef
		}
	}
	return ref, target
}

func (c *serverConn) handleRepoWatch(ctx context.Context, logger log.Logger, repo *serverRepo, params RepoWatchParams) error {
	if params.Repo == "" {
		return &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: "repo is required"}
	}

	for _, ref := range params.Refspecs {
		if ref != "*" {
			CheckRefName(ref)
		}
	}

	if err := params.validate(); err != nil {
		return err
	}

	{
		c.mu.Lock()
		if c.watchingRepos == nil {
			c.watchingRepos = map[fpath.KeyString][]string{}
		}
		level.Info(logger).Log("set-watch-refspec", fmt.Sprint(params.Refspecs), "old", fmt.Sprint(c.watchingRepos[fpath.Key(params.Repo)]))
		c.watchingRepos[fpath.Key(params.Repo)] = params.Refspecs
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

			logger := log.With(logger, "update-ref-downstream-with-initial-state", ref.Name)
			level.Debug(logger).Log()
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

			logger := log.With(logger, "update-symbolic-ref-with-initial-state", ref.Name)
			level.Debug(logger).Log()
			if err := c.conn.Notify(ctx, "ref/updateSymbolic", RefUpdateSymbolicParams{
				RefIdentifier: RefIdentifier{Repo: params.Repo, Ref: ref.Name},
				Target:        ref.Target,
			}); err != nil {
				return err
			}
		}
	} else {
		level.Debug(logger).Log("no-matching-refs", "")
	}

	return nil
}

func refsMatchingRefspecs(db *refdb.SyncRefDB, refspecs []string) []refdb.Ref {
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
