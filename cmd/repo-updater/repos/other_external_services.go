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
	ticks := time.NewTicker(interval)
	defer ticks.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticks.C:
			log15.Debug("syncing all OTHER external services")

			results, err := s.syncAll(ctx)
			if err != nil {
				log15.Error("error syncing other external services repos", "error", err)
			}

			for _, err := range results.Errors() {
				log15.Error("sync error", err.Error())
			}
		}
	}
}

// syncAll syncrhonizes all "OTHER" external services.
func (s *OtherReposSyncer) syncAll(ctx context.Context) (OtherSyncResults, error) {
	svcs, err := s.api.ExternalServicesList(ctx, api.ExternalServicesListRequest{Kind: "OTHER"})
	if err != nil {
		return nil, err
	}
	return s.SyncMany(ctx, svcs...), nil
}

// OtherSyncResults is a helper type for lists of OtherSyncResults.
type OtherSyncResults []*OtherSyncResult

// Errors returns all Errors in the list of OtherSyncResults.
func (rs OtherSyncResults) Errors() (errs []*OtherSyncError) {
	for _, res := range rs {
		errs = append(errs, res.Errors...)
	}
	return errs
}

// OtherSyncErrors is a helper type for lists of OtherSyncErrors.
type OtherSyncErrors []*OtherSyncError

// Error implements the error interface.
func (errs OtherSyncErrors) Error() string {
	var sb strings.Builder
	for _, err := range errs {
		sb.WriteString(err.Error() + "; ")
	}
	return sb.String()
}

// RepoErrors returns all OtherSyncErrors that have a Repo set.
func (errs OtherSyncErrors) RepoErrors() OtherSyncErrors {
	se := make(OtherSyncErrors, 0, len(errs))
	for _, err := range errs {
		if err.Repo != nil {
			se = append(se, err)
		}
	}
	return se
}

// OtherSyncError is an error type containing information about a failed sync of an external service
// of kind "OTHER".
type OtherSyncError struct {
	// External service that had an error synchronizing.
	Service *api.ExternalService
	// Repo that failed synchronizing. This may be nil if the synchronization
	// process failed before attempting to sync each repo of defined by the external
	// service config.
	Repo *protocol.RepoInfo
	// The actual error.
	Err string
}

// Error implements the error interface.
func (e OtherSyncError) Error() string {
	if e.Repo == nil {
		return fmt.Sprintf("external-service=%d: %s", e.Service.ID, e.Err)
	}
	return fmt.Sprintf("external-service=%d repo=%q: %s", e.Service.ID, e.Repo.Name, e.Err)
}

// OtherSyncResult is returned by Sync to indicate which external services and their
// repos synced successfully and which didn't.
type OtherSyncResult struct {
	// The external service of kind "OTHER" that had its repos synced.
	Service *api.ExternalService
	// Repos that succeeded to be synced.
	Synced []*protocol.RepoInfo
	// Repos that failed to be synced.
	Errors OtherSyncErrors
}

// SyncMany synchonizes the repos defined by all the given external services of kind "OTHER".
// It return a OtherSyncResults containing which repos were synced and which failed to.
func (s *OtherReposSyncer) SyncMany(ctx context.Context, svcs ...*api.ExternalService) OtherSyncResults {
	if len(svcs) == 0 {
		return nil
	}

	ch := make(chan *OtherSyncResult, len(svcs))
	for _, svc := range svcs {
		go func(svc *api.ExternalService) {
			ch <- s.Sync(ctx, svc)
		}(svc)
	}

	results := make([]*OtherSyncResult, 0, len(svcs))
	for i := 0; i < cap(ch); i++ {
		res := <-ch
		results = append(results, res)
	}

	return results
}

// Sync synchronizes the repositories of a single external service of kind "OTHER"
func (s *OtherReposSyncer) Sync(ctx context.Context, svc *api.ExternalService) (res *OtherSyncResult) {
	defer func(began time.Time) {
		id, now := strconv.FormatInt(svc.ID, 10), time.Now().UTC()
		otherExternalServicesLastSync.WithLabelValues(id).Set(float64(now.Unix()))
		otherExternalServicesSyncedReposTotal.WithLabelValues(id, "synced").Add(float64(len(res.Synced)))
		otherExternalServicesSyncedReposTotal.WithLabelValues(id, "errored").Add(float64(len(res.Errors.RepoErrors())))
		otherExternalServicesSyncDuration.WithLabelValues(id).Observe(time.Since(began).Seconds())
	}(time.Now().UTC())

	cloneURLs, err := otherExternalServiceCloneURLs(svc)
	if err != nil {
		return &OtherSyncResult{
			Service: svc,
			Errors:  OtherSyncErrors{{Service: svc, Err: err.Error()}},
		}
	}

	repos := make([]*protocol.RepoInfo, 0, len(cloneURLs))
	for _, u := range cloneURLs {
		repos = append(repos, repoFromCloneURL(u))
	}

	return s.store(ctx, svc, repos...)
}

func repoFromCloneURL(u *url.URL) *protocol.RepoInfo {
	repoURL := u.String()
	repoName := otherRepoName(u)
	u.Path, u.RawQuery = "", ""
	serviceID := u.String()

	return &protocol.RepoInfo{
		Name: repoName,
		VCS:  protocol.VCSInfo{URL: repoURL},
		ExternalRepo: &api.ExternalRepoSpec{
			ID:          string(repoName),
			ServiceType: "other",
			ServiceID:   serviceID,
		},
	}
}

// store upserts the given repos through the FrontendAPI, returning which succeeded
// and which failed to be processed.
func (s *OtherReposSyncer) store(ctx context.Context, svc *api.ExternalService, repos ...*protocol.RepoInfo) *OtherSyncResult {
	if len(repos) == 0 {
		return &OtherSyncResult{Service: svc}
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

	res := OtherSyncResult{Service: svc}
	for i := 0; i < cap(ch); i++ {
		if op := <-ch; op.err != nil {
			res.Errors = append(res.Errors, &OtherSyncError{
				Service: svc,
				Repo:    op.repo,
				Err:     op.err.Error(),
			})
		} else {
			res.Synced = append(res.Synced, op.repo)
			s.cache(op.repo)
		}
	}
	return &res
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
	u.User = nil
	u.Scheme = ""
	u.RawQuery = ""
	u.Fragment = ""
	return api.RepoName(otherRepoNameReplacer.Replace(u.String()))
}

// otherExternalServiceCloneURLs returns all cloneURLs of the given "OTHER" external service.
func otherExternalServiceCloneURLs(s *api.ExternalService) ([]*url.URL, error) {
	var c schema.OtherExternalServiceConnection
	if err := jsonc.Unmarshal(s.Config, &c); err != nil {
		return nil, fmt.Errorf("config error: %s", err)
	}

	if len(c.Repos) == 0 {
		return nil, nil
	}

	parseRepo := url.Parse
	if c.Url != "" {
		baseURL, err := url.Parse(c.Url)
		if err != nil {
			return nil, err
		}
		parseRepo = baseURL.Parse
	}

	cloneURLs := make([]*url.URL, 0, len(c.Repos))
	for _, repo := range c.Repos {
		cloneURL, err := parseRepo(repo)
		if err != nil {
			log15.Error("skipping invalid repo clone URL", "repo", repo, "url", c.Url, "error", err)
			continue
		}
		cloneURLs = append(cloneURLs, cloneURL)
	}

	return cloneURLs, nil
}
