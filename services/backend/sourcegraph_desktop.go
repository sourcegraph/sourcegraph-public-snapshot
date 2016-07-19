package backend

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sqs/pbtypes"
)

var Desktop sourcegraph.DesktopServer = &desktop{}

type desktop struct{}

var _ sourcegraph.DesktopServer = (*desktop)(nil)

func (s *desktop) GetLatest(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.LatestDesktopVersion, error) {
	gh := githubutil.Default.UnauthedClient()

	latestRelease, _, err := gh.Repositories.GetLatestRelease("sourcegraph", "sourcegraph-desktop")
	if err != nil {
		return nil, err
	}

	latest := &sourcegraph.LatestDesktopVersion{
		Version: *latestRelease.TagName,
	}

	return latest, nil

}
