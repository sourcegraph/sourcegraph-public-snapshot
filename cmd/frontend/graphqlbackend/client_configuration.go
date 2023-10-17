package graphqlbackend

import (
	"context"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
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
	services, err := r.db.ExternalServices().List(ctx, database.ExternalServicesListOptions{
		Kinds: []string{
			extsvc.KindGitHub,
			extsvc.KindBitbucketServer,
			extsvc.KindGitLab,
			extsvc.KindPhabricator,
		},
	})
	if err != nil {
		return nil, err
	}

	urlMap := make(map[string]struct{})
	for _, service := range services {
		rawConfig, err := service.Config.Decrypt(ctx)
		if err != nil {
			return nil, err
		}
		var url string
		switch service.Kind {
		case extsvc.KindGitHub:
			var ghConfig schema.GitHubConnection
			err = jsonc.Unmarshal(rawConfig, &ghConfig)
			url = ghConfig.Url
		case extsvc.KindBitbucketServer:
			var bbsConfig schema.BitbucketServerConnection
			err = jsonc.Unmarshal(rawConfig, &bbsConfig)
			url = bbsConfig.Url
		case extsvc.KindGitLab:
			var glConfig schema.GitLabConnection
			err = jsonc.Unmarshal(rawConfig, &glConfig)
			url = glConfig.Url
		case extsvc.KindPhabricator:
			var phConfig schema.PhabricatorConnection
			err = jsonc.Unmarshal(rawConfig, &phConfig)
			url = phConfig.Url
		}
		if err != nil {
			return nil, err
		}
		urlMap[url] = struct{}{}
	}

	contentScriptUrls := make([]string, 0, len(urlMap))
	for k := range urlMap {
		contentScriptUrls = append(contentScriptUrls, k)
	}

	cfg := conf.Get()
	var parentSourcegraph parentSourcegraphResolver
	if cfg.ParentSourcegraph != nil {
		parentSourcegraph.url = cfg.ParentSourcegraph.Url
	}

	return &clientConfigurationResolver{
		contentScriptUrls: contentScriptUrls,
		parentSourcegraph: &parentSourcegraph,
	}, nil
}

// stripPassword strips the password from u if it can be parsed as a URL.
// If not, it is left unchanged
// This is a modified version of stringPassword from the standard lib
// in net/http/client.go
func stripPassword(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		return s
	}
	_, passSet := u.User.Password()
	if passSet {
		return strings.Replace(u.String(), u.User.String()+"@", u.User.Username()+":***@", 1)
	}
	return s
}
