package app

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

func editorBranch(ctx context.Context, repoURI, branchName string) (string, error) {
	if branchName == "HEAD" {
		return "", nil // Detached head state
	}
	repo, err := backend.Repos.GetByURI(ctx, repoURI)
	if err != nil {
		// We weren't able to fetch the repo. This means it either doesn't
		// exist (unlikely) or that the user is not logged in (most likely). In
		// this case, the best user experience is to send them to the branch
		// they asked for. The front-end will inform them if the branch does
		// not exist.
		return "@" + branchName, nil
	}
	if branchName == repo.DefaultBranch {
		return "", nil // default branch, so make a clean URL without a branch.
	}
	return "@" + branchName, nil
}

func serveEditor(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	editor := q.Get("editor")                      // Editor name: "Atom", "Sublime", etc.
	version := q.Get("version")                    // Editor extension version.
	utmProductName := q.Get("utm_product_name")    // Editor product name, for JetBrains (e.g. "IntelliJ", "Gogland").
	utmProductVersion := q.Get("utm_product_name") // Editor product version, for JetBrains.

	// search query parameters. Only present if it is a search request.
	search := q.Get("search")

	// open-file parameters. Only present if it is a open-file request.
	remoteURL := q.Get("remote_url")                // Git repository remote URL.
	branch := q.Get("branch")                       // Git branch name.
	file := q.Get("file")                           // File relative to repository root.
	startRow, _ := strconv.Atoi(q.Get("start_row")) // zero-based
	startCol, _ := strconv.Atoi(q.Get("start_col")) // zero-based
	endRow, _ := strconv.Atoi(q.Get("end_row"))     // zero-based
	endCol, _ := strconv.Atoi(q.Get("end_col"))     // zero-based

	if search != "" {
		// Search request.
		u := &url.URL{Path: "/"}
		q := u.Query()
		q.Add("search", search)
		q.Add("utm_source", editor+"-"+version)
		if utmProductName != "" {
			q.Add("utm_product_name", utmProductName)
		}
		if utmProductVersion != "" {
			q.Add("utm_product_version", utmProductVersion)
		}
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusSeeOther)
		return nil
	}

	// Open-file request.
	repoURI, err := gitserver.DefaultClient.RepoFromRemoteURL(remoteURL)
	if err != nil {
		return err
	}
	if repoURI == "" {
		// Any error here is a problem with the user's configured git remote
		// URL. We want them to actually read this error message.
		msg := fmt.Sprintf("Git remote URL %q not supported", remoteURL)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, html.EscapeString(msg))
		return nil
	}
	branch, err = editorBranch(r.Context(), repoURI, branch)
	if err != nil {
		return err
	}
	u := &url.URL{Path: path.Join("/", repoURI+branch, "/-/blob/", file)}
	q = u.Query()
	q.Add("utm_source", editor+"-"+version)
	if utmProductName != "" {
		q.Add("utm_product_name", utmProductName)
	}
	if utmProductVersion != "" {
		q.Add("utm_product_version", utmProductVersion)
	}
	u.RawQuery = q.Encode()
	if startRow == endRow && startCol == endCol {
		u.Fragment = fmt.Sprintf("L%d:%d", startRow+1, startCol+1)
	} else {
		u.Fragment = fmt.Sprintf("L%d:%d-%d:%d", startRow+1, startCol+1, endRow+1, endCol+1)
	}
	http.Redirect(w, r, u.String(), http.StatusSeeOther)
	return nil
}
