package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
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

	project *bitbucket.Project
	once    sync.Once

	page    int
	perPage int
	done    bool
	err     error
}

func NewBitbucketCodeHost(logger log.Logger, def *CodeHostDefinition) (*BitbucketCodeHost, error) {
	u, err := url.Parse(def.URL)
	if err != nil {
		return nil, err
	}

	// The basic auth client has more power in the rest API than the token based client
	c := bitbucket.NewBasicAuthClient(def.Username, def.Password, u, bitbucket.WithTimeout(15*time.Second))

	return &BitbucketCodeHost{
		logger:  logger.Scoped("bitbucket"),
		def:     def,
		c:       c,
		perPage: 30,
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

func (bt *BitbucketCodeHost) GitOpts() []GitOpt {
	return nil
}

func (bt *BitbucketCodeHost) AddSSHKey(ctx context.Context) (int64, error) {
	return 0, nil
}

func (bt *BitbucketCodeHost) DropSSHKey(ctx context.Context, keyID int64) error {
	return nil
}

func (bt *BitbucketCodeHost) InitializeFromState(ctx context.Context, stateRepos []*store.Repo) (int, int, error) {
	return bt.def.RepositoryLimit, -1, nil
}

// listRepos retrieves all repos from the bitbucket server. After all repos are retrieved the http or https clone
// url is extracted. Note that the repo name has the following format: <project key>_-_<repo name>. Thus if you
// just want the repo name you would have to strip the project key and '_-_' separator out.
func (bt *BitbucketCodeHost) listRepos(ctx context.Context, page int, perPage int) ([]*store.Repo, int, error) {
	bt.logger.Debug("fetching repos")

	var outerErr error
	bt.once.Do(func() {
		projects, err := bt.c.ListProjects(ctx)
		if err != nil {
			outerErr = err
		}

		for _, p := range projects {
			if p.Name == bt.def.Path {
				bt.project = p
				break
			}
		}
	})
	if outerErr != nil {
		return nil, 0, outerErr
	}

	if bt.project == nil {
		return nil, 0, errors.Newf("project named %s not found", bt.def.Path)
	}

	repos, next, err := bt.c.ListRepos(ctx, bt.project, page, perPage)
	if err != nil {
		bt.logger.Debug("failed to list repos", log.Error(err))
		return nil, 0, err
	}

	bt.logger.Debug("fetched list of repos", log.Int("repos", len(repos)))

	results := make([]*store.Repo, 0, len(repos))
	for _, r := range repos {
		cloneUrl, err := getCloneUrl(r)
		if err != nil {
			bt.logger.Debug("failed to get clone url", log.String("repo", r.Name), log.String("project", r.Project.Key), log.Error(err))
			return nil, 0, err
		}

		// to be able to push this repo we need to project key, incase we need to create the project before pushing
		results = append(results, &store.Repo{
			Name:   fmt.Sprintf("%s%s%s", r.Project.Key, separator, r.Name),
			GitURL: cloneUrl.String(),
		})
	}

	return results, next, nil
}

func (bt *BitbucketCodeHost) Iterator() Iterator[[]*store.Repo] {
	return bt
}

func (bt *BitbucketCodeHost) Done() bool {
	return bt.done
}

func (bt *BitbucketCodeHost) Err() error {
	return bt.err
}

func (bt *BitbucketCodeHost) Next(ctx context.Context) []*store.Repo {
	if bt.done {
		return nil
	}

	results, next, err := bt.listRepos(ctx, bt.page, bt.perPage)
	if err != nil {
		bt.err = err
		return nil
	}

	// when next is 0, it means the Github api returned the nextPage as 0, which indicates that there are not more pages to fetch
	if next > 0 {
		// Ensure that the next request starts at the next page
		bt.page = next
	} else {
		bt.done = true
	}

	return results
}

func (bt *BitbucketCodeHost) projectKeyAndNameFrom(name string) (string, string) {
	parts := strings.Split(name, separator)
	// If this name originates from a Bitbucket client it will have the format <project key>_-_<repo name>.
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	// The name must originate from some other codehost so now we use the path from the config
	return bt.def.Path, name
}

// CreateRepo creates a repo on bitbucket. It is assumed that the repo name has the following format: <project key>_-_<repo name>.
// A repo can only be created under a project in bitbucket, therefore the project is extract from the repo name format and a
// project is created first, if and only if, the project does not exist already. If the project already exists, the repo
// will be created and the created repos git clone url will be returned.
func (bt *BitbucketCodeHost) CreateRepo(ctx context.Context, name string) (*url.URL, error) {
	key, repoName := bt.projectKeyAndNameFrom(name)

	if len(key) == 0 || len(repoName) == 0 {
		return nil, errors.Errorf("could not extract key and name from unknown repo format %q", name)
	}

	var apiErr *bitbucket.APIError
	_, err := bt.c.GetProjectByKey(ctx, key)
	if err != nil {
		var apiErr *bitbucket.APIError
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
				bt.logger.Debug("created project", log.String("project", p.Key))
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
