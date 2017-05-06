package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func remoteURLToRepoURI(ctx context.Context, remoteURL string) (string, error) {
	cpy := remoteURL
	schemeIndex := strings.Index(cpy, "://")
	if schemeIndex == -1 {
		cpy = "git://" + cpy
	}
	u, err := url.Parse(cpy)
	if err != nil {
		return "", err
	}
	if u.Hostname() != "github.com" {
		// TODO: Handle on-premises here. Consider e.g. ORIGIN_MAP and env.example (file).
		return "", fmt.Errorf("Git remote URL %q not supported.", remoteURL)
	}

	// GitHub remotes
	u.Path = strings.TrimSuffix(u.Path, ".git")
	if strings.Contains(u.Host, ":") {
		return path.Join(u.Hostname(), strings.Split(u.Host, ":")[1], u.Path), nil
	}
	return path.Join(u.Hostname(), u.Path), nil
}

func editorBranch(ctx context.Context, repoURI, branchName string) (string, error) {
	if branchName == "HEAD" {
		return "", nil // Detached head state
	}
	repo, err := backend.Repos.GetByURI(ctx, repoURI)
	if err != nil {
		return "", err
	}
	if branchName == repo.DefaultBranch {
		return "", nil // default branch, so make a clean URL without a branch.
	}
	_, err = backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: repo.ID,
		Rev:  branchName,
	})
	if err == vcs.ErrRevisionNotFound {
		// Branch does not exist. The user probably didn't push it to the
		// remote. Using the default branch is the best hope.
		return "", nil
	}
	return "@" + branchName, nil
}

func serveEditor(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	editor := q.Get("editor")   // Editor name: "Atom", "Sublime", etc.
	version := q.Get("version") // Editor extension version.

	// search query parameters. Only present if it is a search request.
	search := q.Get("search")

	// open-file parameters. Only present if it is a open-file request.
	remoteURL := q.Get("remote_url")                // Git repository remote URL.
	branch := q.Get("branch")                       // Git branch name.
	file := q.Get("file")                           // File relative to repository root.
	startRow, _ := strconv.Atoi(q.Get("start_row")) // one-based
	startCol, _ := strconv.Atoi(q.Get("start_col")) // one-based
	endRow, _ := strconv.Atoi(q.Get("end_row"))     // one-based
	endCol, _ := strconv.Atoi(q.Get("end_col"))     // one-based

	if search != "" {
		// Search request.
		u := &url.URL{Path: "/search"}
		q := u.Query()
		q.Add("q", search)
		q.Add("utm_source", editor+"-"+version)
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusSeeOther)
		return nil
	}

	// Open-file request.
	repoURI, err := remoteURLToRepoURI(r.Context(), remoteURL)
	if err != nil {
		return err
	}
	branch, err = editorBranch(r.Context(), repoURI, branch)
	if err != nil {
		return err
	}
	u := &url.URL{Path: path.Join("/", repoURI, branch, "/-/blob/", file)}
	q = u.Query()
	q.Add("utm_source", editor+"-"+version)
	u.RawQuery = q.Encode()
	if startRow == endRow && startCol == endCol {
		u.Fragment = fmt.Sprintf("L%d-%d", startRow+1, startCol+1)
	} else {
		u.Fragment = fmt.Sprintf("L%d-%d:%d-%d", startRow+1, startCol+1, endRow+1, endCol+1)
	}
	http.Redirect(w, r, u.String(), http.StatusSeeOther)
	return nil
}
