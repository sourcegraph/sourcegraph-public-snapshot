package zap

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/experimental_level"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap/server/refdb"
)

type serverRepo struct {
	refdb refdb.RefDB // the repo's refdb (safe for concurrent access)

	mu sync.Mutex
	// workspace:
	workspace       Workspace // set for non-bare repos added via workspace/add
	workspaceCancel func()    // tear down workspace
	config          RepoConfiguration
}

func (s *Server) getRepo(ctx context.Context, log *log.Context, repoName string) (*serverRepo, error) {
	repo, err := s.getRepoIfExists(ctx, log, repoName)
	if err != nil {
		return nil, err
	}
	if repo != nil {
		return repo, nil
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

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.watchingRepos == nil {
		c.watchingRepos = map[string]string{}
	}
	level.Info(log).Log("set-watch-refspec", params.Refspec, "old", c.watchingRepos[params.Repo])
	c.watchingRepos[params.Repo] = params.Refspec

	// Send over current state of all matched refs.
	if refs := repo.refdb.List(params.Refspec); len(refs) > 0 {
		for _, ref := range refs {
			refName := ref.Name // use orig (symbolic/non-resolved) ref name
			if ref.IsSymbolic() {
				ref2, _ := repo.refdb.Resolve(ref.Name)
				if ref2 == nil {
					continue
				}
				ref = *ref2
			}

			log := log.With("update-ref-downstream-with-initial-state", refName)
			level.Debug(log).Log()
			refObj := ref.Object.(serverRef)
			// TODO(sqs): make this a request so we make sure it is
			// received (to eliminate race conditions).
			if err := c.conn.Notify(ctx, "ref/update", RefUpdateDownstreamParams{
				RefIdentifier: RefIdentifier{Repo: params.Repo, Ref: refName},
				State: &RefState{
					RefBaseInfo: RefBaseInfo{GitBase: refObj.gitBase, GitBranch: refObj.gitBranch},
					History:     refObj.history(),
				},
			}); err != nil {
				return err
			}
		}
	} else {
		level.Warn(log).Log("no-matching-refs", "")
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
