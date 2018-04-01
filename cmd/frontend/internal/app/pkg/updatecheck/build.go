package updatecheck

import "github.com/coreos/go-semver/semver"

// build is the JSON shape of the update check handler's response body.
type build struct {
	Version semver.Version `json:"version"`

	// IsReleased is not used, except old clients check to see that it is true.
	// We could remove this after a while.
	IsReleased bool `json:"isReleased"`

	// This field is not used, except old clients check to see that it is non-empty.
	// We could remove this after a while.
	Assets []struct{} `json:"assets"`
}

func newBuild(version string) build {
	return build{
		Version:    *semver.New(version),
		IsReleased: true,
		Assets: []struct{}{
			{},
		},
	}
}
