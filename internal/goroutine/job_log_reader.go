package goroutine

import (
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetBackgroundJobInfos returns information about all known jobs.
func GetBackgroundJobInfos(c *rcache.Cache, after string, recentRunCount int32, dayCountForStats int32) ([]types.BackgroundJobInfo, error) {
	// Get known job names sorted by name, ascending
	knownJobNames, err := getKnownJobNames(c)
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

// GetBackgroundJobInfo returns information about the given job.
func GetBackgroundJobInfo(c *rcache.Cache, jobName string, recentRunCount int32, dayCountForStats int32) (types.BackgroundJobInfo, error) {
	allHostNames, err := getKnownHostNames(c)
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

// getKnownJobNames returns a list of all known job names, ascending, filtered by their “last seen” time.
func getKnownJobNames(c *rcache.Cache) ([]string, error) {
	jobNames, err := c.GetHashAll("knownJobNames")
	if err != nil {
		return nil, err
	}

	// Get the values only from the map
	var values []string
	for jobName, lastSeenString := range jobNames {
		// Parse “last seen” time
		lastSeen, err := time.Parse(time.RFC3339, lastSeenString)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse job last seen time")
		}

		// Check if job is still running
		if time.Since(lastSeen) > seenTimeout {
			continue
		}

		values = append(values, jobName)
	}

	// Sort the values
	sort.Strings(values)

	return values, nil
}

// getKnownHostNames returns a list of all known host names, ascending, filtered by their “last seen” time.
func getKnownHostNames(c *rcache.Cache) ([]string, error) {
	hostNames, err := c.GetHashAll("knownHostNames")
	if err != nil {
		return nil, err
	}

	// Get the values only from the map
	var values []string
	for hostName, lastSeenString := range hostNames {
		// Parse “last seen” time
		lastSeen, err := time.Parse(time.RFC3339, lastSeenString)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse host last seen time")
		}

		// Check if job is still running
		if time.Since(lastSeen) > seenTimeout {
			continue
		}

		values = append(values, hostName)
	}

	// Sort the values
	sort.Strings(values)

	return values, nil
}

// getKnownRoutinesForJob returns a list of all known routines for the given job name, ascending.
func getKnownRoutinesForJob(c *rcache.Cache, jobName string) ([]backgroundRoutineForRedis, error) {
	// Get all routines
	routines, err := getKnownRoutines(c)
	if err != nil {
		return nil, err
	}

	// Filter by job name
	var routinesForJob []backgroundRoutineForRedis
	for _, routine := range routines {
		if routine.JobName == jobName {
			routinesForJob = append(routinesForJob, routine)
		}
	}

	// Sort them by name
	sort.Slice(routinesForJob, func(i, j int) bool {
		return routinesForJob[i].Name < routinesForJob[j].Name
	})

	return routinesForJob, nil
}

// getKnownRoutines returns a list of all known routines, unfiltered, in no particular order.
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

// getRoutineInfo returns the info for a single routine: its instances, recent runs, and stats.
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
	stats, err := loadRunStats(c, routine.Name, time.Now(), dayCountForStats)
	if err != nil {
		return types.BackgroundRoutineInfo{}, errors.Wrap(err, "load run stats")
	}
	routineInfo.Stats = stats

	return routineInfo, nil
}

// getRoutineInstanceInfo returns the info for a single routine instance.
func getRoutineInstanceInfo(c *rcache.Cache, routineName string, hostName string) (types.BackgroundRoutineInstanceInfo, error) {
	var lastStart *time.Time
	var lastStop *time.Time

	lastStartBytes, ok := c.Get(routineName + ":" + hostName + ":" + "lastStart")
	if ok {
		t, err := time.Parse(time.RFC3339, string(lastStartBytes))
		if err != nil {
			return types.BackgroundRoutineInstanceInfo{}, errors.Wrap(err, "parse last start")
		}
		lastStart = &t
	}

	lastStopBytes, ok := c.Get(routineName + ":" + hostName + ":" + "lastStop")
	if ok {
		t, err := time.Parse(time.RFC3339, string(lastStopBytes))
		if err != nil {
			return types.BackgroundRoutineInstanceInfo{}, errors.Wrap(err, "parse last stop")
		}
		lastStop = &t
	}

	return types.BackgroundRoutineInstanceInfo{
		HostName:      hostName,
		LastStartedAt: lastStart,
		LastStoppedAt: lastStop,
	}, nil
}

// loadRecentRuns loads the recent runs for a routine.
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

// loadRunStats loads the run stats for a routine.
func loadRunStats(c *rcache.Cache, routineName string, now time.Time, dayCount int32) (types.BackgroundRoutineRunStats, error) {
	// Get all stats
	var stats types.BackgroundRoutineRunStats
	for i := int32(0); i < dayCount; i++ {
		isoDate := now.AddDate(0, 0, -int(i)).Format("2006-01-02")
		statsRaw, found := c.Get(routineName + ":runStats:" + isoDate)
		if found {
			var statsForDay types.BackgroundRoutineRunStats
			err := json.Unmarshal(statsRaw, &statsForDay)
			if err != nil {
				return types.BackgroundRoutineRunStats{}, errors.Wrap(err, "deserialize stats for day")
			}

			if stats.Since == nil {
				dayStart, err := time.Parse("2006-01-02", isoDate)
				if err != nil {
					return types.BackgroundRoutineRunStats{}, errors.Wrap(err, "parse day start")
				}
				stats.Since = &dayStart
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
