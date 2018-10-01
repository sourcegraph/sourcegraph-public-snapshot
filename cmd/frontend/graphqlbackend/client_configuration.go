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
	for _, gh := range cfg.Github {
		contentScriptUrls = append(contentScriptUrls, gh.Url)
	}
	for _, bb := range cfg.BitbucketServer {
		contentScriptUrls = append(contentScriptUrls, bb.Url)
	}
	for _, gl := range cfg.Gitlab {
		contentScriptUrls = append(contentScriptUrls, gl.Url)
	}
	for _, ph := range cfg.Phabricator {
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
