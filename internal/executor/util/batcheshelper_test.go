package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/executor/util"
)

func TestFormatPreKey(t *testing.T) {
	actual := util.FormatPreKey(1)
	assert.Equal(t, "step.1.pre", actual)
}

func TestFormatRunKey(t *testing.T) {
	actual := util.FormatRunKey(1)
	assert.Equal(t, "step.1.run", actual)
}

func TestFormatPostKey(t *testing.T) {
	actual := util.FormatPostKey(1)
	assert.Equal(t, "step.1.post", actual)
}

func TestIsPreStepKey(t *testing.T) {
	actual := util.IsPreStepKey("step.1.pre")
	assert.True(t, actual)
}
