package httpapi

import (
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"

	srccli "github.com/sourcegraph/sourcegraph/internal/src-cli"
)

var srcCliDownloadsURL = "https://github.com/sourcegraph/src-cli/releases/download"

var allowedFilenames = []string{
	"src_darwin_amd64",
	"src_darwin_arm64",
	"src_linux_amd64",
	"src_linux_arm64",
	"src_windows_amd64.exe",
	"src_windows_arm64.exe",
}

func srcCliVersionServe(w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, &struct {
		Version string `json:"version"`
	}{
		Version: srcCliVersion(),
	})
}

func srcCliDownloadServe(w http.ResponseWriter, r *http.Request) error {
	filename := mux.Vars(r)["rest"]
	if !isExpectedRelease(filename) {
		http.NotFound(w, r)
		return nil
	}

	u, err := url.Parse(srcCliDownloadsURL)
	if err != nil {
		log15.Error("Illegal base src-cli download URL", "url", srcCliDownloadsURL, "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return nil
	}
	u.Path = path.Join(u.Path, srcCliVersion(), filename)
	http.Redirect(w, r, u.String(), http.StatusFound)
	return nil
}

func srcCliVersion() string {
	version, err := srccli.Version()
	if err != nil {
		// If we can't recommend a more specific version, just recommend the minimum version.
		// This is always safe, but may not include some newer features released via patch.
		// Use of the src-cli will warn users about an update once any transient error
		// resolves.
		log15.Warn("Failed to retrieve latest src-cli version", "err", err)
		return srccli.MinimumVersion
	}

	return version
}

func isExpectedRelease(filename string) bool {
	for _, v := range allowedFilenames {
		if filename == v {
			return true
		}
	}
	return false
}
