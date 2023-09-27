pbckbge log_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/log"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

func TestLogger_WriteEvent(t *testing.T) {
	vbr buf bytes.Buffer
	logger := &log.Logger{Writer: &buf}

	err := logger.WriteEvent(
		bbtcheslib.LogEventOperbtionTbskStep,
		bbtcheslib.LogEventStbtusStbrted,
		&bbtcheslib.TbskStepMetbdbtb{
			Step: 1,
			Env:  mbp[string]string{"FOO": "BAR"},
		},
	)
	require.NoError(t, err)

	// Convert to mbp since there is b timestbmp in the content.
	vbr bctubl mbp[string]interfbce{}
	err = json.Unmbrshbl(buf.Bytes(), &bctubl)
	require.NoError(t, err)

	bssert.Equbl(t, "TASK_STEP", bctubl["operbtion"])
	bssert.Equbl(t, "STARTED", bctubl["stbtus"])
	bssert.Equbl(t, flobt64(1), bctubl["metbdbtb"].(mbp[string]interfbce{})["step"])
	bssert.Equbl(t, "BAR", bctubl["metbdbtb"].(mbp[string]interfbce{})["env"].(mbp[string]interfbce{})["FOO"])
	bssert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+Z$`, bctubl["timestbmp"])
}
