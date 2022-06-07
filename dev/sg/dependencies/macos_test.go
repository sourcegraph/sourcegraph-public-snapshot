package dependencies

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

func TestMacOS(t *testing.T) {
	runner := check.NewRunner(nil, std.NewFixedOutput(os.Stdout, false), MacOS)

	err := runner.Check(context.Background(), CheckArgs{
		InRepo:   true,
		Teammate: false,
	})
	assert.Nil(t, err)
}
