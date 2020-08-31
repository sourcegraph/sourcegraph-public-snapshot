package indexer

import (
	"fmt"
	"regexp"
)

var imagePattern = regexp.MustCompile(`([^:]+)(?::([^@]+))?(?:@sha256:([a-z0-9]{64}))?`)

// sanitizeImage sanitizes the given docker image for use by ignite. The ignite utility has
// some issue parsing docker tags that include _both_ a tag and a sha256 hash. In this case,
// we remove the tag but keep the hash as it uniquely identifies a particular image, whereas
// tags are mutable.
func sanitizeImage(image string) string {
	if matches := imagePattern.FindStringSubmatch(image); len(matches) == 4 && matches[2] != "" && matches[3] != "" {
		return fmt.Sprintf("%s@sha256:%s", matches[1], matches[3])
	}

	return image
}
