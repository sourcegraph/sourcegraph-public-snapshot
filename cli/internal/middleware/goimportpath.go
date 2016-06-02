package middleware

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
)

// goImportMetaTag represents a go-import meta tag.
type goImportMetaTag struct {
	// ImportPrefix is the import path corresponding to the repository root.
	// It must be a prefix or an exact match of the package being fetched with "go get".
	// If it's not an exact match, another http request is made at the prefix to verify
	// the <meta> tags match.
	ImportPrefix string

	// VCS is one of "git", "hg", "svn", etc.
	VCS string

	// RepoRoot is the root of the version control system containing a scheme and
	// not containing a .vcs qualifier.
	RepoRoot string
}

// goImportMetaTagTemplate is an HTML template for rendering a blank page with a go-import meta tag.
var goImportMetaTagTemplate = template.Must(template.New("").Parse(`<html><head><meta name="go-import" content="{{.ImportPrefix}} {{.VCS}} {{.RepoRoot}}"></head><body></body></html>`))

// SourcegraphComGoGetHandler is middleware for serving go-import meta tags for requests with ?go-get=1 query
// on sourcegraph.com.
//
// It implements the following mapping:
//
// 1. If the request is an existing hosted repository, it is served directly, and its clone URL is the import path.
// 2. Otherwise, if the username (first path element) is "sourcegraph", consider it to be a vanity
//    import path pointing to github.com/sourcegraph/<repo> as the clone URL.
// 3. All other requests are served with 404 Not Found.
func SourcegraphComGoGetHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("go-get") != "1" {
			next.ServeHTTP(w, req)
			return
		}

		httpctx.SetRouteName(req, "go-get")
		if !strings.HasPrefix(req.URL.Path, "/") {
			err := fmt.Errorf("req.URL.Path doesn't have a leading /: %q", req.URL.Path)
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := httpctx.FromRequest(req)
		cl, err := sourcegraph.NewClientFromContext(ctx)
		if err != nil {
			log.Println("SourcegraphComGoGetHandler: sourcegraph.NewClientFromContext:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		pathElements := strings.Split(req.URL.Path[1:], "/")

		// Check if the requested path or its prefix is a hosted repository.
		//
		// If there are 3 path elements, e.g., "/alpha/beta/gamma", start by checking
		// repo path "alpha", then "alpha/beta", and finally "alpha/beta/gamma".
		for i := 1; i <= len(pathElements); i++ {
			repoPath := strings.Join(pathElements[:i], "/")

			repo, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{
				URI: repoPath,
			})
			if err == nil && repo.Mirror {
				continue
			} else if errcode.HTTP(err) == http.StatusNotFound {
				continue
			} else if err != nil {
				// TODO: Distinguish between other known/expected errors vs unexpected errors,
				//       and treat unexpected errors appropriately. Doing this requires Repos.Get
				//       method to be documented to specify which known error types it can return.
				log.Println("SourcegraphComGoGetHandler: cl.Repos.Get:", err)
				http.Error(w, "error getting repository", http.StatusInternalServerError)
				return
			}

			// Repo found. Serve a go-import meta tag.

			appURL := conf.AppURL(ctx)
			scheme := appURL.Scheme
			host := appURL.Host

			goImportMetaTagTemplate.Execute(w, goImportMetaTag{
				ImportPrefix: path.Join(host, repoPath),
				VCS:          "git",
				RepoRoot:     scheme + "://" + host + "/" + repoPath,
			})
			if err != nil {
				log.Println("goImportMetaTagTemplate.Execute:", err)
			}
			return
		}

		// Handle "go get sourcegraph.com/{sourcegraph,sqs}/*" for all non-hosted repositories.
		// It's a vanity import path that maps to "github.com/{sourcegraph,sqs}/*" clone URLs.
		if len(pathElements) >= 2 && (pathElements[0] == "sourcegraph" || pathElements[0] == "sqs") {
			host := conf.AppURL(ctx).Host

			user := pathElements[0]
			repo := pathElements[1]

			err := goImportMetaTagTemplate.Execute(w, goImportMetaTag{
				ImportPrefix: path.Join(host, user, repo),
				VCS:          "git",
				RepoRoot:     "https://github.com/" + user + "/" + repo,
			})
			if err != nil {
				log.Println("goImportMetaTagTemplate.Execute:", err)
			}
			return
		}

		// If we get here, there isn't a Go package for this request.
		http.Error(w, "no such repository", http.StatusNotFound)
		return
	})
}
