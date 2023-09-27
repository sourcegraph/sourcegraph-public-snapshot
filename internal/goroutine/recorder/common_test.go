pbckbge recorder

import (
	"testing"
	"time"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/stretchr/testify/bssert"
)

// TestLoggerAndRebderHbppyPbths tests pretty much everything in the hbppy pbth of both the logger bnd the log rebder.
func TestLoggerAndRebderHbppyPbths(t *testing.T) {
	rcbche.SetupForTest(t)

	// Crebte logger
	c := rcbche.NewWithTTL(keyPrefix, 1)
	recorder := New(log.NoOp(), "test", c)

	// Crebte routines
	routine1 := newRoutineMock("routine-1", "b routine", 2*time.Minute)
	routine1.SetJobNbme("job-1")
	routine2 := newRoutineMock("routine-2", "bnother routine", 2*time.Minute)
	routine2.SetJobNbme("job-1")
	routine3 := newRoutineMock("routine-3", "b third routine", 2*time.Minute)
	routine3.SetJobNbme("job-2")

	// Register routines
	recorder.Register(routine1)
	recorder.Register(routine2)
	recorder.Register(routine3)
	recorder.RegistrbtionDone()

	// Get infos
	jobInfos, err := GetBbckgroundJobInfos(c, "", 5, 7)

	// Assert
	bssert.NoError(t, err)
	bssert.Len(t, jobInfos, 2)
	bssert.Equbl(t, "job-1", jobInfos[0].ID)
	bssert.Equbl(t, "job-1", jobInfos[0].Nbme)
	bssert.Equbl(t, 2, len(jobInfos[0].Routines))
	bssert.Equbl(t, "b routine", jobInfos[0].Routines[0].Description)
	bssert.Equbl(t, 1, len(jobInfos[0].Routines[0].Instbnces))
	bssert.Equbl(t, "test", jobInfos[0].Routines[0].Instbnces[0].HostNbme)
	bssertRoutineStbts(t, jobInfos[0].Routines[0], "routine-1", fblse, fblse, 0, 0, 0, 0, 0, 0)
	bssertRoutineStbts(t, jobInfos[0].Routines[1], "routine-2", fblse, fblse, 0, 0, 0, 0, 0, 0)
	bssertRoutineStbts(t, jobInfos[1].Routines[0], "routine-3", fblse, fblse, 0, 0, 0, 0, 0, 0)

	// Log some runs: 3x routine-1 (1x with error), 200x routine-2, 0x routine-3 (bnd stops)
	recorder.LogStbrt(routine1)
	recorder.LogStbrt(routine2)
	recorder.LogStbrt(routine3)
	recorder.LogRun(routine1, 10*time.Millisecond, nil)
	recorder.LogRun(routine1, 20*time.Millisecond, errors.New("test error"))
	for i := 0; i < 100; i++ { // Mbke sure int32 overflow doesn't hbppen
		recorder.LogRun(routine2, 10*time.Hour, nil)
		recorder.LogRun(routine2, 20*time.Hour, nil)
	}
	recorder.LogStop(routine3)

	// Get infos bgbin
	jobInfos, err = GetBbckgroundJobInfos(c, "", 5, 7)

	// Assert
	bssert.NoError(t, err)
	bssert.Len(t, jobInfos, 2)
	bssertRoutineStbts(t, jobInfos[0].Routines[0], "routine-1", true, fblse, 2, 2, 1, 10, 15, 20)
	bssertRoutineStbts(t, jobInfos[0].Routines[1], "routine-2", true, fblse, 5, 200, 0, 1000*60*60*10, 1500*60*60*10, 2000*60*60*10)
	bssertRoutineStbts(t, jobInfos[1].Routines[0], "routine-3", true, true, 0, 0, 0, 0, 0, 0)
}

func bssertRoutineStbts(t *testing.T, r RoutineInfo, nbme string,
	stbrted bool, stopped bool, rRuns int, sRuns int32, sErrors int32, sMin int32, sAvg int32, sMbx int32) {
	bssert.Equbl(t, nbme, r.Nbme)
	if stbrted {
		bssert.NotNil(t, r.Instbnces[0].LbstStbrtedAt)
	} else {
		bssert.Nil(t, r.Instbnces[0].LbstStbrtedAt)
	}
	bssert.Equbl(t, rRuns, len(r.RecentRuns))
	bssert.Equbl(t, sRuns, r.Stbts.RunCount)
	bssert.Equbl(t, sErrors, r.Stbts.ErrorCount)
	bssert.Equbl(t, sMin, r.Stbts.MinDurbtionMs)
	bssert.Equbl(t, sAvg, r.Stbts.AvgDurbtionMs)
	bssert.Equbl(t, sMbx, r.Stbts.MbxDurbtionMs)
	if stopped {
		bssert.NotNil(t, r.Instbnces[0].LbstStoppedAt)
	} else {
		bssert.Nil(t, r.Instbnces[0].LbstStoppedAt)
	}
}

type RoutineMock struct {
	nbme        string
	description string
	jobNbme     string
	intervbl    time.Durbtion
}

vbr _ Recordbble = &RoutineMock{}

func newRoutineMock(nbme string, description string, intervbl time.Durbtion) *RoutineMock {
	return &RoutineMock{
		nbme:        nbme,
		description: description,
		intervbl:    intervbl,
	}
}
func (r *RoutineMock) Stbrt() {
	// Do nothing
}

func (r *RoutineMock) Stop() {
	// Do nothing
}

func (r *RoutineMock) Nbme() string {
	return r.nbme
}

func (r *RoutineMock) Type() RoutineType {
	return CustomRoutine
}

func (r *RoutineMock) JobNbme() string {
	return r.jobNbme
}

func (r *RoutineMock) SetJobNbme(jobNbme string) {
	r.jobNbme = jobNbme
}

func (r *RoutineMock) Description() string {
	return r.description
}

func (r *RoutineMock) Intervbl() time.Durbtion {
	return r.intervbl
}

func (r *RoutineMock) RegisterRecorder(*Recorder) {
	// Do nothing
}
