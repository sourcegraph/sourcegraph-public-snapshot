package images

import (
	"fmt"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/opencontainers/go-digest"

	"github.com/sourcegraph/sourcegraph/dev/ci/images"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var listTagRoute = "https://%s/v2/%s/tags/list"
var fetchDigestRoute = "https://%s/v2/%s/manifests/%s"

// Registry abstracts interacting with various registries, such as
// GCR or Docker.io. There are subtle differences, mostly in how to authenticate.
type Registry interface {
	GetByTag(repo string, tag string) (*Repository, error)
	GetLatest(repo string, latest func(tags []string) (string, error)) (*Repository, error)
	Host() string
	Org() string
	Public() bool
}

type Repository struct {
	registry string
	name     string
	org      string
	tag      string
	digest   digest.Digest
}

func (r *Repository) Ref() string {
	return fmt.Sprintf(
		"%s/%s/%s:%s@%s",
		r.registry,
		r.org,
		r.name,
		r.tag,
		r.digest,
	)
}

func (r *Repository) Name() string {
	return r.name
}

func (r *Repository) Tag() string {
	return r.tag
}

func ParseRepository(rawImg string) (*Repository, error) {
	ref, err := reference.ParseNormalizedNamed(strings.TrimSpace(rawImg))
	if err != nil {
		return nil, err
	}

	r := &Repository{
		registry: reference.Domain(ref),
	}

	if nameTagged, ok := ref.(reference.NamedTagged); ok {
		r.tag = nameTagged.Tag()
		parts := strings.Split(reference.Path(nameTagged), "/")
		if len(parts) != 2 {
			return nil, errors.Newf("failed to parse org/name in %q", reference.Path(nameTagged))
		}
		r.org = parts[0]
		r.name = parts[1]
		if canonical, ok := ref.(reference.Canonical); ok {
			newNamed, err := reference.WithName(canonical.Name())
			if err != nil {
				return nil, err
			}
			newCanonical, err := reference.WithDigest(newNamed, canonical.Digest())
			if err != nil {
				return nil, err
			}
			r.digest = newCanonical.Digest()
		}
	}
	return r, nil
}

type cacheKey struct {
	name string
	tag  string
}

type repositoryCache map[cacheKey]*Repository

func IsSourcegraph(r *Repository) bool {
	// If the container org doesn't contain Sourcegraph, we don't already
	// know it's not ours.
	if !strings.Contains(r.org, "sourcegraph") {
		return false
	}

	// Check our internal images list
	for _, ourImages := range images.SourcegraphDockerImages {
		if strings.HasPrefix(r.name, ourImages) {
			return true
		}
	}
	return false
}
