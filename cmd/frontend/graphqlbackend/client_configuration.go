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

	// The following code makes serial database calls.
	// Ideally these could be done in parallel, but the table is small
	// and I don't think real world perf is going to be bad.

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

	var parentSourcegraph parentSourcegraphResolver
	if cfg.ParentSourcegraph != nil {
		parentSourcegraph.url = cfg.ParentSourcegraph.Url
	}

	return &clientConfigurationResolver{
		contentScriptUrls: contentScriptUrls,
		parentSourcegraph: &parentSourcegraph,
	}, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_116(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
