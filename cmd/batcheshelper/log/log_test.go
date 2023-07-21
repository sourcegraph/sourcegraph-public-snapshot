package log_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/batcheshelper/log"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestLogger_WriteEvent(t *testing.T) {
	var buf bytes.Buffer
	logger := &log.Logger{Writer: &buf}

	err := logger.WriteEvent(
		batcheslib.LogEventOperationTaskStep,
		batcheslib.LogEventStatusStarted,
		&batcheslib.TaskStepMetadata{
			Step: 1,
			Env:  map[string]string{"FOO": "BAR"},
		},
	)
	require.NoError(t, err)

	// Convert to map since there is a timestamp in the content.
	var actual map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, "TASK_STEP", actual["operation"])
	assert.Equal(t, "STARTED", actual["status"])
	assert.Equal(t, float64(1), actual["metadata"].(map[string]interface{})["step"])
	assert.Equal(t, "BAR", actual["metadata"].(map[string]interface{})["env"].(map[string]interface{})["FOO"])
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+Z$`, actual["timestamp"])
}
