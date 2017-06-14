package bundle

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var mu sync.Mutex
var cache = make(map[string]*asset)

type asset struct {
	status      int
	contentType string
	body        []byte
}

var browserPkg = env.Get("VSCODE_BROWSER_PKG", "", "load vscode assets from disk")

func Handler() http.Handler {
	if browserPkg != "" {
		Version = "dev"
		return http.StripPrefix("/"+Version, http.FileServer(http.Dir(browserPkg)))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a, err := fetch(r.URL.Path)
		if err != nil {
			log.Print(err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if a.status == 200 {
			w.Header().Set("Content-Type", a.contentType)
			w.Header().Set("Cache-Control", "max-age=31536000, public")
		}
		w.WriteHeader(a.status)
		w.Write(a.body)
	})
}

func fetch(path string) (*asset, error) {
	mu.Lock()
	a, ok := cache[path]
	mu.Unlock()
	if !ok {
		resp, err := http.Get("https://app.sourcegraph.com" + path)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		mu.Lock()
		a = &asset{resp.StatusCode, resp.Header.Get("Content-Type"), body}
		cache[path] = a
		mu.Unlock()
	}

	return a, nil
}
