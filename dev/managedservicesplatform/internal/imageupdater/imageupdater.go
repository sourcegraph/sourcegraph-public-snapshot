package imageupdater

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Provider interface {
	// ResolveTagAndDigest resolves the tag to the format of "<tag>@<digest>" given the repo.
	// repo - image repo, e.g., "gcr.io/sourcegraph-dev/server"
	// tag - image tag with or without digest, e.g., "3.14.0", "3.14.0@sha256:somehash". If digest is specified, it is returned as it-is.
	//
	// e.g., "gcr.io/sourcegraph-dev/server:3.14.0" -> 3.14.0@sha256:somehash
	ResolveTagAndDigest(repo string, tag string) (string, error)
}

// New returns a new image updater to resolve container tags.
func New() (Provider, error) {
	return &updater{}, nil
}

type updater struct{}

func (*updater) ResolveTagAndDigest(repoStr string, tagStr string) (string, error) {
	refStr := fmt.Sprintf("%s:%s", repoStr, tagStr)

	ref, err := name.ParseReference(refStr)
	if err != nil {
		return "", errors.Wrapf(err, "invalid image ref %q", refStr)
	}

	// if the digest is already specified, we return as it-is and do not attempt to resolve it
	//
	// see https://github.com/google/go-containerregistry/issues/1768
	// if the ref contains both tag & digest, tag is discarded by the parser
	if _, ok := ref.(name.Digest); ok {
		return tagStr, nil
	}

	tag, ok := ref.(name.Tag)
	if !ok {
		return "", errors.Errorf("invalid image ref %q", refStr)
	}

	var opts []remote.Option
	if isGCPRegistry(ref.Context().Registry.Name()) {
		// SECURITY: We use the application default credentials to authenticate with GCR and Artifact Registry.
		opts = append(opts, remote.WithAuthFromKeychain(google.Keychain))
	} else {
		opts = append(opts, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	}

	img, err := remote.Image(ref, opts...)
	if err != nil {
		return "", errors.Wrapf(err, "fetch image %q", ref.String())
	}

	digest, err := img.Digest()
	if err != nil {
		return "", errors.Wrapf(err, "fetch digest %q", ref.String())
	}

	return fmt.Sprintf("%s@%s", tag.TagStr(), digest.String()), nil
}

func isGCPRegistry(registry string) bool {
	return slices.ContainsFunc(
		[]string{"pkg.dev", "gcr.io"},
		func(s string) bool {
			return strings.HasSuffix(registry, s)
		},
	)
}
