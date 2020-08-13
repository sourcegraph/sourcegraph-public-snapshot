package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

type clientConfigurationResolver struct {
	contentScriptUrls []string
	parentSourcegraph *parentSourcegraphResolver
}

type parentSourcegraphResolver struct {
	url string
}

func (r *clientConfigurationResolver) ContentScriptURLs() []string {
	return r.contentScriptUrls
}

func (r *clientConfigurationResolver) ParentSourcegraph() *parentSourcegraphResolver {
	return r.parentSourcegraph
}

func (r *parentSourcegraphResolver) URL() string {
	return r.url
}

func (r *schemaResolver) ClientConfiguration(ctx context.Context) (*clientConfigurationResolver, error) {
	cfg := conf.Get()

	// The following code makes serial database calls.
	// Ideally these could be done in parallel, but the table is small
	// and I don't think real world perf is going to be bad.

	// We could have multiple services with the same URL so we dedupe them
	urlMap := make(map[string]struct{})

	githubs, err := conf.GitHubConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for _, gh := range githubs {
		urlMap[gh.Url] = struct{}{}
	}

	bitbucketservers, err := conf.BitbucketServerConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for _, bb := range bitbucketservers {
		urlMap[bb.Url] = struct{}{}
	}

	gitlabs, err := conf.GitLabConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for _, gl := range gitlabs {
		urlMap[gl.Url] = struct{}{}
	}

	phabricators, err := conf.PhabricatorConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for _, ph := range phabricators {
		urlMap[ph.Url] = struct{}{}
	}

	contentScriptUrls := make([]string, 0, len(urlMap))
	for k := range urlMap {
		contentScriptUrls = append(contentScriptUrls, k)
	}

	var parentSourcegraph parentSourcegraphResolver
	if cfg.ParentSourcegraph != nil {
		parentSourcegraph.url = cfg.ParentSourcegraph.Url
	}

	return &clientConfigurationResolver{
		contentScriptUrls: contentScriptUrls,
		parentSourcegraph: &parentSourcegraph,
	}, nil
}
