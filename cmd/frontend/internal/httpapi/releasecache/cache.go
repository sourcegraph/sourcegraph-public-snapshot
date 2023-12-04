package releasecache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ReleaseCache provides a cache of the latest release of each branch of a
// specific GitHub repository.
type ReleaseCache interface {
	goroutine.Handler
	Current(branch string) (string, error)
	UpdateNow(ctx context.Context) error
}

type releaseCache struct {
	logger log.Logger

	// The repository to query, along with a client for the right GitHub host.
	client *github.V4Client
	owner  string
	name   string

	// The actual cache of branches and their current release versions.
	mu       sync.RWMutex
	branches map[string]string
}

func newReleaseCache(logger log.Logger, client *github.V4Client, owner, name string) ReleaseCache {
	return &releaseCache{
		client:   client,
		logger:   logger.Scoped("ReleaseCache"),
		branches: map[string]string{},
		owner:    owner,
		name:     name,
	}
}

func (rc *releaseCache) Current(branch string) (string, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	if version, ok := rc.branches[branch]; ok {
		return version, nil
	}
	return "", branchNotFoundError(branch)
}

// Handle implements goroutine.Handler, and updates the release cache each time
// it is invoked.
func (rc *releaseCache) Handle(ctx context.Context) error {
	rc.logger.Debug("handling request to update the release cache")
	err := rc.fetch(ctx)
	if err != nil {
		rc.logger.Error("error updating the release cache", log.Error(err))
	}

	return err
}

func (rc *releaseCache) UpdateNow(ctx context.Context) error {
	return rc.fetch(ctx)
}

func (rc *releaseCache) fetch(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	branches := map[string]string{}
	params := github.ReleasesParams{
		Name:  rc.name,
		Owner: rc.owner,
	}
	// The releases query is paginated, so we'll iterate until we run out of
	// pages. This isn't terribly efficient — practically, most branches will
	// never see an update — but it's the simplest way to ensure we have
	// everything up to date, and we're not going to do this very often anyway.
	for {
		releases, err := rc.client.Releases(ctx, &params)
		if err != nil {
			return errors.Wrap(err, "getting releases")
		}

		processReleases(rc.logger, branches, releases.Nodes)

		if !releases.PageInfo.HasNextPage {
			break
		}
		params.After = releases.PageInfo.EndCursor
	}

	// Actually update the release cache.
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.branches = branches
	return nil
}

func processReleases(logger log.Logger, branches map[string]string, releases []github.Release) {
	for _, release := range releases {
		if release.IsDraft || release.IsPrerelease {
			continue
		}

		version, err := semver.NewVersion(release.TagName)
		if err != nil {
			logger.Debug("ignoring malformed release", log.Error(err), log.String("TagName", release.TagName))
			continue
		}

		// Since V4Client.Releases always returns the releases in descending
		// release order, we don't have to do any version comparisons: we can
		// simply use the first release on the branch only and ignore the rest.
		branch := fmt.Sprintf("%d.%d", version.Major, version.Minor)
		if _, found := branches[branch]; !found {
			branches[branch] = release.TagName
		}
	}
}

type branchNotFoundError string

func (e branchNotFoundError) Error() string {
	return "branch not found: " + string(e)
}

func (e branchNotFoundError) NotFound() bool { return true }
