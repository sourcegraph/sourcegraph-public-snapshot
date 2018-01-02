package updatecheck

// build is the JSON shape of the update check handler's response body.
//
// This is inspired by the Build type returned by the sourcegraph-distro update
// server, defined at
// https://github.com/sourcegraph/sourcegraph-distro/blob/master/pkg/distroutil/build.go#L14.
type build struct {
	Timestamp  int64   `json:"timestamp"`
	Version    string  `json:"version"`
	IsReleased bool    `json:"isReleased"`
	Assets     []asset `json:"assets"`
}

type asset struct {
	Name           string `json:"name"`
	ReleaseNotes   string `json:"releaseNotes,omitempty"`
	Version        string `json:"version"`
	ProductVersion string `json:"productVersion"`
	Platform       string `json:"platform"`
	Type           string `json:"type"`
	URL            string `json:"url,omitempty"`
}
