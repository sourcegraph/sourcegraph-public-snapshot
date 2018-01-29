package repoupdater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/github"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
)

// Server is a repoupdater server.
type Server struct{}

// Handler returns the http.Handler that should be used to serve requests.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/repo-lookup", s.handleRepoLookup)
	return mux
}

func (s *Server) handleRepoLookup(w http.ResponseWriter, r *http.Request) {
	var args protocol.RepoLookupArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := repoLookup(r.Context(), args)
	if err != nil {
		code := github.HTTPErrorCode(err)
		if code == 0 {
			code = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), code)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var (
	githubdotcomClient  *github.Client
	githubdotcomBaseURL = *repos.NormalizeGitHubBaseURL(&url.URL{Scheme: "https", Host: "github.com", Path: "/"})
	githubdotcomAPIURL  = url.URL{Scheme: "https", Host: "api.github.com", Path: "/"}
)

func init() {
	var githubdotcomToken string
	if c := conf.FirstGitHubDotComConnectionWithToken(); c != nil {
		githubdotcomToken = c.Token
	}
	githubdotcomClient = github.NewClient(&githubdotcomAPIURL, githubdotcomToken, nil)
}

var mockRepoLookup func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)

func repoLookup(ctx context.Context, args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
	if mockRepoLookup != nil {
		return mockRepoLookup(args)
	}

	var result protocol.RepoLookupResult

	switch {
	// TODO(sqs): support GitHub (Enterprise) hosts other than github.com
	case strings.HasPrefix(strings.ToLower(string(args.Repo)), "github.com/"):
		nameWithOwner := string(args.Repo[len("github.com/"):])
		owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
		if err != nil {
			return nil, err
		}
		repo, err := githubdotcomClient.GetRepository(ctx, owner, name)
		if err == nil {
			result.Repo = &protocol.RepoInfo{
				URI:          api.RepoURI("github.com/" + repo.NameWithOwner),
				Description:  repo.Description,
				Fork:         repo.IsFork,
				ExternalRepo: repos.GitHubExternalRepoSpec(repo, githubdotcomBaseURL),
			}
		} else if err != nil && !github.IsNotFound(err) {
			return nil, err
		}
	}

	return &result, nil
}
