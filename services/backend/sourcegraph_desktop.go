package backend

import (
	"context"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

var Desktop sourcegraph.DesktopServer = &desktop{}

type desktop struct{}

var _ sourcegraph.DesktopServer = (*desktop)(nil)

func (s *desktop) LatestExists(ctx context.Context, clientVersion *sourcegraph.ClientDesktopVersion) (*sourcegraph.LatestDesktopVersion, error) {
	version_number := clientVersion.ClientVersion
	res, err := http.Get(`https://storage.googleapis.com/sgreleasedesktop/releases/latest/` + version_number)

	latest := &sourcegraph.LatestDesktopVersion{}

	if res.Status == "200 OK" {
		latest.NewVersion = false
		return latest, err
	}

	latest.NewVersion = true
	return latest, err

}
