package httpapi

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var srcCliDownloadsURL = "https://github.com/sourcegraph/src-cli/releases/download"

var whitelistedFilenames = []string{
	"src_darwin_amd64",
	"src_linux_amd64",
	"src_windows_amd64.exe",
}

func srcCliServe(w http.ResponseWriter, r *http.Request) error {
	filename := mux.Vars(r)["rest"]
	if !isExpectedRelease(filename) {
		http.NotFound(w, r)
		return nil
	}

	target := fmt.Sprintf("%s/%s/%s", srcCliDownloadsURL, SrcCliVersion, filename)
	http.Redirect(w, r, target, http.StatusFound)
	return nil
}

func isExpectedRelease(filename string) bool {
	for _, v := range whitelistedFilenames {
		if filename == v {
			return true
		}
	}
	return false
}
