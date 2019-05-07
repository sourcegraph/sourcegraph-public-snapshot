// Package repoupdater implements the repo-updater service HTTP handler.
package repoupdater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Server is a repoupdater server.
type Server struct {
	// Kinds of external services synced with the new syncer
	Kinds []string
	repos.Store
	*repos.Syncer
	InternalAPI interface {
		ReposUpdateMetadata(ctx context.Context, repo api.RepoName, description string, fork, archived bool) error
	}
}

// Handler returns the http.Handler that should be used to serve requests.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/repo-update-scheduler-info", s.handleRepoUpdateSchedulerInfo)
	mux.HandleFunc("/repo-lookup", s.handleRepoLookup)
	mux.HandleFunc("/repo-external-services", s.handleRepoExternalServices)
	mux.HandleFunc("/enqueue-repo-update", s.handleEnqueueRepoUpdate)
	mux.HandleFunc("/exclude-repo", s.handleExcludeRepo)
	mux.HandleFunc("/sync-external-service", s.handleExternalServiceSync)
	return mux
}

func (s *Server) handleRepoExternalServices(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoExternalServicesRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	rs, err := s.Store.ListRepos(r.Context(), repos.StoreListReposArgs{
		IDs: []uint32{req.ID},
	})

	if err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	if len(rs) == 0 {
		respond(w, http.StatusNotFound, errors.Errorf("repository with ID %v does not exist", req.ID))
		return
	}

	var resp protocol.RepoExternalServicesResponse

	svcIDs := rs[0].ExternalServiceIDs()
	if len(svcIDs) == 0 {
		respond(w, http.StatusOK, resp)
		return
	}

	args := repos.StoreListExternalServicesArgs{
		IDs: svcIDs,
	}

	es, err := s.Store.ListExternalServices(r.Context(), args)
	if err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	resp.ExternalServices = newExternalServices(es...)

	respond(w, http.StatusOK, resp)
}

func (s *Server) handleExcludeRepo(w http.ResponseWriter, r *http.Request) {
	var resp protocol.ExcludeRepoResponse

	if len(s.Kinds) == 0 {
		respond(w, http.StatusOK, &resp)
		return
	}

	var req protocol.ExcludeRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	rs, err := s.Store.ListRepos(r.Context(), repos.StoreListReposArgs{
		IDs:   []uint32{req.ID},
		Kinds: s.Kinds,
	})

	if err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	if len(rs) == 0 {
		log15.Warn("exclude-repo: repo not found. skipping", "repo.id", req.ID)
		respond(w, http.StatusOK, resp)
		return
	}

	args := repos.StoreListExternalServicesArgs{
		Kinds: repos.Repos(rs).Kinds(),
	}

	es, err := s.Store.ListExternalServices(r.Context(), args)
	if err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	for _, e := range es {
		if err := e.Exclude(rs...); err != nil {
			break
		}
	}

	if err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	err = s.Store.UpsertExternalServices(r.Context(), es...)
	if err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	resp.ExternalServices = newExternalServices(es...)

	respond(w, http.StatusOK, resp)
}

// TODO(tsenart): Reuse this function in all handlers.
func respond(w http.ResponseWriter, code int, v interface{}) {
	switch val := v.(type) {
	case error:
		if val != nil {
			log15.Error(val.Error())
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(code)
			fmt.Fprintf(w, "%v", val)
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		bs, err := json.Marshal(v)
		if err != nil {
			respond(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(code)
		if _, err = w.Write(bs); err != nil {
			log15.Error("failed to write response", "error", err)
		}
	}
}

func newExternalServices(es ...*repos.ExternalService) []api.ExternalService {
	svcs := make([]api.ExternalService, 0, len(es))

	for _, e := range es {
		svc := api.ExternalService{
			ID:          e.ID,
			Kind:        e.Kind,
			DisplayName: e.DisplayName,
			Config:      e.Config,
			CreatedAt:   e.CreatedAt,
			UpdatedAt:   e.UpdatedAt,
		}

		if e.IsDeleted() {
			svc.DeletedAt = &e.DeletedAt
		}

		svcs = append(svcs, svc)
	}

	return svcs
}

func (s *Server) handleRepoUpdateSchedulerInfo(w http.ResponseWriter, r *http.Request) {
	var args protocol.RepoUpdateSchedulerInfoArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := repos.Scheduler.ScheduleInfo(args.ID)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleRepoLookup(w http.ResponseWriter, r *http.Request) {
	var args protocol.RepoLookupArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t := time.Now()
	result, err := s.repoLookup(r.Context(), args)
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
		respond(w, http.StatusBadRequest, err)
		return
	}

	args := repos.StoreListReposArgs{Names: []string{string(req.Repo)}}
	rs, err := s.Store.ListRepos(r.Context(), args)
	if err != nil {
		respond(w, http.StatusInternalServerError, errors.Wrap(err, "store.list-repos"))
		return
	}

	if len(rs) != 1 {
		err := errors.Errorf("repo %q not found in store", req.Repo)
		respond(w, http.StatusNotFound, err)
		return
	}

	repo := rs[0]
	if req.URL == "" {
		if urls := repo.CloneURLs(); len(urls) > 0 {
			req.URL = urls[0]
		}
	}

	repos.Scheduler.UpdateOnce(repo.ID, req.Repo, req.URL)

	respond(w, http.StatusOK, &protocol.RepoUpdateResponse{
		ID:   repo.ID,
		Name: repo.Name,
		URL:  req.URL,
	})
}

func (s *Server) handleExternalServiceSync(w http.ResponseWriter, r *http.Request) {
	var req protocol.ExternalServiceSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if s.Syncer == nil {
		log15.Debug("server.external-service-sync", "syncer", "disabled")
		return
	}

	for _, kind := range s.Kinds {
		if req.ExternalService.Kind != kind {
			continue
		}

		_, err := s.Syncer.Sync(r.Context(), kind)
		switch {
		case err == nil:
			log15.Info("server.external-service-sync", "synced", req.ExternalService.Kind)
			_ = json.NewEncoder(w).Encode(&protocol.ExternalServiceSyncResult{
				ExternalService: req.ExternalService,
				Error:           err,
			})
		default:
			log15.Error("server.external-service-sync", "kind", req.ExternalService.Kind, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

var mockRepoLookup func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)

func (s *Server) repoLookup(ctx context.Context, args protocol.RepoLookupArgs) (result *protocol.RepoLookupResult, err error) {
	tr, ctx := trace.New(ctx, "repoLookup", args.String())
	defer func() {
		log15.Debug("repoLookup", "result", result, "error", err)
		if result != nil {
			tr.LazyPrintf("result: %s", result)
		}
		tr.SetError(err)
		tr.Finish()
	}()

	if args.Repo == "" && args.ExternalRepo == nil {
		return nil, errors.New("at least one of Repo and ExternalRepo must be set (both are empty)")
	}

	if mockRepoLookup != nil {
		return mockRepoLookup(args)
	}

	type getfn struct {
		kind string
		fn   func(context.Context, protocol.RepoLookupArgs) (*protocol.RepoInfo, bool, error)
	}

	var fns []getfn

	if s.Syncer != nil {
		fns = append(fns, getfn{"SYNCER", func(ctx context.Context, args protocol.RepoLookupArgs) (*protocol.RepoInfo, bool, error) {
			repos, err := s.Store.ListRepos(ctx, repos.StoreListReposArgs{
				Names: []string{string(args.Repo)},
			})

			if err != nil || len(repos) != 1 {
				return nil, false, err
			}

			info, err := newRepoInfo(repos[0])
			if err != nil {
				return nil, false, err
			}

			return info, true, nil
		}})
	} else {
		fns = append(fns,
			getfn{"GITHUB", repos.GetGitHubRepository},
			getfn{"GITLAB", repos.GetGitLabRepository},
			getfn{"BITBUCKETSERVER", repos.GetBitbucketServerRepository},
			getfn{"AWSCODECOMMIT", repos.GetAWSCodeCommitRepository},
			getfn{"GITOLITE", repos.GetGitoliteRepository},
		)
	}

	var (
		repo          *protocol.RepoInfo
		authoritative bool
	)

	// Find the authoritative source of the repository being looked up.
	for _, get := range fns {
		if repo, authoritative, err = get.fn(ctx, args); authoritative {
			log15.Debug("repoupdater.lookup-repo", "source", get.kind)
			tr.LazyPrintf("authorative: %s", get.kind)
			break
		}
	}

	result = &protocol.RepoLookupResult{}
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
				err := s.InternalAPI.ReposUpdateMetadata(context.Background(), repo.Name, repo.Description, repo.Fork, repo.Archived)
				if err != nil {
					log15.Warn("Error updating repo metadata", "repo", repo.Name, "err", err)
				}
			}()
		}
		if err != nil {
			return nil, err
		}
		result.Repo = repo
		return result, nil
	}

	// No configured code hosts are authoritative for this repository.
	result.ErrorNotFound = true
	return result, nil
}

func newRepoInfo(r *repos.Repo) (*protocol.RepoInfo, error) {
	urls := r.CloneURLs()
	if len(urls) == 0 {
		return nil, fmt.Errorf("no clone urls for repo id=%q name=%q", r.ID, r.Name)
	}

	info := protocol.RepoInfo{
		Name:         api.RepoName(r.Name),
		Description:  r.Description,
		Fork:         r.Fork,
		Archived:     r.Archived,
		VCS:          protocol.VCSInfo{URL: urls[0]},
		ExternalRepo: &r.ExternalRepo,
	}

	switch strings.ToLower(r.ExternalRepo.ServiceType) {
	case "github":
		ghrepo := r.Metadata.(*github.Repository)
		info.Links = &protocol.RepoLinks{
			Root:   ghrepo.URL,
			Tree:   pathAppend(ghrepo.URL, "/tree/{rev}/{path}"),
			Blob:   pathAppend(ghrepo.URL, "/blob/{rev}/{path}"),
			Commit: pathAppend(ghrepo.URL, "/commit/{commit}"),
		}
	case "gitlab":
		proj := r.Metadata.(*gitlab.Project)
		info.Links = &protocol.RepoLinks{
			Root:   proj.WebURL,
			Tree:   pathAppend(proj.WebURL, "/tree/{rev}/{path}"),
			Blob:   pathAppend(proj.WebURL, "/blob/{rev}/{path}"),
			Commit: pathAppend(proj.WebURL, "/commit/{commit}"),
		}
	case "bitbucketserver":
		repo := r.Metadata.(*bitbucketserver.Repo)
		if len(repo.Links.Self) == 0 {
			break
		}

		href := repo.Links.Self[0].Href
		root := strings.TrimSuffix(href, "/browse")
		info.Links = &protocol.RepoLinks{
			Root:   href,
			Tree:   pathAppend(root, "/browse/{path}?at={rev}"),
			Blob:   pathAppend(root, "/browse/{path}?at={rev}"),
			Commit: pathAppend(root, "/commits/{commit}"),
		}
	}

	return &info, nil
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

func pathAppend(base, p string) string {
	return strings.TrimRight(base, "/") + p
}
