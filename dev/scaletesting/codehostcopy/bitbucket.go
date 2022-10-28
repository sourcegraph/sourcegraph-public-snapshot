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

const separator = "_-_"

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

	// The basic auth client has more power in the rest API than the token based client
	c := bitbucket.NewBasicAuthClient(def.Username, def.Password, u, bitbucket.WithTimeout(15*time.Second))

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
		if l.Name == "https" || l.Name == "http" {
			return url.Parse(l.Url)
		}
	}
	return nil, errors.New("no https url found")
}

// ListRepos retrieves all repos from the bitbucket server. After all repos are retrieved the http or https clone
// url is extracted. Note that the repo name has the following format: <project key>_-_<repo name>. Thus if you
// just want the repo name you would have to strip the project key and '_-_' separator out.
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
		cloneUrl.User = url.UserPassword(bt.def.Username, bt.def.Password)
		results = append(results, &store.Repo{
			Name:   fmt.Sprintf("%s%s%s", r.Project.Key, separator, r.Name),
			GitURL: cloneUrl.String(),
		})
	}

	return results, nil
}

// CreateRepo creates a repo on bitbucket. It is assumed that the repo name has the following format: <project key>_-_<repo name>.
// A repo can only be created under a project in bitbucket, therefore the project is extract from the repo name format and a
// project is created first, if and only if, the project does not exist already. If the project already exists, the repo
// will be created and the created repos git clone url will be returned.
func (bt *BitbucketCodeHost) CreateRepo(ctx context.Context, name string) (*url.URL, error) {
	parts := strings.Split(name, separator)
	if len(parts) != 2 {
		return nil, errors.New("invalid name format - expected <project key>_-_<repo name>")
	}
	key := parts[0]
	repoName := parts[1]

	var apiErr *bitbucket.APIError
	_, err := bt.c.GetProjectByKey(ctx, key)
	if err != nil {
		// if the error is an api error, log it and continue
		// otherwise something severe is wrong and we must quit
		// early
		if errors.As(err, &apiErr) {
			// if the project was 'not found' create it
			if apiErr.StatusCode == 404 {
				bt.logger.Debug("creating project", log.String("key", key))
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
		// If the repo already exists, get it and assign it to repo
		if errors.As(err, &apiErr) && apiErr.StatusCode == 409 {
			bt.logger.Warn("repo already exists", log.String("project", key), log.String("repo", repoName))
			repo, err = bt.c.GetRepo(ctx, key, repoName)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	bt.logger.Info("created repo", log.String("project", repo.Project.Key), log.String("repo", repo.Name))
	gitURL, err := getCloneUrl(repo)
	if err != nil {
		return nil, err
	}
	// for bitbucket, you can't use the account password to git push - you actually need to use the Token ...
	gitURL.User = url.UserPassword(bt.def.Username, bt.def.Token)
	return gitURL, err
}
