package apidocs

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"

	approuter "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func init() {
	platform.RegisterFrame(platform.RepoFrame{
		ID:      "apidocs",
		Title:   "API Docs",
		Icon:    "rocket",
		Handler: handlerWithError(handler),
	})
}

// handleWithError takes a custom handler that may return an error and returns
// a valid `http.Handler`. If an error is returned, it is captured and handled.
func handlerWithError(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func handler(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)

	repoRevSpec, exists := pctx.RepoRevSpec(ctx)
	if !exists {
		return errors.New("could not parse repository spec from URL")
	}
	r.URL.Path = strings.TrimPrefix(path.Clean(r.URL.Path), "/")

	// Handle RepoTree referer.
	err, handled := handleReferer(w, r, repoRevSpec)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	// Right now we require that the path be to a directory, but for the root dir
	// the path will be empty. Handle this now.
	requestDir := r.URL.Path
	if requestDir == "" {
		requestDir = "."
	}
	fullPath := path.Join(repoRevSpec.URI, r.URL.Path)

	// Grab all the definitions for the requested directory.
	defs, err := defsForDir(ctx, repoRevSpec, requestDir)
	if err != nil {
		return err
	}

	// Find the subdirectories for the requested directory.
	subDirs, err := subDirsForDir(ctx, repoRevSpec, requestDir)
	if err != nil {
		return err
	}

	// urlToDef generates a URL to a definition at whichever revision the user is
	// browsing at.
	urlToDef := func(def *sourcegraph.Def) string {
		rev := repoRevSpec.Rev
		if rev == "" {
			rev = repoRevSpec.CommitID
		}
		if rev == "" {
			return approuter.Rel.URLToDef(def.DefKey).String()
		}
		return approuter.Rel.URLToDefAtRev(def.DefKey, rev).String()
	}

	return executeTemplate(w, r, "home.html", &struct {
		TmplCommon
		Defs     []*sourcegraph.Def
		FullPath string
		URL      *url.URL
		SubDirs  []string
		URLToDef func(d *sourcegraph.Def) string
	}{
		Defs:     defs,
		FullPath: fullPath,
		URL:      r.URL,
		SubDirs:  subDirs,
		URLToDef: urlToDef,
	})
}

// handleReferer checks the referer header for a RepoTree path (i.e. that checks
// if the user came from the code browser). If they were, then the request is
// handled by redirecting the user to the right page for apidocs in the same dir
// that they were viewing.
func handleReferer(w http.ResponseWriter, r *http.Request, repoRevSpec sourcegraph.RepoRevSpec) (error, bool) {
	// Avoid constantly redirecting, as referer header will not be updated on our
	// redirect.
	if r.URL.Query().Get("redirect") == "true" {
		return nil, false
	}

	// Check for the referer header.
	ref := r.Header.Get("Referer")
	if ref == "" {
		return nil, false
	}

	// Parse referer header.
	u, err := url.Parse(ref)
	if err != nil {
		return err, false
	}

	// HACK: parse the revision out of the URL path. Platform apps should be
	// able to "opt-in" to keeping the @revision specifier on repo frame tab
	// links.
	nameAtRev := strings.Split(strings.TrimPrefix(u.Path, "/"+repoRevSpec.URI), "/")[0]
	rev := ""
	if split := strings.Split(nameAtRev, "@"); len(split) == 2 {
		rev = split[1]
	}

	treePrefix := approuter.Rel.URLToRepoTreeEntry(repoRevSpec.URI, rev, "").Path
	if !strings.HasPrefix(u.Path, treePrefix) {
		// User didn't come from a RepoTree page.
		return nil, false
	}

	// HACK: we want to keep our revision!
	if split := strings.Split(u.Path, "/"); path.Ext(split[len(split)-1]) != "" {
		// use only dir if they were viewing a file.
		u.Path = strings.Join(split[:len(split)-1], "/")
	}
	u.Path = strings.Replace(u.Path, "/.tree/", "/.apidocs/", 1)
	u.RawQuery = "redirect=true"

	// Redirect and passthrough HTTP response.
	w.Header().Set(platform.HTTPHeaderVerbatim, "true")
	http.Redirect(w, r, u.String(), http.StatusSeeOther)
	return nil, true
}
