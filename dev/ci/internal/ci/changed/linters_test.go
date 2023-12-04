package changed

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/dev/sg/linters"
)

func TestGetLinterTargets(t *testing.T) {
	lintTargets := make(map[string]bool)
	for _, target := range linters.Targets {
		lintTargets[target.Name] = true
	}

	targets := GetLinterTargets(All)
	assert.NotZero(t, len(targets))

	for _, target := range targets {
		if _, exists := lintTargets[target]; !exists {
			t.Errorf("target %q is not a lint target", target)
		}
	}
}
