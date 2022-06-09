package dependencies

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

func TestUbuntuFix(t *testing.T) {
	if !strings.Contains(*sgSetupTests, string(OSUbuntu)) {
		t.Skip("Skipping Ubuntu sg setup tests")
	}

	runner := check.NewRunner(nil, std.NewSimpleOutput(os.Stdout, true), Ubuntu)
	ctx := context.Background()

	// automatically fix everything!
	err := runner.Fix(ctx, testArgs)
	require.Nil(t, err)

	// now check that everything was fixed
	err = runner.Check(ctx, testArgs)
	assert.Nil(t, err)
}
