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
	spec     string
	t        time.Time
}

func (g *gitCLIBackend) RevAtTime(ctx context.Context, spec string, t time.Time) (api.CommitID, error) {
	key := revAtTimeCacheKey{g.repoName, spec, t}
	if entry, ok := g.revAtTimeCache.Get(key); ok {
		return entry, nil
	}

	r, err := g.NewCommand(ctx, WithArguments(
		"log",
		"--format=format:%H", // only hash
		"--first-parent",     // children before parents, but otherwise sort by date
		fmt.Sprintf("--before=%d", t.Unix()),
		"--max-count=1", // only one commit
		spec,
	))
	if err != nil {
		return "", err
	}

	stdout, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	commitID := api.CommitID(bytes.TrimSpace(stdout))
	g.revAtTimeCache.Add(key, commitID)
	return commitID, nil
}
