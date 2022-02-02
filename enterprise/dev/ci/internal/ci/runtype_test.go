package ci

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunTypeString(t *testing.T) {
	// Check all individual types have a name defined at least
	for rt := PullRequest; rt < None; rt += 1 {
		assert.NotEmpty(t, rt.String(), "RunType: %d", rt)
	}
}
