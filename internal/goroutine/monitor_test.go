package goroutine

import (
	"context"
	"testing"

	"cuelang.org/go/pkg/time"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/stretchr/testify/assert"
)

func TestGetBackgroundJobInfos(t *testing.T) {
	rcache.SetupForTest(t)

	// Create monitor
	c := GetMonitorCache(1)
	monitor := NewRedisMonitor(log.NoOp(), "test", c)

	// Create and register routines
	ctx := context.Background()
	emptyHandler := HandlerFunc(func(ctx context.Context) error { return nil })
	routine1 := NewPeriodicGoroutine(ctx, "routine1", "a routine", 2*time.Minute, emptyHandler)
	routine1.SetJobName("job-1")
	monitor.Register(routine1)
	routine2 := NewPeriodicGoroutine(ctx, "routine2", "another routine", 2*time.Minute, emptyHandler)
	routine2.SetJobName("job-1")
	monitor.Register(routine2)
	routine3 := NewPeriodicGoroutine(ctx, "routine3", "a third routine", 2*time.Minute, emptyHandler)
	routine3.SetJobName("job-2")
	monitor.Register(routine3)
	monitor.RegistrationDone()

	// Get infos
	jobInfos, err := GetBackgroundJobInfos(c, "", 5, 7)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, jobInfos, 2)
	assert.Equal(t, "job-1", jobInfos[0].ID)
	assert.Equal(t, "job-1", jobInfos[0].Name)
	assert.Equal(t, 2, len(jobInfos[0].Routines))
	assert.Equal(t, "routine1", jobInfos[0].Routines[0].Name)
	assert.Equal(t, "a routine", jobInfos[0].Routines[0].Description)
	assert.Equal(t, 1, len(jobInfos[0].Routines[0].Instances))
	assert.Equal(t, "test", jobInfos[0].Routines[0].Instances[0].HostName)
	assert.Nil(t, jobInfos[0].Routines[0].Instances[0].LastStartedAt)
	assert.Nil(t, jobInfos[0].Routines[0].Instances[0].LastStoppedAt)
	assert.Equal(t, 0, len(jobInfos[0].Routines[0].RecentRuns))
	assert.Equal(t, int32(0), jobInfos[0].Routines[0].Stats.Count)
	assert.Equal(t, int32(0), jobInfos[0].Routines[0].Stats.ErrorCount)
}
