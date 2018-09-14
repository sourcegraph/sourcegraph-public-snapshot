package graphqlbackend

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type extensionManifestResolver struct {
	raw string

	// cache result because it is used by multiple fields
	once   sync.Once
	result *schema.SourcegraphExtensionManifest
	err    error
}

func newExtensionManifestResolver(raw *string) *extensionManifestResolver {
	if raw == nil {
		return nil
	}
	return &extensionManifestResolver{raw: *raw}
}

func (r *extensionManifestResolver) parse() (*schema.SourcegraphExtensionManifest, error) {
	r.once.Do(func() {
		r.err = jsonc.Unmarshal(r.raw, &r.result)
	})
	return r.result, r.err
}

func (r *extensionManifestResolver) Raw() string { return r.raw }

func (r *extensionManifestResolver) Title() (*string, error) {
	parsed, err := r.parse()
	if parsed == nil || err != nil {
		return nil, err
	}
	if parsed.Title == "" {
		return nil, nil
	}
	return &parsed.Title, nil
}

func (r *extensionManifestResolver) Description() (*string, error) {
	parsed, err := r.parse()
	if parsed == nil || err != nil {
		return nil, err
	}
	if parsed.Description == "" {
		return nil, nil
	}
	return &parsed.Description, nil
}

func (r *extensionManifestResolver) BundleURL() (*string, error) {
	parsed, err := r.parse()
	if parsed == nil || err != nil {
		return nil, err
	}
	if parsed.Url == "" {
		return nil, nil
	}
	return &parsed.Url, nil
}
