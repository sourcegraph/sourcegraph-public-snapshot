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
	mu    sync.RWMutex
	repos map[string]*protocol.RepoInfo
}

// NewOtherReposSyncer returns a new OtherReposSyncer.
func NewOtherReposSyncer() *OtherReposSyncer {
	return &OtherReposSyncer{
		repos: map[string]*protocol.RepoInfo{},
	}
}

// GetRepoInfoByName returns dummy repo info of the repository with the given name.
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
			if err := s.sync(ctx); err != nil {
				log15.Error("error syncing other external services repos", "error", err)
			}
		}
	}
}

func (s *OtherReposSyncer) sync(ctx context.Context) error {
	svcs, err := api.InternalClient.ExternalServicesList(ctx, api.ExternalServicesListRequest{Kind: "OTHER"})
	if err != nil {
		return err
	}

	type syncOp struct {
		svc *api.ExternalService
		err error
	}

	ch := make(chan syncOp, len(svcs))
	for i := range svcs {
		go func(op syncOp) {
			op.err = s.syncExternalService(ctx, op.svc)
			ch <- op
		}(syncOp{svc: &svcs[i]})
	}

	for i := 0; i < cap(ch); i++ {
		if op := <-ch; op.err != nil {
			log15.Error(
				"failed to sync external service",
				"id", op.svc.ID,
				"displayName", op.svc.DisplayName,
				"kind", op.svc.Kind,
				"config", op.svc.Config,
				"err", op.err,
			)
		} else {
			id := strconv.FormatInt(op.svc.ID, 10)
			otherExternalServicesUpdateTime.WithLabelValues(id).Set(float64(time.Now().Unix()))
		}
	}

	return nil
}

func (s *OtherReposSyncer) syncExternalService(ctx context.Context, svc *api.ExternalService) error {
	osvc, err := newOtherExternalService(svc)
	if err != nil {
		return err
	}

	cloneURLs, err := osvc.CloneURLs()
	if err != nil {
		return err
	}

	ch := make(chan repoCreateOrUpdateRequest)
	defer close(ch)

	go createEnableUpdateRepos(ctx, fmt.Sprintf("other:%d", osvc.ID), ch)

	for _, u := range cloneURLs {
		repoURL := u.String()
		repoName := otherRepoName(u)
		u.Path, u.RawQuery = "", ""
		serviceID := u.String()

		ch <- repoCreateOrUpdateRequest{
			RepoCreateOrUpdateRequest: api.RepoCreateOrUpdateRequest{
				RepoName: repoName,
				ExternalRepo: &api.ExternalRepoSpec{
					ID:          string(repoName),
					ServiceType: "other",
					ServiceID:   serviceID,
				},
				Enabled: true,
			},
			URL: repoURL,
		}

		s.cacheRepo(repoName)
	}

	return nil
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

func (s *OtherReposSyncer) cacheRepo(name api.RepoName) {
	s.mu.Lock()
	s.repos[string(name)] = &protocol.RepoInfo{Name: name}
	s.mu.Unlock()
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
