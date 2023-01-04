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
	Description() string
	JobName() string
	SetJobName(string)
	RegisterJobLogger(jobLogger *JobLogger)
}

type JobLogger struct {
	rcache   *rcache.Cache
	logger   log.Logger
	routines []Loggable
	hostName string
}

// seenTimeout signifies the maximum time to have no records of a host, job, or routine.
// After this time, we consider them nonexistent, and they'll be removed
const seenTimeout = 5 * 24 * time.Hour // 5 days

const keyPrefix = "background-job-logger:"

// backgroundRoutineForRedis represents a single routine in a background job, and is used for serialization to/from Redis.
type backgroundRoutineForRedis struct {
	Name        string                      `json:"name"`
	Type        types.BackgroundRoutineType `json:"type"`
	JobName     string                      `json:"jobName"`
	Description string                      `json:"description"`
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
// It will save the known job names, host names, and routine names in Redis, along with updating their “last seen” date/time.
func (m *JobLogger) RegistrationDone() {
	// Save/update all known job names
	err := saveKnownJobNames(m.rcache, m.routines)
	if err != nil {
		m.logger.Error("failed to save known job names", log.Error(err))
	}

	// Save/update all known host names
	err = m.rcache.SetHashItem("knownHostNames", m.hostName, time.Now().Format(time.RFC3339))
	if err != nil {
		m.logger.Error("failed to save known host names", log.Error(err))
	}

	// Save/update all known routines
	err = saveKnownRoutines(m.rcache, m.routines)
	if err != nil {
		m.logger.Error("failed to save known routines", log.Error(err))
	}
}

// saveKnownJobNames saves all known job names in Redis, along with updating their “last seen” date/time.
func saveKnownJobNames(c *rcache.Cache, routines []Loggable) error {
	// Collect all job names
	var allJobNames []string
	for _, routine := range routines {
		jobName := routine.JobName()
		if slices.Contains(allJobNames, jobName) {
			continue
		}
		allJobNames = append(allJobNames, jobName)
	}

	// Save/update job names
	for _, jobName := range allJobNames {
		err := c.SetHashItem("knownJobNames", jobName, time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}
	}

	return nil
}

func saveKnownRoutines(c *rcache.Cache, routines []Loggable) error {
	for _, r := range routines {
		routine := backgroundRoutineForRedis{
			Name:        r.Name(),
			Type:        r.Type(),
			JobName:     r.JobName(),
			Description: r.Description(),
			LastSeen:    time.Now().Format(time.RFC3339),
		}

		// Serialize Routine
		routineJson, err := json.Marshal(routine)
		if err != nil {
			return err
		}

		// Save/update Routine
		err = c.SetHashItem("knownRoutines", routine.Name, string(routineJson))
		if err != nil {
			return err
		}
	}

	return nil
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

func saveRunStats(c *rcache.Cache, routineName string, durationMs int32, errored bool) error {
	// Prepare data
	isoDate := time.Now().Format("2006-01-02")
	errorCount := 0
	if errored {
		errorCount = 1
	}

	// Get stats and update them
	lastStatsRaw, found := c.Get(routineName + ":runStats:" + isoDate)
	var newStats types.BackgroundRoutineRunStats
	if found {
		// Read old stats
		var lastStats types.BackgroundRoutineRunStats
		err := json.Unmarshal(lastStatsRaw, &lastStats)
		if err != nil {
			return errors.Wrap(err, "deserialize last stats")
		}

		newStats = types.BackgroundRoutineRunStats{
			Count:         lastStats.Count + 1,
			ErrorCount:    lastStats.ErrorCount + int32(errorCount),
			MinDurationMs: int32(math.Min(float64(lastStats.MinDurationMs), float64(durationMs))),
			AvgDurationMs: int32(math.Max(float64(lastStats.MaxDurationMs), float64(durationMs))),
			MaxDurationMs: int32(math.Round((float64(lastStats.AvgDurationMs)*float64(lastStats.Count) + float64(durationMs)) / float64(lastStats.Count+1))),
		}
	} else {
		newStats = types.BackgroundRoutineRunStats{
			Count:         1,
			ErrorCount:    int32(errorCount),
			MinDurationMs: durationMs,
			AvgDurationMs: durationMs,
			MaxDurationMs: durationMs,
		}
	}

	// Serialize and save updated stats
	updatedStatsJson, err := json.Marshal(newStats)
	if err != nil {
		return errors.Wrap(err, "serialize updated stats")
	}
	c.Set(routineName+":runStats:"+isoDate, updatedStatsJson)

	return nil
}
