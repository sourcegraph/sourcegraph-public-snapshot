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
	repos.Store
	*repos.Syncer
	GithubDotComSource interface {
		GetRepo(ctx context.Context, nameWithOwner string) (*repos.Repo, error)
	}
	Scheduler interface {
		UpdateQueueLen() int
		UpdateOnce(id uint32, name api.RepoName, url string)
		ScheduleInfo(id uint32) *protocol.RepoUpdateSchedulerInfoResult
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
	mux.HandleFunc("/status-messages", s.handleStatusMessages)
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
	var req protocol.ExcludeRepoRequest
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

	var resp protocol.ExcludeRepoResponse
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
			respond(w, http.StatusInternalServerError, err)
			return
		}
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

	result := s.Scheduler.ScheduleInfo(args.ID)
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

	s.Scheduler.UpdateOnce(repo.ID, req.Repo, req.URL)

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

	_, err := s.Syncer.Sync(r.Context(), req.ExternalService.Kind)
	switch {
	case err == nil:
		log15.Info("server.external-service-sync", "synced", req.ExternalService.Kind)
		respond(w, http.StatusOK, &protocol.ExternalServiceSyncResult{
			ExternalService: req.ExternalService,
		})
	case err == github.ErrIncompleteResults:
		log15.Info("server.external-service-sync", "kind", req.ExternalService.Kind, "error", err)
		syncResult := &protocol.ExternalServiceSyncResult{
			ExternalService: req.ExternalService,
			Error:           err.Error(),
		}
		respond(w, http.StatusOK, syncResult)
	default:
		log15.Error("server.external-service-sync", "kind", req.ExternalService.Kind, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	if args.Repo == "" {
		return nil, errors.New("Repo must be set (is blank)")
	}

	if mockRepoLookup != nil {
		return mockRepoLookup(args)
	}

	var repo *repos.Repo
	result = &protocol.RepoLookupResult{}

	if s.shouldGetGithubDotComRepo(args) {
		nameWithOwner := strings.TrimPrefix(string(args.Repo), "github.com/")
		repo, err = s.GithubDotComSource.GetRepo(ctx, nameWithOwner)
		if err != nil {
			if github.IsNotFound(err) {
				result.ErrorNotFound = true
				return result, nil
			}
			if isUnauthorized(err) {
				result.ErrorUnauthorized = true
				return result, nil
			}
			if isTemporarilyUnavailable(err) {
				result.ErrorTemporarilyUnavailable = true
				return result, nil
			}
			return nil, err
		}

		_, err = s.Syncer.SyncSubset(ctx, repo)
		if err != nil {
			return nil, err
		}
	} else {
		repos, err := s.Store.ListRepos(ctx, repos.StoreListReposArgs{
			Names: []string{string(args.Repo)},
		})
		if err != nil {
			return nil, err
		}

		if len(repos) != 1 {
			result.ErrorNotFound = true
			return result, nil
		}
		repo = repos[0]
	}

	repoInfo, err := newRepoInfo(repo)
	if err != nil {
		return nil, err
	}

	result.Repo = repoInfo
	return result, nil
}

func (s *Server) shouldGetGithubDotComRepo(args protocol.RepoLookupArgs) bool {
	if s.GithubDotComSource == nil {
		return false
	}

	repoName := strings.ToLower(string(args.Repo))
	return strings.HasPrefix(repoName, "github.com/")
}

func (s *Server) handleStatusMessages(w http.ResponseWriter, r *http.Request) {
	resp := protocol.StatusMessagesResponse{
		Messages: []protocol.StatusMessage{},
	}

	enqueued := s.Scheduler.UpdateQueueLen()
	if enqueued != 0 {
		resp.Messages = append(resp.Messages, protocol.StatusMessage{
			Message: fmt.Sprintf("Currently updating %d repositories...", enqueued),
			Type:    protocol.CloningStatusMessage,
		})
	}

	log15.Debug("TRACE handleStatusMessages", "messages", resp.Messages)

	respond(w, http.StatusOK, resp)
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
	case "awscodecommit":
		repo := r.Metadata.(*awscodecommit.Repository)
		if repo.ARN == "" {
			break
		}

		splittedARN := strings.Split(strings.TrimPrefix(repo.ARN, "arn:aws:codecommit:"), ":")
		if len(splittedARN) == 0 {
			break
		}
		region := splittedARN[0]
		webURL := fmt.Sprintf("https://%s.console.aws.amazon.com/codecommit/home#/repository/%s", region, repo.Name)
		info.Links = &protocol.RepoLinks{
			Root:   webURL,
			Tree:   webURL + "/browse/{rev}/--/{path}",
			Blob:   webURL + "/browse/{rev}/--/{path}",
			Commit: webURL + "/commit/{commit}",
		}
	}

	return &info, nil
}

func pathAppend(base, p string) string {
	return strings.TrimRight(base, "/") + p
}

func isUnauthorized(err error) bool {
	code := github.HTTPErrorCode(err)
	return code == http.StatusUnauthorized || code == http.StatusForbidden
}

func isTemporarilyUnavailable(err error) bool {
	return github.IsRateLimitExceeded(err)
}
