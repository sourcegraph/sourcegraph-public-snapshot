package httpapi

import (
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sqs/pbtypes"
)

const downloadLink = "https://github.com/sourcegraph/sourcegraph-desktop/releases/download/"
const zips = "/Sourcegraph.zip"

func serveSourcegraphDesktopUpdateURL(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	clientVersion := r.Header.Get("Sourcegraph-Version")

	latestVersion, err := cl.Desktop.GetLatest(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	if latestVersion.Version == clientVersion {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	url := map[string]string{
		"url": strings.Join([]string{downloadLink, latestVersion.Version, zips}, ""),
	}
	return writeJSON(w, url)
}
