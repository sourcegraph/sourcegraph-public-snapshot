package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/batcheshelper/util"
)

func TestStepJSONFile(t *testing.T) {
	actual := util.StepJSONFile(0)
	assert.Equal(t, "step0.json", actual)
}

func TestFilesMountPath(t *testing.T) {
	actual := util.FilesMountPath("/tmp", 0)
	assert.Equal(t, "/tmp/step0files", actual)
}

func TestWriteSkipFile(t *testing.T) {
	wd := t.TempDir()
	err := util.WriteSkipFile(wd, 1)
	require.NoError(t, err)

	dir, err := os.ReadDir(wd)
	require.NoError(t, err)
	require.Len(t, dir, 1)
	assert.Equal(t, "skip.json", dir[0].Name())
	b, err := os.ReadFile(filepath.Join(wd, "skip.json"))
	require.NoError(t, err)
	assert.JSONEq(t, `{"nextStep": "step.1.pre"}`, string(b))
}

func TestWriteSkipFile_MultipleWrites(t *testing.T) {
	wd := t.TempDir()
	err := util.WriteSkipFile(wd, 1)
	require.NoError(t, err)

	dir, err := os.ReadDir(wd)
	require.NoError(t, err)
	require.Len(t, dir, 1)
	require.Equal(t, "skip.json", dir[0].Name())
	b, err := os.ReadFile(filepath.Join(wd, "skip.json"))
	require.NoError(t, err)
	assert.JSONEq(t, `{"nextStep": "step.1.pre"}`, string(b))

	err = util.WriteSkipFile(wd, 2)
	require.NoError(t, err)

	dir, err = os.ReadDir(wd)
	require.NoError(t, err)
	require.Len(t, dir, 1)
	require.Equal(t, "skip.json", dir[0].Name())
	b, err = os.ReadFile(filepath.Join(wd, "skip.json"))
	require.NoError(t, err)
	assert.JSONEq(t, `{"nextStep": "step.2.pre"}`, string(b))
}
