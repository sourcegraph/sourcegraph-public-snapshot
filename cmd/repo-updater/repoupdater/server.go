package repoupdater

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/gitlab"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/github"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
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
		log15.Error("repoLookup failed", "args", args, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var mockRepoLookup func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)

func repoLookup(ctx context.Context, args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
	if args.Repo == "" && args.ExternalRepo == nil {
		return nil, errors.New("at least one of Repo and ExternalRepo must be set (both are empty)")
	}

	if mockRepoLookup != nil {
		return mockRepoLookup(args)
	}

	var result protocol.RepoLookupResult

	// Try all GetXyzRepository funcs until one returns authoritatively.
	repo, authoritative, err := repos.GetGitHubRepository(ctx, args)
	if !authoritative {
		repo, authoritative, err = repos.GetGitLabRepository(ctx, args)
	}
	if authoritative {
		if isNotFound(err) {
			result.ErrorNotFound = true
			err = nil
		} else if isUnauthorized(err) {
			result.ErrorUnauthorized = true
			err = nil
		}
		if err != nil {
			return nil, err
		}
		result.Repo = repo
		return &result, nil
	}

	// No configured code hosts are authoritative for this repository.
	result.ErrorNotFound = true
	return &result, nil
}

func isNotFound(err error) bool {
	// TODO(sqs): reduce duplication
	return github.IsNotFound(err) || gitlab.IsNotFound(err)
}

func isUnauthorized(err error) bool {
	// TODO(sqs): reduce duplication
	code := github.HTTPErrorCode(err)
	if code == 0 {
		code = gitlab.HTTPErrorCode(err)
	}
	return code == http.StatusUnauthorized || code == http.StatusForbidden
}
