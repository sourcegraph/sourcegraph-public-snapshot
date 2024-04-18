package gitcli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/hashicorp/golang-lru/v2"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// A simple in-process cache to mitigate the cost of any regular
// rev:at.time() queries.
//
// Each cache entry takes ~128 bytes, so 8192 entries should keep this cache
// around 1MB.
var globalRevAtTimeCache, _ = lru.New[revAtTimeCacheKey, api.CommitID](8192)

type revAtTimeCacheKey struct {
	repoName api.RepoName
	sha      api.CommitID
	t        time.Time
}

func (g *gitCLIBackend) RevAtTime(ctx context.Context, spec string, t time.Time) (api.CommitID, error) {
	// First, try to resolve the revspec so we can return a useful RevisionNotFound error
	sha, err := g.ResolveRevision(ctx, spec)
	if err != nil {
		return "", err
	}

	key := revAtTimeCacheKey{g.repoName, sha, t}
	if entry, ok := g.revAtTimeCache.Get(key); ok {
		return entry, nil
	}

	r, err := g.NewCommand(ctx, WithArguments(
		"log",
		"--format=format:%H", // only hash
		"--first-parent",     // linearize history
		fmt.Sprintf("--before=%d", t.Unix()),
		"--max-count=1", // only one commit
		string(sha),
	))
	if err != nil {
		return "", err
	}
	defer r.Close()

	stdout, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	commitID := api.CommitID(bytes.TrimSpace(stdout))
	g.revAtTimeCache.Add(key, commitID)
	return commitID, nil
}
