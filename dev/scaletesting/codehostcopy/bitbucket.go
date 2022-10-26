package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/scaletesting/codehostcopy/bitbucket"
	"github.com/sourcegraph/sourcegraph/dev/scaletesting/internal/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type BitbucketCodeHost struct {
	logger log.Logger
	def    *CodeHostDefinition
	c      *bitbucket.Client
}

func NewBitbucketCodeHost(ctx context.Context, logger log.Logger, def *CodeHostDefinition) (*BitbucketCodeHost, error) {
	u, err := url.Parse(def.URL)
	if err != nil {
		return nil, err
	}

	var c *bitbucket.Client
	if def.Token != "" {
		c = bitbucket.NewTokenClient(def.Token, u, bitbucket.WithTimeout(15*time.Second))

	} else {
		c = bitbucket.NewBasicAuthClient(def.Username, def.Password, u, bitbucket.WithTimeout(15*time.Second))
	}

	return &BitbucketCodeHost{
		logger: logger.Scoped("bitbucket", "client that interacts with bitbucket server rest api"),
		def:    def,
		c:      c,
	}, nil
}

func getCloneUrl(repo *bitbucket.Repo) (*url.URL, error) {
	cloneLinks, ok := repo.Links["clone"]
	if !ok {
		return nil, errors.Newf("no clone links on repo %s", repo.Name)
	}
	for _, l := range cloneLinks {
		if l.Name == "https" {
			return url.Parse(l.Url)
		}
	}
	return nil, errors.New("no https url found")
}

func (bt *BitbucketCodeHost) ListRepos(ctx context.Context) ([]*store.Repo, error) {
	bt.logger.Info("fetching repos")
	repos, err := bt.c.ListRepos(ctx)
	if err != nil {
		bt.logger.Error("failed to list repos", log.Error(err))
	}

	bt.logger.Info("fetched list of repos", log.Int("repos", len(repos)))

	results := make([]*store.Repo, 0, len(repos))
	for _, r := range repos {
		cloneUrl, err := getCloneUrl(r)
		if err != nil {
			bt.logger.Warn("failed to get clone url", log.String("repo", r.Name), log.String("project", r.Project.Key), log.Error(err))
			continue
		}

		// to be able to push this repo we need to project key, incase we need to create the project before pushing
		results = append(results, &store.Repo{
			Name:   fmt.Sprintf("%s::%s", r.Project.Key, r.Name),
			GitURL: cloneUrl.String(),
		})
	}

	return results, nil
}

func (bt *BitbucketCodeHost) CreateRepo(ctx context.Context, name string) (*url.URL, error) {
	parts := strings.Split(name, "::")
	if len(parts) != 2 {
		return nil, errors.New("invalid name format - expected <project key>::<repo name>")
	}
	key := parts[0]
	repoName := parts[1]

	_, err := bt.c.GetProjectByKey(ctx, key)
	if err != nil {
		var apiErr *bitbucket.APIError
		// if the error is an api error, log it and continue
		// otherwise something severe is wrong and we must quit
		// early
		if errors.As(err, &apiErr) {
			// if the project was 'not found' create it
			if apiErr.StatusCode == 404 {
				p, err := bt.c.CreateProject(ctx, &bitbucket.Project{Key: key})
				if err != nil {
					return nil, err
				}
				bt.logger.Info("created project", log.String("project", p.Key))
			}
		} else {
			return nil, err
		}
	}
	// project already exists so lets just return the url to use
	repo, err := bt.c.CreateRepo(ctx, &bitbucket.Project{Key: key}, repoName)
	if err != nil {
		return nil, err
	}
	bt.logger.Info("created repo", log.String("project", repo.Project.Key), log.String("repo", repo.Name))
	return getCloneUrl(repo)

}
