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

func TestMacFix(t *testing.T) {
	if !strings.Contains(*sgSetupTests, string(OSMac)) && !strings.Contains(*sgSetupTests, "macos") {
		t.Skip("Skipping Mac sg setup tests")
	}

	runner := check.NewRunner(nil, std.NewSimpleOutput(os.Stdout, true), Mac)
	ctx := context.Background()

	// automatically fix everything!
	err := runner.Fix(ctx, testArgs)
	require.Nil(t, err)

	// now check that everything was fixed
	err = runner.Check(ctx, testArgs)
	assert.Nil(t, err)
}
