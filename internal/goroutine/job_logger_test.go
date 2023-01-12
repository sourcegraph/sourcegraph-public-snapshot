package goroutine

import (
	"context"
	"testing"

	"cuelang.org/go/pkg/time"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/assert"
)

// TestLoggerAndReaderHappyPaths tests pretty much everything in the happy path of both the logger and the log reader.
func TestLoggerAndReaderHappyPaths(t *testing.T) {
	rcache.SetupForTest(t)

	// Create logger
	c := GetLoggerCache(1)
	jobLogger := NewJobLogger(log.NoOp(), "test", c)

	// Create and register routines
	// - job-1 has two routines: routine-1 and routine-2
	// - job-2 has one routine: routine-3
	ctx := context.Background()
	emptyHandler := HandlerFunc(func(ctx context.Context) error { return nil })
	routine1 := NewPeriodicGoroutine(ctx, "routine-1", "a routine", 2*time.Minute, emptyHandler)
	routine1.SetJobName("job-1")
	jobLogger.Register(routine1)
	routine2 := NewPeriodicGoroutine(ctx, "routine-2", "another routine", 2*time.Minute, emptyHandler)
	routine2.SetJobName("job-1")
	jobLogger.Register(routine2)
	routine3 := NewPeriodicGoroutine(ctx, "routine-3", "a third routine", 2*time.Minute, emptyHandler)
	routine3.SetJobName("job-2")
	jobLogger.Register(routine3)
	jobLogger.RegistrationDone()

	// Get infos
	jobInfos, err := GetBackgroundJobInfos(c, "", 5, 7)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, jobInfos, 2)
	assert.Equal(t, "job-1", jobInfos[0].ID)
	assert.Equal(t, "job-1", jobInfos[0].Name)
	assert.Equal(t, 2, len(jobInfos[0].Routines))
	assert.Equal(t, "a routine", jobInfos[0].Routines[0].Description)
	assert.Equal(t, 1, len(jobInfos[0].Routines[0].Instances))
	assert.Equal(t, "test", jobInfos[0].Routines[0].Instances[0].HostName)
	assertRoutineStats(t, jobInfos[0].Routines[0], "routine-1", false, false, 0, 0, 0, 0, 0, 0)
	assertRoutineStats(t, jobInfos[0].Routines[1], "routine-2", false, false, 0, 0, 0, 0, 0, 0)
	assertRoutineStats(t, jobInfos[1].Routines[0], "routine-3", false, false, 0, 0, 0, 0, 0, 0)

	// Log some runs: 3x routine-1 (1x with error), 1x routine-2, 0x routine-3
	jobLogger.LogStart(routine1)
	jobLogger.LogStart(routine2)
	jobLogger.LogStart(routine3)
	jobLogger.LogRun(routine1, 10*time.Millisecond, nil)
	jobLogger.LogRun(routine1, 15*time.Millisecond, nil)
	jobLogger.LogRun(routine1, 20*time.Millisecond, errors.New("test error"))
	jobLogger.LogRun(routine2, 2*time.Second, nil)
	jobLogger.LogStop(routine3)

	// Get infos again
	jobInfos, err = GetBackgroundJobInfos(c, "", 5, 7)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, jobInfos, 2)
	assertRoutineStats(t, jobInfos[0].Routines[0], "routine-1", true, false, 3, 3, 1, 10, 15, 20)
	assertRoutineStats(t, jobInfos[0].Routines[1], "routine-2", true, false, 1, 1, 0, 2000, 2000, 2000)
	assertRoutineStats(t, jobInfos[1].Routines[0], "routine-3", true, true, 0, 0, 0, 0, 0, 0)
}

func assertRoutineStats(t *testing.T, r types.BackgroundRoutineInfo, name string,
	started bool, stopped bool, rRuns int, sRuns int32, sErrors int32, sMin int32, sAvg int32, sMax int32) {
	assert.Equal(t, name, r.Name)
	if started {
		assert.NotNil(t, r.Instances[0].LastStartedAt)
	} else {
		assert.Nil(t, r.Instances[0].LastStartedAt)
	}
	assert.Equal(t, rRuns, len(r.RecentRuns))
	assert.Equal(t, sRuns, r.Stats.RunCount)
	assert.Equal(t, sErrors, r.Stats.ErrorCount)
	assert.Equal(t, sMin, r.Stats.MinDurationMs)
	assert.Equal(t, sAvg, r.Stats.AvgDurationMs)
	assert.Equal(t, sMax, r.Stats.MaxDurationMs)
	if stopped {
		assert.NotNil(t, r.Instances[0].LastStoppedAt)
	} else {
		assert.Nil(t, r.Instances[0].LastStoppedAt)
	}
}
