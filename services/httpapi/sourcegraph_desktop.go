package httpapi

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

const downloadLink = "https://storage.googleapis.com/sgreleasedesktop/releases/latest/Sourcegraph.zip"
const zips = "/Sourcegraph.zip"

func serveSourcegraphDesktopUpdateURL(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	clientVersion := &sourcegraph.ClientDesktopVersion{
		ClientVersion: r.Header.Get("Sourcegraph-Version"),
	}

	res, err := cl.Desktop.LatestExists(r.Context(), clientVersion)
	if err != nil {
		return err
	}

	if res.NewVersion == false {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	url := map[string]string{
		"url": downloadLink,
	}
	return writeJSON(w, url)
}
