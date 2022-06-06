package dependencies

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

func TestMacOS(t *testing.T) {
	var output strings.Builder
	runner := NewRunner("darwin", check.IO{
		Input:  nil,
		Output: std.NewFixedOutput(&output, false),
	})

	err := runner.Check(context.Background(), CheckArgs{})
	assert.Nil(t, err)

	t.Logf(output.String())
}
