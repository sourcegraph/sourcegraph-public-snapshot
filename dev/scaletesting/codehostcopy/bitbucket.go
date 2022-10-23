package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/scaletesting/codehostcopy/bitbucket"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type BitbucketCodeHost struct {
	def *CodeHostDefinition
	c   *bitbucket.Client
}

func NewBitbucketCodeHost(ctx context.Context, def *CodeHostDefinition) *BitbucketCodeHost {
	u, _ := url.Parse(def.URL)
	c := bitbucket.NewClient(def.Username, def.Password, u)

	return &BitbucketCodeHost{
		def: def,
		c:   c,
	}
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

func (bt *BitbucketCodeHost) ListRepos(ctx context.Context) ([]*Repo, error) {
	repos := bt.c.ListRepos(ctx)

	results := make([]*Repo, 0, len(repos))
	for i, r := range repos {
		cloneUrl, err := getCloneUrl(r)
		if err != nil {
			log.Printf("WARN: %d: %+v", i, r)
			continue
		}

		// to be able to push this repo we need to project key, incase we need to create the project before pushing
		results = append(results, &Repo{
			name: fmt.Sprintf("%s::%s", r.Project.Key, r.Name),
			url:  cloneUrl.String(),
		})
	}

	return results, nil
}

func (bt *BitbucketCodeHost) CreateRepo(ctx context.Context, name string) (*url.URL, error) {
	parts := strings.Split(name, "::")
	if len(parts) != 2 {
		return nil, errors.New("invalid name format - expected <project key>::<clone url>")
	}
	key := parts[0]
	repoName := parts[1]

	_, err := bt.c.GetProjectByKey(ctx, key)
	if err != nil {
		var apiErr *bitbucket.APIError
		if errors.As(err, &apiErr) {
			// if the project was 'not found' create it
			if apiErr.StatusCode == 404 {
				p, err := bt.c.CreateProject(ctx, &bitbucket.Project{Key: key})
				if err != nil {
					return nil, err
				}
				log.Printf("created project: %+v", p)
			}
		} else {
			return nil, err
		}
	}
	//project already exists so lets just return the url to use
	repo, err := bt.c.CreateRepo(ctx, &bitbucket.Project{Key: key}, repoName)
	if err != nil {
		return nil, err
	}
	log.Printf("created repo %q in project with key %q", repo.Project.Key, repo.Name)
	return getCloneUrl(repo)

}
