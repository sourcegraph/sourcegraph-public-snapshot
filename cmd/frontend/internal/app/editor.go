package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

func editorRev(ctx context.Context, repoName api.RepoName, rev string, beExplicit bool) (string, error) {
	if beExplicit {
		return "@" + rev, nil
	}
	if rev == "HEAD" {
		return "", nil // Detached head state
	}
	repo, err := backend.Repos.GetByName(ctx, repoName)
	if err != nil {
		// We weren't able to fetch the repo. This means it either doesn't
		// exist (unlikely) or that the user is not logged in (most likely). In
		// this case, the best user experience is to send them to the branch
		// they asked for. The front-end will inform them if the branch does
		// not exist.
		return "@" + rev, nil
	}
	// If we are on the default branch we want to return a clean URL without a
	// branch. If we fail its best to return the full URL and allow the
	// front-end to inform them of anything that is wrong.
	defaultBranchCommitID, err := backend.Repos.ResolveRev(ctx, repo, "")
	if err != nil {
		return "@" + rev, nil
	}
	branchCommitID, err := backend.Repos.ResolveRev(ctx, repo, rev)
	if err != nil {
		return "@" + rev, nil
	}
	if defaultBranchCommitID == branchCommitID {
		return "", nil // default branch, so make a clean URL without a branch.
	}
	return "@" + rev, nil
}

// editorRequest represents the parameters to a Sourcegraph "open file", "search", etc. editor request.
type editorRequest struct {
	// Fields that are required in all requests.
	editor  string // editor name, e.g. "Atom", "Sublime", etc.
	version string // editor extension version

	// Fields that are optional in all requests.
	utmProductName    string // Editor product name. Only present in JetBrains today (e.g. "IntelliJ", "Gogland")
	utmProductVersion string // Editor product version. Only present in JetBrains today.

	// openFileRequest is non-nil if this is an "open file on Sourcegraph" request.
	openFileRequest *editorOpenFileRequest

	// searchRequest is non-nil if this is a "search on Sourcegraph" request.
	searchRequest *editorSearchRequest
}

// editorSearchRequest represents parameters for "open file on Sourcegraph" editor requests.
type editorOpenFileRequest struct {
	remoteURL         string            // Git repository remote URL.
	branch            string            // Git branch name.
	revision          string            // Git revision.
	file              string            // File relative to repository root.
	hostnameToPattern map[string]string // map of Git remote URL hostnames to patterns describing how they map to Sourcegraph repositories

	// Zero-based cursor selection parameters. Required.
	startRow, endRow int
	startCol, endCol int
}

// editorSearchRequest represents parameters for "search on Sourcegraph" editor requests.
type editorSearchRequest struct {
	query string // The literal search query
}

// searchRedirect returns the redirect URL for the pre-validated search request.
func (r *editorRequest) searchRedirect() string {
	// Search request. The search is intentionally not scoped to a repository, because it's assumed the
	// user prefers to perform the search in their last-used search scope. Searching in their current
	// repo is not actually very useful, since they can usually do that better in their editor.
	u := &url.URL{Path: "/search"}
	q := u.Query()
	q.Add("q", r.searchRequest.query)
	q.Add("patternType", "literal")
	q.Add("utm_source", r.editor+"-"+r.version)
	if r.utmProductName != "" {
		q.Add("utm_product_name", r.utmProductName)
	}
	if r.utmProductVersion != "" {
		q.Add("utm_product_version", r.utmProductVersion)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// openFile returns the redirect URL for the pre-validated open-file request.
func (r *editorRequest) openFileRedirect(ctx context.Context) (string, error) {
	of := r.openFileRequest
	// Determine the repo name and branch.
	//
	// TODO(sqs): This used to hit gitserver, which would be more accurate in case of nonstandard
	// clone URLs.  It now generates the guessed repo name statically, which means in some cases it
	// won't work, but it is worth the increase in simplicity (plus there is an error message for
	// users). In the future we can let users specify a custom mapping to the Sourcegraph repo in
	// their local Git repo (instead of having them pass it here).
	repoName := guessRepoNameFromRemoteURL(of.remoteURL, of.hostnameToPattern)
	if repoName == "" {
		// Any error here is a problem with the user's configured git remote
		// URL. We want them to actually read this error message.
		return "", fmt.Errorf("Git remote URL %q not supported", of.remoteURL)
	}

	inputRev, beExplicit := of.revision, true
	if inputRev == "" {
		inputRev, beExplicit = of.branch, false
	}
	rev, err := editorRev(ctx, repoName, inputRev, beExplicit)
	if err != nil {
		return "", err
	}

	u := &url.URL{Path: path.Join("/", string(repoName)+rev, "/-/blob/", of.file)}
	q := u.Query()
	q.Add("utm_source", r.editor+"-"+r.version)
	if r.utmProductName != "" {
		q.Add("utm_product_name", r.utmProductName)
	}
	if r.utmProductVersion != "" {
		q.Add("utm_product_version", r.utmProductVersion)
	}
	u.RawQuery = q.Encode()
	if of.startRow == of.endRow && of.startCol == of.endCol {
		u.Fragment = fmt.Sprintf("L%d:%d", of.startRow+1, of.startCol+1)
	} else {
		u.Fragment = fmt.Sprintf("L%d:%d-%d:%d", of.startRow+1, of.startCol+1, of.endRow+1, of.endCol+1)
	}
	return u.String(), nil
}

// openFile returns the redirect URL for the pre-validated request.
func (r *editorRequest) redirectURL(ctx context.Context) (string, error) {
	if r.searchRequest != nil {
		return r.searchRedirect(), nil
	} else if r.openFileRequest != nil {
		return r.openFileRedirect(ctx)
	}
	return "", errors.New("could not determine request type, missing ?search or ?remote_url")
}

// parseEditorRequest parses an editor request from the search query values.
func parseEditorRequest(q url.Values) (*editorRequest, error) {
	v := &editorRequest{
		editor:            q.Get("editor"),
		version:           q.Get("version"),
		utmProductName:    q.Get("utm_product_name"),
		utmProductVersion: q.Get("utm_product_name"),
	}
	if v.editor == "" {
		return nil, fmt.Errorf("expected URL parameter missing: editor=$EDITOR_NAME")
	}
	if v.version == "" {
		return nil, fmt.Errorf("expected URL parameter missing: version=$EDITOR_EXTENSION_VERSION")
	}

	if search := q.Get("search"); search != "" {
		// Search request parsing
		v.searchRequest = &editorSearchRequest{
			query: q.Get("search"),
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

func serveEditor(w http.ResponseWriter, r *http.Request) error {
	editorRequest, err := parseEditorRequest(r.URL.Query())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err.Error())
	}
	redirectURL, err := editorRequest.redirectURL(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err.Error())
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	return nil
}

// gitProtocolRegExp is a regular expression that matches any URL that looks like it has a git protocol
var gitProtocolRegExp = lazyregexp.New("^(git|(git+)?(https?|ssh))://")

// guessRepoNameFromRemoteURL return a guess at the repo name for the given remote URL.
//
// It first normalizes the remote URL (ensuring a scheme exists, stripping any "git@" username in
// the host, stripping any trailing ".git" from the path, etc.). It then returns the repo name as
// templatized by the pattern specified, which references the hostname and path of the normalized
// URL. Patterns are keyed by hostname in the hostnameToPattern parameter. The default pattern is
// "{hostname}/{path}".
//
// For example, given "https://github.com/foo/bar.git" and an empty hostnameToPattern, it returns
// "github.com/foo/bar". Given the same remote URL and hostnametoPattern
// `map[string]string{"github.com": "{path}"}`, it returns "foo/bar".
func guessRepoNameFromRemoteURL(urlStr string, hostnameToPattern map[string]string) api.RepoName {
	if !gitProtocolRegExp.MatchString(urlStr) {
		urlStr = "ssh://" + strings.Replace(strings.TrimPrefix(urlStr, "git@"), ":", "/", 1)
	}
	urlStr = strings.TrimSuffix(urlStr, ".git")
	u, _ := url.Parse(urlStr)
	if u == nil {
		return ""
	}

	pattern := "{hostname}/{path}"
	if hostnameToPattern != nil {
		if p, ok := hostnameToPattern[u.Hostname()]; ok {
			pattern = p
		}
	}

	return api.RepoName(strings.NewReplacer(
		"{hostname}", u.Hostname(),
		"{path}", strings.TrimPrefix(u.Path, "/"),
	).Replace(pattern))
}
