package repoupdater

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/awscodecommit"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/github"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/gitlab"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Server is a repoupdater server.
type Server struct {
}

// Handler returns the http.Handler that should be used to serve requests.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/repo-lookup", s.handleRepoLookup)
	mux.HandleFunc("/enqueue-repo-update", s.handleEnqueueRepoUpdate)
	return mux
}

func (s *Server) handleRepoLookup(w http.ResponseWriter, r *http.Request) {
	var args protocol.RepoLookupArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t := time.Now()
	result, err := repoLookup(r.Context(), args)
	if err != nil {
		if err == context.Canceled {
			http.Error(w, "request canceled", http.StatusGatewayTimeout)
			return
		}
		log15.Error("repoLookup failed", "args", &args, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log15.Debug("TRACE repoLookup", "args", &args, "result", result, "duration", time.Since(t))

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleEnqueueRepoUpdate(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repos.UpdateOnce(r.Context(), req.Repo, req.URL)
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
	if !authoritative {
		repo, authoritative, err = repos.GetBitbucketServerRepository(ctx, args)
	}
	if !authoritative {
		repo, authoritative, err = repos.GetAWSCodeCommitRepository(ctx, args)
	}
	if !authoritative {
		repo, authoritative, err = repos.GetExplicitlyConfiguredRepository(ctx, args)
	}
	if !authoritative {
		repo, authoritative, err = repos.GetGitoliteRepository(ctx, args)
	}
	if authoritative {
		if isNotFound(err) {
			result.ErrorNotFound = true
			err = nil
		} else if isUnauthorized(err) {
			result.ErrorUnauthorized = true
			err = nil
		} else if isTemporarilyUnavailable(err) {
			result.ErrorTemporarilyUnavailable = true
			err = nil
		}
		if err != nil {
			return nil, err
		}
		if repo != nil {
			go func() {
				err := api.InternalClient.ReposUpdateMetadata(ctx, repo.URI, repo.Description, repo.Fork, repo.Archived)
				if err != nil {
					log15.Warn("Error updating repo metadata", "repo", repo.URI, "err", err)
				}
			}()
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
	return github.IsNotFound(err) || gitlab.IsNotFound(err) || awscodecommit.IsNotFound(err) || errcode.IsNotFound(err)
}

func isUnauthorized(err error) bool {
	// TODO(sqs): reduce duplication
	if awscodecommit.IsUnauthorized(err) || errcode.IsUnauthorized(err) {
		return true
	}
	code := github.HTTPErrorCode(err)
	if code == 0 {
		code = gitlab.HTTPErrorCode(err)
	}
	return code == http.StatusUnauthorized || code == http.StatusForbidden
}

func isTemporarilyUnavailable(err error) bool {
	return err == repos.ErrGitHubAPITemporarilyUnavailable || github.IsRateLimitExceeded(err)
}
