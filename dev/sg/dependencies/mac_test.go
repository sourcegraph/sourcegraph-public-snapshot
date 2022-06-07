package dependencies

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

func TestMacOS(t *testing.T) {
	if !strings.Contains(*sgSetupTests, string(OSMac)) && !strings.Contains(*sgSetupTests, "macos") {
		t.Skip("Skipping Mac sg setup tests")
	}

	runner := check.NewRunner(nil, std.NewFixedOutput(os.Stdout, true), Mac)

	err := runner.Check(context.Background(), testArgs)
	assert.Nil(t, err)
}
