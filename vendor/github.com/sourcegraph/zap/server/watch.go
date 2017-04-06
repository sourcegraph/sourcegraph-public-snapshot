package server

import (
	"context"
	"fmt"
	"sort"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/pkg/fpath"
	"github.com/sourcegraph/zap/server/refdb"
	"github.com/sourcegraph/zap/server/refstate"
	"github.com/sourcegraph/zap/server/repodb"
)

// isWatching returns whether c is watching the given ref (either via
// explicit ref/watch or by matching one of the refspec patterns
// provided to repo/watch).
//
// The caller must hold c.mu.
func (c *Conn) isWatching(ref zap.RefIdentifier) bool {
	if refspecs, ok := c.watchingRepos[fpath.Key(ref.Repo)]; ok {
		if matchAnyRefspec(refspecs, ref.Ref) {
			return true
		}
	}
	return false
}

func (c *Conn) handleRepoWatch(ctx context.Context, logger log.Logger, repo repodb.OwnedRepo, params zap.RepoWatchParams) error {
	if params.Repo == "" {
		return &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: "repo is required"}
	}

	for _, ref := range params.Refspecs {
		if ref != "*" {
			zap.CheckRefName(ref)
		}
	}

	if err := params.Validate(); err != nil {
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

	// Send over current state of all matched refs.
	//
	// From here on, clients will receive all future updates, so this
	// means they always have the full state of the repository.
	refs := refsMatchingRefspecs(repo.RefDB, params.Refspecs)
	if len(refs) > 0 {
		for _, ref := range refs {
			refState := ref.Data.(refstate.RefState).RefState
			level.Debug(logger).Log("initial-watch-of-ref", ref.Name, "state", refState)
			if err := c.Conn.Notify(ctx, "ref/update", zap.RefUpdateDownstreamParams{
				RefIdentifier: zap.RefIdentifier{Repo: params.Repo, Ref: ref.Name},
				State:         &refState,
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
