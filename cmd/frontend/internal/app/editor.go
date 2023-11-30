package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/cloneurls"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func editorRev(ctx context.Context, logger log.Logger, db database.DB, repoName api.RepoName, rev string, beExplicit bool) string {
	if beExplicit {
		return "@" + rev
	}
	if rev == "HEAD" {
		return ""
	}
	repos := backend.NewRepos(logger, db, gitserver.NewClient("http.editorrev"))
	repo, err := repos.GetByName(ctx, repoName)
	if err != nil {
		// We weren't able to fetch the repo. This means it either doesn't
		// exist (unlikely) or that the user is not logged in (most likely). In
		// this case, the best user experience is to send them to the branch
		// they asked for. The front-end will inform them if the branch does
		// not exist.
		return "@" + rev
	}
	// If we are on the default branch we want to return a clean URL without a
	// branch. If we fail its best to return the full URL and allow the
	// front-end to inform them of anything that is wrong.
	defaultBranchCommitID, err := repos.ResolveRev(ctx, repo, "")
	if err != nil {
		return "@" + rev
	}
	branchCommitID, err := repos.ResolveRev(ctx, repo, rev)
	if err != nil {
		return "@" + rev
	}
	if defaultBranchCommitID == branchCommitID {
		return ""
	}
	return "@" + rev
}

// editorRequest represents the parameters to a Sourcegraph "open file", "search", etc. editor request.
type editorRequest struct {
	logger log.Logger
	db     database.DB

	// openFileRequest is non-nil if this is an "open file on Sourcegraph" request.
	openFileRequest *editorOpenFileRequest

	// searchRequest is non-nil if this is a "search on Sourcegraph" request.
	searchRequest *editorSearchRequest
}

// editorSearchRequest represents parameters for "open file on Sourcegraph" editor requests.
type editorOpenFileRequest struct {
	remoteURL         string            // Git repository remote URL.
	hostnameToPattern map[string]string // Map of Git remote URL hostnames to patterns describing how they map to Sourcegraph repositories
	branch            string            // Git branch name.
	revision          string            // Git revision.
	file              string            // Unix filepath relative to repository root.

	// Zero-based cursor selection parameters. Required.
	startRow, endRow int
	startCol, endCol int
}

// editorSearchRequest represents parameters for "search on Sourcegraph" editor requests.
type editorSearchRequest struct {
	query string // The literal search query

	// Optional git repository remote URL. When present, the search will be performed just
	// in the repository (not globally).
	remoteURL         string
	hostnameToPattern map[string]string // Map of Git remote URL hostnames to patterns describing how they map to Sourcegraph repositories

	// Optional git repository branch name and revision. When one is present and remoteURL
	// is present, the search will be performed just at this branch/revision.
	branch   string
	revision string

	// Optional unix filepath relative to the repository root. When present, the search
	// will be performed with a file: search filter.
	file string
}

// searchRedirect returns the redirect URL for the pre-validated search request.
func (r *editorRequest) searchRedirect(ctx context.Context) (string, error) {
	s := r.searchRequest

	// Handle searches scoped to a specific repository.
	var repoFilter string
	if s.remoteURL != "" {
		// Search in this repository.
		repoName, err := cloneurls.RepoSourceCloneURLToRepoName(ctx, r.db, s.remoteURL)
		if err != nil {
			return "", err
		}
		if repoName == "" {
			// Any error here is a problem with the user's configured git remote
			// URL. We want them to actually read this error message.
			return "", errors.Errorf("Git remote URL %q not supported", s.remoteURL)
		}
		// Note: we do not use ^ at the front of the repo filter because repoName may
		// produce imprecise results and a suffix match seems better than no match.
		repoFilter = "repo:" + regexp.QuoteMeta(string(repoName)) + "$"
	}

	// Handle searches scoped to a specific revision/branch.
	if repoFilter != "" && s.revision != "" {
		// Search in just this revision.
		repoFilter += "@" + s.revision
	} else if repoFilter != "" && s.branch != "" {
		// Search in just this branch.
		repoFilter += "@" + s.branch
	}

	// Handle searches scoped to a specific file.
	var fileFilter string
	if s.file != "" {
		fileFilter = "file:^" + regexp.QuoteMeta(s.file) + "$"
	}

	// Compose the final search query.
	parts := make([]string, 0, 3)
	for _, part := range []string{repoFilter, fileFilter, r.searchRequest.query} {
		if part != "" {
			parts = append(parts, part)
		}
	}
	searchQuery := strings.Join(parts, " ")

	// Build the redirect URL.
	u := &url.URL{Path: "/search"}
	q := u.Query()
	q.Add("q", searchQuery)
	q.Add("patternType", "literal")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// openFile returns the redirect URL for the pre-validated open-file request.
func (r *editorRequest) openFileRedirect(ctx context.Context) (string, error) {
	of := r.openFileRequest
	// Determine the repo name and branch.
	repoName, err := cloneurls.RepoSourceCloneURLToRepoName(ctx, r.db, of.remoteURL)
	if err != nil {
		return "", err
	}
	if repoName == "" {
		// Any error here is a problem with the user's configured git remote
		// URL. We want them to actually read this error message.
		return "", errors.Errorf("git remote URL %q not supported", of.remoteURL)
	}

	inputRev, beExplicit := of.revision, true
	if inputRev == "" {
		inputRev, beExplicit = of.branch, false
	}

	rev := editorRev(ctx, r.logger, r.db, repoName, inputRev, beExplicit)

	u := &url.URL{Path: path.Join("/", string(repoName)+rev, "/-/blob/", of.file)}
	q := u.Query()
	if of.startRow == of.endRow && of.startCol == of.endCol {
		q.Add(fmt.Sprintf("L%d", of.startRow+1), "")
	} else {
		q.Add(fmt.Sprintf("L%d:%d-%d:%d", of.startRow+1, of.startCol+1, of.endRow+1, of.endCol+1), "")
	}
	// Since the line information is added as the key as a query parameter with
	// an empty value, the URL encoding will add an = sign followed by an empty
	// string.
	//
	// Since we don't want the equal sign as it provides no value, we remove it.
	u.RawQuery = strings.TrimSuffix(q.Encode(), "=")
	return u.String(), nil
}

// openFile returns the redirect URL for the pre-validated request.
func (r *editorRequest) redirectURL(ctx context.Context) (string, error) {
	if r.searchRequest != nil {
		return r.searchRedirect(ctx)
	} else if r.openFileRequest != nil {
		return r.openFileRedirect(ctx)
	}
	return "", errors.New("could not determine request type, missing ?search or ?remote_url")
}

// parseEditorRequest parses an editor request from the search query values.
func parseEditorRequest(db database.DB, q url.Values) (*editorRequest, error) {
	if q == nil {
		return nil, errors.New("could not determine query string")
	}

	v := &editorRequest{
		db:     db,
		logger: log.Scoped("editor"),
	}

	if search := q.Get("search"); search != "" {
		// Search request parsing
		v.searchRequest = &editorSearchRequest{
			query:     q.Get("search"),
			remoteURL: q.Get("search_remote_url"),
			branch:    q.Get("search_branch"),
			revision:  q.Get("search_revision"),
			file:      q.Get("search_file"),
		}
		if hostnameToPatternStr := q.Get("search_hostname_patterns"); hostnameToPatternStr != "" {
			if err := json.Unmarshal([]byte(hostnameToPatternStr), &v.searchRequest.hostnameToPattern); err != nil {
				return nil, err
			}
		}
	} else if remoteURL := q.Get("remote_url"); remoteURL != "" {
		// Open-file request parsing
		startRow, _ := strconv.Atoi(q.Get("start_row"))
		endRow, _ := strconv.Atoi(q.Get("end_row"))
		startCol, _ := strconv.Atoi(q.Get("start_col"))
		endCol, _ := strconv.Atoi(q.Get("end_col"))
		v.openFileRequest = &editorOpenFileRequest{
			remoteURL: remoteURL,
			branch:    q.Get("branch"),
			revision:  q.Get("revision"),
			file:      q.Get("file"),
			startRow:  startRow,
			endRow:    endRow,
			startCol:  startCol,
			endCol:    endCol,
		}
		if hostnameToPatternStr := q.Get("hostname_patterns"); hostnameToPatternStr != "" {
			if err := json.Unmarshal([]byte(hostnameToPatternStr), &v.openFileRequest.hostnameToPattern); err != nil {
				return nil, err
			}
		}
	}
	return v, nil
}

func serveEditor(db database.DB) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		editorRequest, err := parseEditorRequest(db, r.URL.Query())
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", err.Error())
			return nil
		}
		redirectURL, err := editorRequest.redirectURL(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", err.Error())
			return nil
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return nil
	}
}
