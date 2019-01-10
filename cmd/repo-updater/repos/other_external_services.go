package repos

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// OtherReposSyncer periodically synchronizes the configured repos in "OTHER" external service
// connections with the stored repos in Sourcegraph.
type OtherReposSyncer struct {
	// InternalAPI client used to fetch all external servicess and upsert repos.
	api InternalAPI
	// RWMutex synchronizing access to repos below
	mu sync.RWMutex
	// Latest synced repos cache used by GetRepoInfoByName.
	repos map[string]*protocol.RepoInfo
	// Channel passed in NewOtherReposSyncer where synced repos are sent to after being cached.
	synced chan<- *protocol.RepoInfo
}

// InternalAPI captures the internal API methods needed for syncing external services' repos.
type InternalAPI interface {
	ExternalServicesList(context.Context, api.ExternalServicesListRequest) ([]*api.ExternalService, error)
	ReposCreateIfNotExists(context.Context, api.RepoCreateOrUpdateRequest) (*api.Repo, error)
	ReposUpdateMetadata(ctx context.Context, repo api.RepoName, description string, fork, archived bool) error
}

// NewOtherReposSyncer returns a new OtherReposSyncer. Synced repos will be sent on the given channel.
func NewOtherReposSyncer(api InternalAPI, synced chan<- *protocol.RepoInfo) *OtherReposSyncer {
	return &OtherReposSyncer{
		api:    api,
		repos:  map[string]*protocol.RepoInfo{},
		synced: synced,
	}
}

// GetRepoInfoByName returns repo info of the repository with the given name.
func (s *OtherReposSyncer) GetRepoInfoByName(ctx context.Context, name string) *protocol.RepoInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repos[name]
}

// Run periodically synchronizes the configured repos in "OTHER" external service
// connections with the stored repos in Sourcegraph. Termination is done through the passed context.
func (s *OtherReposSyncer) Run(ctx context.Context, interval time.Duration) error {
	ticks := time.NewTimer(interval)
	defer ticks.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticks.C:
			log15.Info("syncing all OTHER external services")
			if err := s.syncAll(ctx); err != nil {
				log15.Error("error syncing other external services repos", "error", err)
			}
		}
	}
}

// syncAll syncrhonizes all "OTHER" external services.
func (s *OtherReposSyncer) syncAll(ctx context.Context) error {
	svcs, err := s.api.ExternalServicesList(ctx, api.ExternalServicesListRequest{Kind: "OTHER"})
	if err != nil {
		return err
	}
	return s.Sync(ctx, svcs...)
}

// SyncError is an aggregate error type returned by the Sync method, containing detailed
// information about which external services failed to sync and why.
type SyncError struct {
	Errors map[*api.ExternalService]error
}

// Error implements the error interface.
func (e SyncError) Error() string {
	var sb strings.Builder
	for svc, err := range e.Errors {
		if err != nil {
			_, _ = fmt.Fprintf(&sb, "external_service[id]=%d: %s; ", svc.ID, err)
		}
	}
	return sb.String()
}

// Sync syncs the given external services.
func (s *OtherReposSyncer) Sync(ctx context.Context, svcs ...*api.ExternalService) error {
	if len(svcs) == 0 {
		return nil
	}

	type syncOp struct {
		svc *api.ExternalService
		err error
	}

	ch := make(chan syncOp, len(svcs))
	for _, svc := range svcs {
		go func(op syncOp) {
			op.err = s.sync(ctx, op.svc)
			ch <- op
		}(syncOp{svc: svc})
	}

	ids := map[string][]string{}
	errs := SyncError{Errors: make(map[*api.ExternalService]error, cap(ch))}
	for i := 0; i < cap(ch); i++ {
		op := <-ch
		if id := strconv.FormatInt(op.svc.ID, 10); op.err != nil {
			errs.Errors[op.svc] = op.err
			ids["err"] = append(ids["err"], id)
		} else {
			ids["ok"] = append(ids["ok"], id)
			otherExternalServicesUpdateTime.WithLabelValues(id).Set(float64(time.Now().Unix()))
		}
	}

	log15.Info("synced OTHER external services", "ok", ids["ok"], "err", ids["err"])

	if len(errs.Errors) > 0 {
		return &errs
	}

	return nil
}

func (s *OtherReposSyncer) sync(ctx context.Context, svc *api.ExternalService) error {
	osvc, err := newOtherExternalService(svc)
	if err != nil {
		return err
	}

	cloneURLs, err := osvc.CloneURLs()
	if err != nil {
		return err
	}

	repos := make([]*protocol.RepoInfo, 0, len(cloneURLs))

	for _, u := range cloneURLs {
		repoURL := u.String()
		repoName := otherRepoName(u)
		u.Path, u.RawQuery = "", ""
		serviceID := u.String()

		repos = append(repos, &protocol.RepoInfo{
			Name: repoName,
			VCS:  protocol.VCSInfo{URL: repoURL},
			ExternalRepo: &api.ExternalRepoSpec{
				ID:          string(repoName),
				ServiceType: "other",
				ServiceID:   serviceID,
			},
		})
	}

	return s.store(ctx, repos...)
}

// StoreError is an aggregate error type returned by the upsert method, containing detailed
// information about which repos of an external service failed to be stored and why.
type StoreError struct {
	Errors map[*protocol.RepoInfo]error
}

// Error implements the error interface.
func (e StoreError) Error() string {
	var sb strings.Builder
	for r, err := range e.Errors {
		if err != nil {
			_, _ = fmt.Fprintf(&sb, "repo[name]=%s: %s; ", r.Name, err)
		}
	}
	return sb.String()
}

func (s *OtherReposSyncer) store(ctx context.Context, repos ...*protocol.RepoInfo) error {
	if len(repos) == 0 {
		return nil
	}

	type storeOp struct {
		repo *protocol.RepoInfo
		err  error
	}

	ch := make(chan storeOp, len(repos))
	for _, repo := range repos {
		go func(op storeOp) {
			op.err = s.upsert(ctx, op.repo)
			ch <- op
		}(storeOp{repo: repo})
	}

	errs := StoreError{Errors: make(map[*protocol.RepoInfo]error, cap(ch))}
	for i := 0; i < cap(ch); i++ {
		if op := <-ch; op.err != nil {
			errs.Errors[op.repo] = op.err
		} else {
			s.cache(op.repo)
		}
	}

	if len(errs.Errors) > 0 {
		return &errs
	}

	return nil
}

func (s *OtherReposSyncer) cache(repo *protocol.RepoInfo) {
	s.mu.Lock()
	s.repos[string(repo.Name)] = repo
	s.mu.Unlock()

	if s.synced != nil {
		s.synced <- repo
	}
}

func (s *OtherReposSyncer) upsert(ctx context.Context, repo *protocol.RepoInfo) error {
	_, err := s.api.ReposCreateIfNotExists(ctx, api.RepoCreateOrUpdateRequest{
		RepoName:     repo.Name,
		Enabled:      true,
		Fork:         repo.Fork,
		Archived:     repo.Archived,
		Description:  repo.Description,
		ExternalRepo: repo.ExternalRepo,
	})

	if err != nil {
		return err
	}

	return s.api.ReposUpdateMetadata(
		ctx,
		repo.Name,
		repo.Description,
		repo.Fork,
		repo.Archived,
	)
}

var otherRepoNameReplacer = strings.NewReplacer(":", "-", "@", "-", "//", "")

func otherRepoName(cloneURL *url.URL) api.RepoName {
	u := *cloneURL
	if u.User != nil {
		user := u.User.Username()
		pass, ok := u.User.Password()
		if !ok {
			u.User = url.User(user)
		} else {
			u.User = url.UserPassword(user, pass)
		}
	}

	u.Scheme = ""
	u.RawQuery = ""
	u.Fragment = ""

	return api.RepoName(otherRepoNameReplacer.Replace(u.String()))
}

type otherExternalService struct {
	*api.ExternalService
	*schema.OtherExternalServiceConnection
}

func newOtherExternalService(s *api.ExternalService) (*otherExternalService, error) {
	var conn schema.OtherExternalServiceConnection
	if err := jsonc.Unmarshal(s.Config, &conn); err != nil {
		return nil, err
	}
	return &otherExternalService{
		ExternalService:                s,
		OtherExternalServiceConnection: &conn,
	}, nil
}

func (s otherExternalService) CloneURLs() ([]*url.URL, error) {
	if len(s.Repos) == 0 {
		return nil, nil
	}

	parseRepo := url.Parse
	if s.Url != "" {
		baseURL, err := url.Parse(s.Url)
		if err != nil {
			return nil, err
		}
		parseRepo = baseURL.Parse
	}

	cloneURLs := make([]*url.URL, 0, len(s.Repos))
	for _, repo := range s.Repos {
		cloneURL, err := parseRepo(repo)
		if err != nil {
			log15.Error("skipping invalid repo clone URL", "repo", repo, "url", s.Url, "error", err)
			continue
		}
		cloneURLs = append(cloneURLs, cloneURL)
	}

	return cloneURLs, nil
}
