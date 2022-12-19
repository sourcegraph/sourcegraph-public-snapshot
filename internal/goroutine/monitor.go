package goroutine

import (
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"k8s.io/utils/strings/slices"
)

type Monitorable interface {
	BackgroundRoutine
	Name() string
	Type() string
	Description() string
	JobName() string
	SetJobName(string)
	RegisterMonitor(monitor *RedisMonitor)
}

type RedisMonitor struct {
	rcache   *rcache.Cache
	logger   log.Logger
	routines []*Monitorable
	hostName string
}

const keyPrefix = "background-routine-monitor:"

// backgroundRoutineForRedis represents a single routine in a background job, and is used for serialization to/from Redis.
type backgroundRoutineForRedis struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	JobName     string `json:"jobName"`
	Description string `json:"description"`
	LastSeen    string `json:"lastSeen"`
}

func NewRedisMonitor(logger log.Logger) *RedisMonitor {
	cache := rcache.NewWithTTL(keyPrefix, 604800)
	return &RedisMonitor{rcache: cache, logger: logger, hostName: env.MyName}
}

func (m *RedisMonitor) Register(r *Monitorable) {
	m.routines = append(m.routines, r)
}

func (m *RedisMonitor) RegistrationDone() {
	// Save/update all known job names
	err := saveKnownJobNames(m.rcache, m.routines)
	if err != nil {
		m.logger.Error("failed to save known job names", log.Error(err))
	}

	// Save/update all known host names
	err = m.rcache.SetHashItem("knownHostNames", m.hostName, time.Now().String())
	if err != nil {
		m.logger.Error("failed to save known host names", log.Error(err))
	}

	// Save/update all known routines
	err = saveKnownRoutines(m.rcache, m.routines)
	if err != nil {
		m.logger.Error("failed to save known routines", log.Error(err))
	}
}

func saveKnownJobNames(c *rcache.Cache, routines []*Monitorable) error {
	// Collect all job names
	var allJobNames []string
	for _, routine := range routines {
		jobName := (*routine).JobName()
		if slices.Contains(allJobNames, jobName) {
			continue
		}
		allJobNames = append(allJobNames, jobName)
	}

	// Save/update job names
	for _, jobName := range allJobNames {
		err := c.SetHashItem("knownJobNames", jobName, time.Now().String())
		if err != nil {
			return err
		}
	}

	return nil
}

// GetKnownJobNames returns a list of all known job names, ascending.
func GetKnownJobNames(c *rcache.Cache) ([]string, error) {
	jobNames, err := c.GetHashAll("knownJobNames")
	if err != nil {
		return nil, err
	}

	// Get the values only from the map
	var values []string
	for _, value := range jobNames {
		values = append(values, value)
	}

	// Sort the values
	sort.Strings(values)

	return values, nil
}

func GetKnownHostNames(c *rcache.Cache) ([]string, error) {
	hostNames, err := c.GetHashAll("knownHostNames")
	if err != nil {
		return nil, err
	}

	// Get the values only from the map
	var values []string
	for _, value := range hostNames {
		values = append(values, value)

	}

	return values, nil
}

func saveKnownRoutines(c *rcache.Cache, routines []*Monitorable) error {
	for _, r := range routines {
		routine := backgroundRoutineForRedis{
			Name:        (*r).Name(),
			Type:        (*r).Type(),
			JobName:     (*r).JobName(),
			Description: (*r).Description(),
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

func getKnownRoutinesForJob(c *rcache.Cache, jobName string) ([]backgroundRoutineForRedis, error) {
	routines, err := getKnownRoutines(c)
	if err != nil {
		return nil, err
	}

	var routinesForJob []backgroundRoutineForRedis
	for _, routine := range routines {
		if routine.JobName == jobName {
			routinesForJob = append(routinesForJob, routine)
		}
	}

	return routinesForJob, nil
}

func getKnownRoutines(c *rcache.Cache) ([]backgroundRoutineForRedis, error) {
	rawItems, err := c.GetHashAll("knownRoutines")
	if err != nil {
		return nil, err
	}

	routines := make([]backgroundRoutineForRedis, 0, len(rawItems))
	for _, rawItem := range rawItems {
		var item backgroundRoutineForRedis
		err := json.Unmarshal([]byte(rawItem), &item)
		if err != nil {
			return nil, err
		}
		routines = append(routines, item)
	}
	return routines, nil
}

func (m *RedisMonitor) LogRun(r Monitorable, duration time.Duration, runErr error) {
	durationMs := int32(duration / time.Millisecond)
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

func loadRecentRuns(c *rcache.Cache, routineName string, hostName string, count int32) ([]types.BackgroundRoutineRun, error) {
	recentRuns, err := c.GetLastListItems(routineName+":"+hostName+":"+"recentRuns", count)
	if err != nil {
		return nil, errors.Wrap(err, "load recent runs")
	}

	runs := make([]types.BackgroundRoutineRun, 0, len(recentRuns))
	for _, serializedRun := range recentRuns {
		var run types.BackgroundRoutineRun
		err := json.Unmarshal([]byte(serializedRun), &run)
		if err != nil {
			return nil, errors.Wrap(err, "deserialize run")
		}
		runs = append(runs, run)
	}

	return runs, nil
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

func loadRunStats(c *rcache.Cache, now time.Time, dayCount int32) (types.BackgroundRoutineRunStats, error) {
	// Get all stats
	var stats types.BackgroundRoutineRunStats
	for i := int32(0); i < dayCount; i++ {
		isoDate := now.AddDate(0, 0, -int(i)).Format("2006-01-02")
		statsRaw, found := c.Get("runStats:" + isoDate)
		if found {
			var statsForDay types.BackgroundRoutineRunStats
			err := json.Unmarshal(statsRaw, &statsForDay)
			if err != nil {
				return types.BackgroundRoutineRunStats{}, errors.Wrap(err, "deserialize stats for day")
			}

			stats.MinDurationMs = int32(math.Min(float64(stats.MinDurationMs), float64(statsForDay.MinDurationMs)))
			stats.MaxDurationMs = int32(math.Max(float64(stats.MaxDurationMs), float64(statsForDay.MaxDurationMs)))
			stats.AvgDurationMs = int32(math.Round((float64(stats.AvgDurationMs)*float64(stats.Count) + float64(statsForDay.AvgDurationMs)) / float64(stats.Count+statsForDay.Count)))
			stats.Count += statsForDay.Count // Do this after calculating the average
			stats.ErrorCount += statsForDay.ErrorCount
		}
	}

	return stats, nil
}

func (m *RedisMonitor) LogStart(r Monitorable) {
	m.rcache.Set(r.Name()+":"+m.hostName+":"+"lastStart", []byte(time.Now().Format(time.RFC3339)))
	m.logger.Debug("" + r.Name() + " just started! 🚀")
}

func (m *RedisMonitor) LogStop(r Monitorable) {
	m.rcache.Set(r.Name()+":"+m.hostName+":"+"lastStop", []byte(time.Now().Format(time.RFC3339)))
	m.logger.Debug("" + r.Name() + " just stopped! 🛑")
}

func GetBackgroundJobInfos(c *rcache.Cache, after string, recentRunCount int32, dayCountForStats int32) ([]types.BackgroundJobInfo, error) {
	// Get known job names sorted by name, ascending
	knownJobNames, err := GetKnownJobNames(c)
	if err != nil {
		return nil, errors.Wrap(err, "get known job names")
	}

	// Get all jobs
	jobs := make([]types.BackgroundJobInfo, 0, len(knownJobNames))
	for _, jobName := range knownJobNames {
		job, err := GetBackgroundJobInfo(c, jobName, recentRunCount, dayCountForStats)
		if err != nil {
			return nil, errors.Wrap(err, "get job info for "+jobName)
		}
		jobs = append(jobs, job)
	}

	// Filter jobs by name to respect "after" (they are ordered by name)
	if after != "" {
		for i, job := range jobs {
			if job.Name > after {
				jobs = jobs[i:]
				break
			}
		}
	}

	return jobs, nil
}

func GetBackgroundJobInfo(c *rcache.Cache, jobName string, recentRunCount int32, dayCountForStats int32) (types.BackgroundJobInfo, error) {
	allHostNames, err := GetKnownHostNames(c)
	if err != nil {
		return types.BackgroundJobInfo{}, err
	}

	routines, err := getKnownRoutinesForJob(c, jobName)
	if err != nil {
		return types.BackgroundJobInfo{}, err
	}

	routineInfos := make([]types.BackgroundRoutineInfo, 0, len(routines))
	for _, routine := range routines {
		routineInfo, err := getRoutineInfo(c, routine, allHostNames, recentRunCount, dayCountForStats)
		if err != nil {
			return types.BackgroundJobInfo{}, err
		}

		routineInfos = append(routineInfos, routineInfo)
	}

	return types.BackgroundJobInfo{ID: jobName, Name: jobName, Routines: routineInfos}, nil
}

func getRoutineInfo(c *rcache.Cache, routine backgroundRoutineForRedis, allHostNames []string, recentRunCount int32, dayCountForStats int32) (types.BackgroundRoutineInfo, error) {
	routineInfo := types.BackgroundRoutineInfo{
		Name:        routine.Name,
		Type:        routine.Type,
		JobName:     routine.JobName,
		Description: routine.Description,
		Instances:   make([]types.BackgroundRoutineInstanceInfo, 0, len(allHostNames)),
		RecentRuns:  []types.BackgroundRoutineRun{},
		Stats:       types.BackgroundRoutineRunStats{},
	}

	// Collect instances
	for _, hostName := range allHostNames {
		instanceInfo, err := getRoutineInstanceInfo(c, routine.Name, hostName)
		if err != nil {
			return types.BackgroundRoutineInfo{}, err
		}

		routineInfo.Instances = append(routineInfo.Instances, instanceInfo)
	}

	// Collect recent runs
	for _, hostName := range allHostNames {
		recentRunsForHost, err := loadRecentRuns(c, routine.Name, hostName, recentRunCount)
		if err != nil {
			return types.BackgroundRoutineInfo{}, err
		}

		routineInfo.RecentRuns = append(routineInfo.RecentRuns, recentRunsForHost...)
	}

	// Sort recent runs
	sort.Slice(routineInfo.RecentRuns, func(i, j int) bool {
		return routineInfo.RecentRuns[i].At.After(routineInfo.RecentRuns[j].At)
	})
	// Limit to recentRunCount
	if len(routineInfo.RecentRuns) > int(recentRunCount) {
		routineInfo.RecentRuns = routineInfo.RecentRuns[:recentRunCount]
	}

	// Collect stats
	stats, err := loadRunStats(c, time.Now(), dayCountForStats)
	if err != nil {
		return types.BackgroundRoutineInfo{}, errors.Wrap(err, "load run stats")
	}
	routineInfo.Stats = stats

	return routineInfo, nil
}

func getRoutineInstanceInfo(c *rcache.Cache, routineName string, hostName string) (types.BackgroundRoutineInstanceInfo, error) {
	var lastStart time.Time
	var lastStop time.Time
	var err error

	lastStartBytes, ok := c.Get(routineName + ":" + hostName + ":" + "lastStart")
	if ok {
		lastStart, err = time.Parse(time.RFC3339, string(lastStartBytes))
		if err != nil {
			return types.BackgroundRoutineInstanceInfo{}, errors.Wrap(err, "parse last start")
		}
	}

	lastStopBytes, ok := c.Get(routineName + ":" + hostName + ":" + "lastStop")
	if ok {
		lastStop, err = time.Parse(time.RFC3339, string(lastStopBytes))
		if err != nil {
			return types.BackgroundRoutineInstanceInfo{}, errors.Wrap(err, "parse last stop")
		}
	}

	return types.BackgroundRoutineInstanceInfo{
		HostName:      hostName,
		LastStartedAt: &lastStart,
		LastStoppedAt: &lastStop,
	}, nil
}

func GetMonitorCache() *rcache.Cache {
	return rcache.NewWithTTL(keyPrefix, 604800)
}
