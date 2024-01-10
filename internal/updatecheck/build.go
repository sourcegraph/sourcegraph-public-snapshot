package updatecheck

import (
	"github.com/coreos/go-semver/semver"
)

// pingResponse is the JSON shape of the update check handler's response body.
type pingResponse struct {
	Version         semver.Version `json:"version"`
	UpdateAvailable bool           `json:"updateAvailable"`
}

func newPingResponse(version string) pingResponse {
	return pingResponse{
		Version: *semver.New(version),
	}
}
