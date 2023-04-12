package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/batcheshelper/util"
)

func TestStepJSONFile(t *testing.T) {
	actual := util.StepJSONFile(0)
	assert.Equal(t, "step0.json", actual)
}

func TestFilesMountPath(t *testing.T) {
	actual := util.FilesMountPath("/tmp", 0)
	assert.Equal(t, "/tmp/step0files", actual)
}
