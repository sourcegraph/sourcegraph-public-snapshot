package httpapi

import "net/http"

func serveSourcegraphDesktopUpdateURL(w http.ResponseWriter, r *http.Request) error {
	res, err := http.Get("https://storage.googleapis.com/sgreleasedesktop/releases/latest/" + r.Header.Get("Sourcegraph-Version"))
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	url := map[string]string{
		"url": "https://storage.googleapis.com/sgreleasedesktop/releases/latest/Sourcegraph.zip",
	}
	return writeJSON(w, url)
}
