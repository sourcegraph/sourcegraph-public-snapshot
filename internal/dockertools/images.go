package dockertools

import (
	"fmt"
	"strings"
)

// ImageReference represents a fully qualified Docker image:
// e.g. index.docker.io/sourcegraph/frontend:insiders@sha256:7173b809ca12ec5dee4506cd86be934c4596dd234ee82c0662eac04a8c2c71dc
type ImageReference struct {
	Registry  string // index.docker.io
	Namespace string // sourcegraph
	Name      string // frontend
	Tag       string // insiders
	Digest    string // sha256:7173b809ca12ec5dee4506cd86be934c4596dd234ee82c0662eac04a8c2c71dc
}

// IsPublicDockerHub determines if the container is coming from a public docker registry
// upstream, either by not specifying a registry or speciying one of the Docker Hub URLs
func (i ImageReference) IsPublicDockerHub() bool {
	return i.Registry == "" || i.Registry == "index.docker.io" || i.Registry == "docker.io"
}

func (image ImageReference) String() string {
	s := image.Name
	if image.Namespace != "" {
		s = fmt.Sprintf("%s/%s", image.Namespace, s)
	}
	if image.Tag != "" {
		s = fmt.Sprintf("%s:%s", s, image.Tag)
	}
	if image.Digest != "" {
		s = fmt.Sprintf("%s@%s", s, image.Digest)
	}
	if image.Registry != "" {
		s = fmt.Sprintf("%s/%s", image.Registry, s)
	}

	return s
}

// ParseImageString parses a docker image into a Go ImageReference type in the
// format: <registry.example.com/subpaths>/<namespace>/<name>:<tag>@<digest>
//
// Supports registries with a variable number of subpaths included in them.
// Please note: this is a best-effort string parsing implementation and may
// struggle with certain registries that allow variable formats (e.g. GitLab)
// Please validate the results manually before trusting this function.
func ParseImageString(dockerName string) ImageReference {
	var image ImageReference

	// Check if the input string contains a "digest" tag (denoted by an "@" symbol after the tag).
	if strings.Contains(dockerName, "@") {
		// Extract the digest from the input string and trim it from the input string.
		atIdx := strings.LastIndex(dockerName, "@")
		image.Digest = dockerName[atIdx+1:]
		dockerName = dockerName[:atIdx]
	}

	// Split the input string into its parts using the last colon as a delimiter.
	colonIdx := strings.LastIndex(dockerName, ":")

	if colonIdx != -1 {
		// Extract the tag from the input string.
		image.Tag = dockerName[colonIdx+1:]
		dockerName = dockerName[:colonIdx]
	}

	parts := strings.Split(dockerName, "/")

	// If the input string does not contain any slashes, assume that the first part is the image name.
	if len(parts) == 1 {
		image.Name = parts[0]
		return image
	}

	// If the input string contains at least one slash, assume that the first part is the registry (if it contains a dot),
	// and the last two parts are the image namespace and name.
	// If the input string contains two or more slashes, assume that the first parts (up to the third slash) are the registry,
	// the fourth part is the region (if present), and the last two parts are the image namespace and name.
	image.Name = parts[len(parts)-1]
	image.Namespace = parts[len(parts)-2]
	if len(parts) > 2 {
		if strings.Contains(parts[0], ".") {
			image.Registry = strings.Join(parts[:len(parts)-2], "/")
		} else if len(parts) > 3 {
			image.Registry = strings.Join(parts[:3], "/")
			if len(parts) > 4 {
				image.Registry += "/" + parts[3]
			}
		}
	}

	return image
}
