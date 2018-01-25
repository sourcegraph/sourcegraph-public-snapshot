package repoupdater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var githubdotcomClient = github.NewClient(
	&url.URL{Scheme: "https", Host: "api.github.com"},
	"", // TODO!(sqs): use token
	nil,
)

func repoLookup(ctx context.Context, args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
	var result protocol.RepoLookupResult

	switch {
	case strings.HasPrefix(strings.ToLower(string(args.Repo)), "github.com/"):
		nameWithOwner := string(args.Repo[len("github.com/"):])
		owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
		if err != nil {
			return nil, err
		}
		repo, err := githubdotcomClient.GetRepository(ctx, owner, name)
		if err == nil {
			result.Repo = &protocol.RepoInfo{
				URI:         api.RepoURI("github.com/" + repo.NameWithOwner),
				Description: repo.Description,
				Fork:        repo.IsFork,
			}
		} else if err != nil && !github.IsNotFound(err) {
			return nil, err
		}

	default:
		if err := gitserver.DefaultClient.IsRepoCloneable(ctx, args.Repo); err == nil {
			result.Repo = &protocol.RepoInfo{
				URI: args.Repo,
			}
		}
	}

	return &result, nil
}
