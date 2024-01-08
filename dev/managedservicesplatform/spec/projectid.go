package spec

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const DefaultSuffixLength = 4

// NewProjectID generates a MSP-standard project ID for a service environment.
func NewProjectID(serviceID, envID string, suffixLength int) (string, error) {
	if suffixLength < 2 {
		return "", errors.New("suffix length must be at least 2 characters long")
	}

	suffix, err := newRandomAlphabeticalString(suffixLength)
	if err != nil {
		return "", errors.Wrap(err, "generate suffix")
	}
	projectID := fmt.Sprintf("%s-%s-%s",
		serviceID, envID, suffix)

	// https://cloud.google.com/resource-manager/reference/rest/v1/projects
	if len(projectID) > 30 {
		return "", errors.Newf("project ID %q must be no longer than 30 characters, try a shorter service ID or environment ID")
	}

	return projectID, nil
}

func newRandomAlphabeticalString(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", errors.Wrap(err, "generate random string")
	}
	// Base 64 can be longer than len
	return hex.EncodeToString(buf)[:length], nil
}
