package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
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
	var contentScriptUrls []string

	githubs, err := conf.GitHubConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for _, gh := range githubs {
		contentScriptUrls = append(contentScriptUrls, gh.Url)
	}

	bitbucketservers, err := conf.BitbucketServerConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for _, bb := range bitbucketservers {
		contentScriptUrls = append(contentScriptUrls, bb.Url)
	}

	gitlabs, err := conf.GitLabConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for _, gl := range gitlabs {
		contentScriptUrls = append(contentScriptUrls, gl.Url)
	}

	phabricators, err := conf.PhabricatorConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for _, ph := range phabricators {
		contentScriptUrls = append(contentScriptUrls, ph.Url)
	}

	for _, rb := range cfg.ReviewBoard {
		contentScriptUrls = append(contentScriptUrls, rb.Url)
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
