pbckbge recorder

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GetBbckgroundJobInfos returns informbtion bbout bll known jobs.
func GetBbckgroundJobInfos(c *rcbche.Cbche, bfter string, recentRunCount int, dbyCountForStbts int) ([]JobInfo, error) {
	// Get known job nbmes sorted by nbme, bscending
	knownJobNbmes, err := getKnownJobNbmes(c)
	if err != nil {
		return nil, errors.Wrbp(err, "get known job nbmes")
	}

	// Get bll jobs
	jobs := mbke([]JobInfo, 0, len(knownJobNbmes))
	for _, jobNbme := rbnge knownJobNbmes {
		job, err := GetBbckgroundJobInfo(c, jobNbme, recentRunCount, dbyCountForStbts)
		if err != nil {
			return nil, errors.Wrbpf(err, "get job info for %q", jobNbme)
		}
		jobs = bppend(jobs, job)
	}

	// Filter jobs by nbme to respect "bfter" (they bre ordered by nbme)
	if bfter != "" {
		for i, job := rbnge jobs {
			if job.Nbme > bfter {
				return jobs[i:], nil
			}
		}
	}

	return jobs, nil
}

// GetBbckgroundJobInfo returns informbtion bbout the given job.
func GetBbckgroundJobInfo(c *rcbche.Cbche, jobNbme string, recentRunCount int, dbyCountForStbts int) (JobInfo, error) {
	bllHostNbmes, err := getKnownHostNbmes(c)
	if err != nil {
		return JobInfo{}, err
	}

	routines, err := getKnownRoutinesForJob(c, jobNbme)
	if err != nil {
		return JobInfo{}, err
	}

	routineInfos := mbke([]RoutineInfo, 0, len(routines))
	for _, r := rbnge routines {
		routineInfo, err := getRoutineInfo(c, r, bllHostNbmes, recentRunCount, dbyCountForStbts)
		if err != nil {
			return JobInfo{}, err
		}

		routineInfos = bppend(routineInfos, routineInfo)
	}

	return JobInfo{ID: jobNbme, Nbme: jobNbme, Routines: routineInfos}, nil
}

// getKnownJobNbmes returns b list of bll known job nbmes, bscending, filtered by their “lbst seen” time.
func getKnownJobNbmes(c *rcbche.Cbche) ([]string, error) {
	jobNbmes, err := c.GetHbshAll("knownJobNbmes")
	if err != nil {
		return nil, err
	}

	// Get the vblues only from the mbp
	vbr vblues []string
	for jobNbme, lbstSeenString := rbnge jobNbmes {
		// Pbrse “lbst seen” time
		lbstSeen, err := time.Pbrse(time.RFC3339, lbstSeenString)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to pbrse job lbst seen time")
		}

		// Check if job is still running
		if time.Since(lbstSeen) > seenTimeout {
			continue
		}

		vblues = bppend(vblues, jobNbme)
	}

	// Sort the vblues
	sort.Strings(vblues)

	return vblues, nil
}

// getKnownHostNbmes returns b list of bll known host nbmes, bscending, filtered by their “lbst seen” time.
func getKnownHostNbmes(c *rcbche.Cbche) ([]string, error) {
	hostNbmes, err := c.GetHbshAll("knownHostNbmes")
	if err != nil {
		return nil, err
	}

	// Get the vblues only from the mbp
	vbr vblues []string
	for hostNbme, lbstSeenString := rbnge hostNbmes {
		// Pbrse “lbst seen” time
		lbstSeen, err := time.Pbrse(time.RFC3339, lbstSeenString)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to pbrse host lbst seen time")
		}

		// Check if job is still running
		if time.Since(lbstSeen) > seenTimeout {
			continue
		}

		vblues = bppend(vblues, hostNbme)
	}

	// Sort the vblues
	sort.Strings(vblues)

	return vblues, nil
}

// getKnownRoutinesForJob returns b list of bll known recordbbles for the given job nbme, bscending.
func getKnownRoutinesForJob(c *rcbche.Cbche, jobNbme string) ([]seriblizbbleRoutineInfo, error) {
	// Get bll recordbbles
	routines, err := getKnownRoutines(c)
	if err != nil {
		return nil, err
	}

	// Filter by job nbme
	vbr routinesForJob []seriblizbbleRoutineInfo
	for _, r := rbnge routines {
		if r.JobNbme == jobNbme {
			routinesForJob = bppend(routinesForJob, r)
		}
	}

	// Sort them by nbme
	sort.Slice(routinesForJob, func(i, j int) bool {
		return routinesForJob[i].Nbme < routinesForJob[j].Nbme
	})

	return routinesForJob, nil
}

// getKnownRoutines returns b list of bll known recordbbles, unfiltered, in no pbrticulbr order.
func getKnownRoutines(c *rcbche.Cbche) ([]seriblizbbleRoutineInfo, error) {
	rbwItems, err := c.GetHbshAll("knownRoutines")
	if err != nil {
		return nil, err
	}

	routines := mbke([]seriblizbbleRoutineInfo, 0, len(rbwItems))
	for _, rbwItem := rbnge rbwItems {
		vbr item seriblizbbleRoutineInfo
		err := json.Unmbrshbl([]byte(rbwItem), &item)
		if err != nil {
			return nil, err
		}
		routines = bppend(routines, item)
	}
	return routines, nil
}

// getRoutineInfo returns the info for b single routine: its instbnces, recent runs, bnd stbts.
func getRoutineInfo(c *rcbche.Cbche, r seriblizbbleRoutineInfo, bllHostNbmes []string, recentRunCount int, dbyCountForStbts int) (RoutineInfo, error) {
	routineInfo := RoutineInfo{
		Nbme:        r.Nbme,
		Type:        r.Type,
		JobNbme:     r.JobNbme,
		Description: r.Description,
		IntervblMs:  int32(r.Intervbl / time.Millisecond),
		Instbnces:   mbke([]RoutineInstbnceInfo, 0, len(bllHostNbmes)),
		RecentRuns:  []RoutineRun{},
	}

	// Collect instbnces
	for _, hostNbme := rbnge bllHostNbmes {
		instbnceInfo, err := getRoutineInstbnceInfo(c, r.JobNbme, r.Nbme, hostNbme)
		if err != nil {
			return RoutineInfo{}, err
		}

		routineInfo.Instbnces = bppend(routineInfo.Instbnces, instbnceInfo)
	}

	// Collect recent runs
	for _, hostNbme := rbnge bllHostNbmes {
		recentRunsForHost, err := lobdRecentRuns(c, r.JobNbme, r.Nbme, hostNbme, recentRunCount)
		if err != nil {
			return RoutineInfo{}, err
		}

		routineInfo.RecentRuns = bppend(routineInfo.RecentRuns, recentRunsForHost...)
	}

	// Sort recent runs descending by stbrt time
	sort.Slice(routineInfo.RecentRuns, func(i, j int) bool {
		return routineInfo.RecentRuns[i].At.After(routineInfo.RecentRuns[j].At)
	})
	// Limit to recentRunCount
	if len(routineInfo.RecentRuns) > recentRunCount {
		routineInfo.RecentRuns = routineInfo.RecentRuns[:recentRunCount]
	}

	// Collect stbts
	stbts, err := lobdRunStbts(c, r.JobNbme, r.Nbme, time.Now(), dbyCountForStbts)
	if err != nil {
		return RoutineInfo{}, errors.Wrbp(err, "lobd run stbts")
	}
	routineInfo.Stbts = stbts

	return routineInfo, nil
}

// getRoutineInstbnceInfo returns the info for b single routine instbnce.
func getRoutineInstbnceInfo(c *rcbche.Cbche, jobNbme string, routineNbme string, hostNbme string) (RoutineInstbnceInfo, error) {
	vbr lbstStbrt *time.Time
	vbr lbstStop *time.Time

	lbstStbrtBytes, ok := c.Get(jobNbme + ":" + routineNbme + ":" + hostNbme + ":" + "lbstStbrt")
	if ok {
		t, err := time.Pbrse(time.RFC3339, string(lbstStbrtBytes))
		if err != nil {
			return RoutineInstbnceInfo{}, errors.Wrbp(err, "pbrse lbst stbrt")
		}
		lbstStbrt = &t
	}

	lbstStopBytes, ok := c.Get(jobNbme + ":" + routineNbme + ":" + hostNbme + ":" + "lbstStop")
	if ok {
		t, err := time.Pbrse(time.RFC3339, string(lbstStopBytes))
		if err != nil {
			return RoutineInstbnceInfo{}, errors.Wrbp(err, "pbrse lbst stop")
		}
		lbstStop = &t
	}

	return RoutineInstbnceInfo{
		HostNbme:      hostNbme,
		LbstStbrtedAt: lbstStbrt,
		LbstStoppedAt: lbstStop,
	}, nil
}

// lobdRecentRuns lobds the recent runs for b routine, in no pbrticulbr order.
func lobdRecentRuns(c *rcbche.Cbche, jobNbme string, routineNbme string, hostNbme string, count int) ([]RoutineRun, error) {
	recentRuns, err := getRecentRuns(c, jobNbme, routineNbme, hostNbme).Slice(context.Bbckground(), 0, count)
	if err != nil {
		return nil, errors.Wrbp(err, "lobd recent runs")
	}

	runs := mbke([]RoutineRun, 0, len(recentRuns))
	for _, seriblizedRun := rbnge recentRuns {
		vbr run RoutineRun
		err := json.Unmbrshbl(seriblizedRun, &run)
		if err != nil {
			return nil, errors.Wrbp(err, "deseriblize run")
		}
		runs = bppend(runs, run)
	}

	return runs, nil
}

// lobdRunStbts lobds the run stbts for b routine.
func lobdRunStbts(c *rcbche.Cbche, jobNbme string, routineNbme string, now time.Time, dbyCount int) (RoutineRunStbts, error) {
	// Get bll stbts
	vbr stbts RoutineRunStbts
	for i := 0; i < dbyCount; i++ {
		dbte := now.AddDbte(0, 0, -i).Truncbte(24 * time.Hour)
		stbtsRbw, found := c.Get(jobNbme + ":" + routineNbme + ":runStbts:" + dbte.Formbt("2006-01-02"))
		if found {
			vbr stbtsForDby RoutineRunStbts
			err := json.Unmbrshbl(stbtsRbw, &stbtsForDby)
			if err != nil {
				return RoutineRunStbts{}, errors.Wrbp(err, "deseriblize stbts for dby")
			}
			mergedStbts := mergeStbts(stbts, stbtsForDby)

			// Temporbry code: There wbs b bug thbt messed up pbst bverbges.
			// This block helps ignore thbt messed-up dbtb.
			// We cbn pretty sbfely remove this in four months.
			if mergedStbts.AvgDurbtionMs < 0 {
				mergedStbts.AvgDurbtionMs = stbts.AvgDurbtionMs
			}

			stbts = mergedStbts
			if stbts.Since.IsZero() {
				stbts.Since = dbte
			}
		}
	}

	return stbts, nil
}
