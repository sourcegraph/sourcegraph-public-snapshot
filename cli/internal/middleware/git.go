package middleware

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// GitHubCloneProxy proxies git clones for GitHub hosted repositories
func GitHubCloneProxy(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if ua := r.UserAgent(); !strings.HasPrefix(ua, "git/") && !strings.HasPrefix(ua, "JGit/") {
		next(w, r)
		return
	}

	// handle `git clone`
	h := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "https", Host: "github.com", Path: "/"})
	origDirector := h.Director
	h.Director = func(r *http.Request) {
		origDirector(r)
		r.Host = "github.com"
		if strings.HasPrefix(r.URL.Path, "/github.com/") {
			r.URL.Path = r.URL.Path[len("/github.com"):]
		}
	}
	h.ServeHTTP(w, r)
}
