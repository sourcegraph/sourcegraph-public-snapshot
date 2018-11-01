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
	config := conf.Get().Basic

	var contentScriptUrls []string
	for _, gh := range config.Github {
		contentScriptUrls = append(contentScriptUrls, gh.Url)
	}
	for _, bb := range config.BitbucketServer {
		contentScriptUrls = append(contentScriptUrls, bb.Url)
	}
	for _, gl := range config.Gitlab {
		contentScriptUrls = append(contentScriptUrls, gl.Url)
	}
	for _, ph := range config.Phabricator {
		contentScriptUrls = append(contentScriptUrls, ph.Url)
	}
	for _, rb := range config.ReviewBoard {
		contentScriptUrls = append(contentScriptUrls, rb.Url)
	}

	var parentSourcegraph parentSourcegraphResolver
	if config.ParentSourcegraph != nil {
		parentSourcegraph.url = config.ParentSourcegraph.Url
	}

	return &clientConfigurationResolver{
		contentScriptUrls: contentScriptUrls,
		parentSourcegraph: &parentSourcegraph,
	}, nil
}
