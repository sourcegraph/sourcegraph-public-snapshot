package goroutine

import (
	"encoding/json"
	"math"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"k8s.io/utils/strings/slices"
)

type Loggable interface {
	BackgroundRoutine
	Name() string
	Type() types.BackgroundRoutineType
	JobName() string
	SetJobName(string)
	Description() string
	Interval() time.Duration
	RegisterJobLogger(jobLogger *JobLogger)
}

type JobLogger struct {
	rcache   *rcache.Cache
	logger   log.Logger
	routines []Loggable
	hostName string
}

// seenTimeout is the maximum time we allow no activity for each host, job, and routine.
// After this time, we consider them non-existent.
const seenTimeout = 6 * 24 * time.Hour // 6 days

const keyPrefix = "background-job-logger"

// backgroundRoutine represents a single routine in a background job, and is used for serialization to/from Redis.
type backgroundRoutine struct {
	Name        string                      `json:"name"`
	Type        types.BackgroundRoutineType `json:"type"`
	JobName     string                      `json:"jobName"`
	Description string                      `json:"description"`
	Interval    time.Duration               `json:"interval"` // Assumes that the routine runs at a fixed interval across all hosts
	LastSeen    string                      `json:"lastSeen"`
}

// NewJobLogger creates a new job logger.
func NewJobLogger(logger log.Logger, hostName string, cache *rcache.Cache) *JobLogger {
	return &JobLogger{rcache: cache, logger: logger, hostName: hostName}
}

// Register registers a new routine with the job logger.
func (m *JobLogger) Register(r Loggable) {
	m.routines = append(m.routines, r)
}

// RegistrationDone should be called after all routines have been registered.
// It saves the known job names, host names, and routine names in Redis, along with updating their “last seen” date/time.
func (m *JobLogger) RegistrationDone() {
	// Save/update known job names
	for _, jobName := range m.collectAllJobNames() {
		m.saveKnownJobName(jobName)
	}

	// Save known host name
	m.saveKnownHostName()

	// Save/update all known routines
	for _, r := range m.routines {
		m.saveKnownRoutine(r)
	}
}

// collectAllJobNames collects all known job names in Redis.
func (m *JobLogger) collectAllJobNames() []string {
	var allJobNames []string
	for _, routine := range m.routines {
		jobName := routine.JobName()
		if slices.Contains(allJobNames, jobName) {
			continue
		}
		allJobNames = append(allJobNames, jobName)
	}

	return allJobNames
}

// saveKnownJobName updates the “lastSeen” date of a known job in Redis. Also adds it to the list of known jobs if it doesn’t exist.
func (m *JobLogger) saveKnownJobName(jobName string) {
	err := m.rcache.SetHashItem("knownJobNames", jobName, time.Now().Format(time.RFC3339))
	if err != nil {
		m.logger.Error("failed to save/update known job name", log.Error(err), log.String("jobName", jobName))
	}
}

// saveKnownHostName updates the “lastSeen” date of a known host in Redis. Also adds it to the list of known hosts if it doesn’t exist.
func (m *JobLogger) saveKnownHostName() {
	err := m.rcache.SetHashItem("knownHostNames", m.hostName, time.Now().Format(time.RFC3339))
	if err != nil {
		m.logger.Error("failed to save/update known host name", log.Error(err), log.String("hostName", m.hostName))
	}
}

// saveKnownRouting updates the routine in Redis. Also adds it to the list of known routines if it doesn’t exist.
func (m *JobLogger) saveKnownRoutine(loggable Loggable) {
	routine := backgroundRoutine{
		Name:        loggable.Name(),
		Type:        loggable.Type(),
		JobName:     loggable.JobName(),
		Description: loggable.Description(),
		Interval:    loggable.Interval(),
		LastSeen:    time.Now().Format(time.RFC3339),
	}

	// Serialize Routine
	routineJson, err := json.Marshal(routine)
	if err != nil {
		m.logger.Error("failed to serialize routine", log.Error(err), log.String("routineName", routine.Name))
		return
	}

	// Save/update Routine
	err = m.rcache.SetHashItem("knownRoutines", routine.Name, string(routineJson))
	if err != nil {
		m.logger.Error("failed to save/update known routine", log.Error(err), log.String("routineName", routine.Name))
	}
}

func (m *JobLogger) LogStart(r Loggable) {
	m.rcache.Set(r.Name()+":"+m.hostName+":"+"lastStart", []byte(time.Now().Format(time.RFC3339)))
	m.logger.Debug("" + r.Name() + " just started! 🚀")
}

func (m *JobLogger) LogStop(r Loggable) {
	m.rcache.Set(r.Name()+":"+m.hostName+":"+"lastStop", []byte(time.Now().Format(time.RFC3339)))
	m.logger.Debug("" + r.Name() + " just stopped! 🛑")
}

func GetLoggerCache(ttlSeconds int) *rcache.Cache {
	return rcache.NewWithTTL(keyPrefix, ttlSeconds)
}

func (m *JobLogger) LogRun(r Loggable, duration time.Duration, runErr error) {
	durationMs := int32(duration.Milliseconds())
	err := saveRun(m.rcache, r.Name(), m.hostName, durationMs, runErr)
	if err != nil {
		m.logger.Error("failed to save run", log.Error(err))
	}

	err = saveRunStats(m.rcache, r.Name(), durationMs, runErr != nil)
	if err != nil {
		m.logger.Error("failed to save run stats", log.Error(err))
	}

	m.logger.Debug("Hello from " + r.Name() + "! 😄")
}

func saveRun(c *rcache.Cache, routineName string, hostName string, durationMs int32, err error) error {
	errorMessage := ""
	stackTrace := ""
	if err != nil {
		errorMessage = err.Error()
		stackTrace = "stack trace goes here"
	}

	// Create Run
	run := types.BackgroundRoutineRun{
		At:           time.Now(),
		HostName:     hostName,
		DurationMs:   durationMs,
		ErrorMessage: errorMessage,
		StackTrace:   stackTrace,
	}

	// Serialize run
	runJson, err := json.Marshal(run)
	if err != nil {
		return errors.Wrap(err, "serialize run")
	}

	// Save run
	err = c.AddToList(routineName+":"+hostName+":"+"recentRuns", string(runJson))
	if err != nil {
		return errors.Wrap(err, "save run")
	}

	return nil
}

// saveRunStats updates the run stats for a routine in Redis.
func saveRunStats(c *rcache.Cache, routineName string, durationMs int32, errored bool) error {
	// Prepare data
	isoDate := time.Now().Format("2006-01-02")

	// Get stats
	lastStatsRaw, found := c.Get(routineName + ":runStats:" + isoDate)
	var lastStats types.BackgroundRoutineRunStats
	if found {
		err := json.Unmarshal(lastStatsRaw, &lastStats)
		if err != nil {
			return errors.Wrap(err, "deserialize last stats")
		}
	}

	// Update stats
	newStats := addRunToStats(lastStats, durationMs, errored)

	// Serialize and save updated stats
	updatedStatsJson, err := json.Marshal(newStats)
	if err != nil {
		return errors.Wrap(err, "serialize updated stats")
	}
	c.Set(routineName+":runStats:"+isoDate, updatedStatsJson)

	return nil
}

// addRunToStats adds a new run to the stats.
func addRunToStats(stats types.BackgroundRoutineRunStats, durationMs int32, errored bool) types.BackgroundRoutineRunStats {
	errorCount := int32(0)
	if errored {
		errorCount = 1
	}
	return mergeStats(stats, types.BackgroundRoutineRunStats{
		RunCount:      1,
		ErrorCount:    errorCount,
		MinDurationMs: durationMs,
		AvgDurationMs: durationMs,
		MaxDurationMs: durationMs,
	})
}

// mergeStats returns the given stats updated with the given run data.
func mergeStats(a types.BackgroundRoutineRunStats, b types.BackgroundRoutineRunStats) types.BackgroundRoutineRunStats {
	// Calculate earlier "since"
	var since time.Time
	if a.Since != nil && (b.Since != nil && b.Since.Before(*a.Since)) {
		since = *b.Since
	} else if a.Since != nil {
		since = *a.Since
	}
	var sincePtr *time.Time
	if !since.IsZero() {
		sincePtr = &since
	}

	// Calculate durations
	var minDurationMs int32
	if a.MinDurationMs == 0 || b.MinDurationMs < a.MinDurationMs {
		minDurationMs = b.MinDurationMs
	} else {
		minDurationMs = a.MinDurationMs
	}
	avgDurationMs := int32(math.Round(float64(a.AvgDurationMs*a.RunCount+b.AvgDurationMs*b.RunCount) / float64(a.RunCount+b.RunCount)))
	var maxDurationMs int32
	if b.MaxDurationMs > a.MaxDurationMs {
		maxDurationMs = b.MaxDurationMs
	} else {
		maxDurationMs = a.MaxDurationMs
	}

	// Return merged stats
	return types.BackgroundRoutineRunStats{
		Since:         sincePtr,
		RunCount:      a.RunCount + b.RunCount,
		ErrorCount:    a.ErrorCount + b.ErrorCount,
		MinDurationMs: minDurationMs,
		AvgDurationMs: avgDurationMs,
		MaxDurationMs: maxDurationMs,
	}
}
