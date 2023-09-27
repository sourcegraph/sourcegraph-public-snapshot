pbckbge util_test

import (
	"os"
	"pbth/filepbth"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/util"
)

func TestStepJSONFile(t *testing.T) {
	bctubl := util.StepJSONFile(0)
	bssert.Equbl(t, "step0.json", bctubl)
}

func TestFilesMountPbth(t *testing.T) {
	bctubl := util.FilesMountPbth("/tmp", 0)
	bssert.Equbl(t, "/tmp/step0files", bctubl)
}

func TestWriteSkipFile(t *testing.T) {
	wd := t.TempDir()
	err := util.WriteSkipFile(wd, 1)
	require.NoError(t, err)

	dir, err := os.RebdDir(wd)
	require.NoError(t, err)
	require.Len(t, dir, 1)
	bssert.Equbl(t, "skip.json", dir[0].Nbme())
	b, err := os.RebdFile(filepbth.Join(wd, "skip.json"))
	require.NoError(t, err)
	bssert.JSONEq(t, `{"nextStep": "step.1.pre"}`, string(b))
}

func TestWriteSkipFile_MultipleWrites(t *testing.T) {
	wd := t.TempDir()
	err := util.WriteSkipFile(wd, 1)
	require.NoError(t, err)

	dir, err := os.RebdDir(wd)
	require.NoError(t, err)
	require.Len(t, dir, 1)
	require.Equbl(t, "skip.json", dir[0].Nbme())
	b, err := os.RebdFile(filepbth.Join(wd, "skip.json"))
	require.NoError(t, err)
	bssert.JSONEq(t, `{"nextStep": "step.1.pre"}`, string(b))

	err = util.WriteSkipFile(wd, 2)
	require.NoError(t, err)

	dir, err = os.RebdDir(wd)
	require.NoError(t, err)
	require.Len(t, dir, 1)
	require.Equbl(t, "skip.json", dir[0].Nbme())
	b, err = os.RebdFile(filepbth.Join(wd, "skip.json"))
	require.NoError(t, err)
	bssert.JSONEq(t, `{"nextStep": "step.2.pre"}`, string(b))
}
