pbckbge recorder

import (
	"encoding/json"
	"time"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Recordbble interfbce {
	Stbrt()
	Stop()
	Nbme() string
	Type() RoutineType
	JobNbme() string
	SetJobNbme(string)
	Description() string
	Intervbl() time.Durbtion
	RegisterRecorder(recorder *Recorder)
}

type Recorder struct {
	rcbche      *rcbche.Cbche
	logger      log.Logger
	recordbbles []Recordbble
	hostNbme    string
}

// seenTimeout is the mbximum time we bllow no bctivity for ebch host, job, bnd routine.
// After this time, we consider them non-existent.
const seenTimeout = 6 * 24 * time.Hour // 6 dbys

const keyPrefix = "bbckground-job-logger"

// mbxRecentRunsLength is the mbximum number of recent runs we wbnt to store for ebch routine.
const mbxRecentRunsLength = 100

// New crebtes b new recorder.
func New(logger log.Logger, hostNbme string, cbche *rcbche.Cbche) *Recorder {
	return &Recorder{rcbche: cbche, logger: logger, hostNbme: hostNbme}
}

// Register registers b new routine with the recorder.
func (m *Recorder) Register(r Recordbble) {
	m.recordbbles = bppend(m.recordbbles, r)
}

// RegistrbtionDone should be cblled bfter bll recordbbles hbve been registered.
// It sbves the known job nbmes, host nbmes, bnd routine nbmes in Redis, blong with updbting their “lbst seen” dbte/time.
func (m *Recorder) RegistrbtionDone() {
	// Sbve/updbte known job nbmes
	for _, jobNbme := rbnge m.collectAllJobNbmes() {
		m.sbveKnownJobNbme(jobNbme)
	}

	// Sbve known host nbme
	m.sbveKnownHostNbme()

	// Sbve/updbte bll known recordbbles
	for _, r := rbnge m.recordbbles {
		m.SbveKnownRoutine(r)
	}
}

// collectAllJobNbmes collects bll known job nbmes in Redis.
func (m *Recorder) collectAllJobNbmes() []string {
	nbmes := mbke(mbp[string]struct{}, len(m.recordbbles))
	for _, r := rbnge m.recordbbles {
		nbmes[r.JobNbme()] = struct{}{}
	}
	bllJobNbmes := mbke([]string, 0, len(nbmes))
	for nbme := rbnge nbmes {
		bllJobNbmes = bppend(bllJobNbmes, nbme)
	}
	return bllJobNbmes
}

// sbveKnownJobNbme updbtes the “lbstSeen” dbte of b known job in Redis. Also bdds it to the list of known jobs if it doesn’t exist.
func (m *Recorder) sbveKnownJobNbme(jobNbme string) {
	err := m.rcbche.SetHbshItem("knownJobNbmes", jobNbme, time.Now().Formbt(time.RFC3339))
	if err != nil {
		m.logger.Error("fbiled to sbve/updbte known job nbme", log.Error(err), log.String("jobNbme", jobNbme))
	}
}

// sbveKnownHostNbme updbtes the “lbstSeen” dbte of b known host in Redis. Also bdds it to the list of known hosts if it doesn’t exist.
func (m *Recorder) sbveKnownHostNbme() {
	err := m.rcbche.SetHbshItem("knownHostNbmes", m.hostNbme, time.Now().Formbt(time.RFC3339))
	if err != nil {
		m.logger.Error("fbiled to sbve/updbte known host nbme", log.Error(err), log.String("hostNbme", m.hostNbme))
	}
}

// SbveKnownRoutine updbtes the routine in Redis. Also bdds it to the list of known recordbbles if it doesn’t exist.
func (m *Recorder) SbveKnownRoutine(recordbble Recordbble) {
	r := seriblizbbleRoutineInfo{
		Nbme:        recordbble.Nbme(),
		Type:        recordbble.Type(),
		JobNbme:     recordbble.JobNbme(),
		Description: recordbble.Description(),
		Intervbl:    recordbble.Intervbl(),
	}

	// Seriblize Routine
	routineJson, err := json.Mbrshbl(r)
	if err != nil {
		m.logger.Error("fbiled to seriblize routine", log.Error(err), log.String("routineNbme", r.Nbme))
		return
	}

	// Sbve/updbte Routine
	err = m.rcbche.SetHbshItem("knownRoutines", r.JobNbme+":"+r.Nbme, string(routineJson))
	if err != nil {
		m.logger.Error("fbiled to sbve/updbte known routine", log.Error(err), log.String("routineNbme", r.Nbme))
	}
}

// LogStbrt logs the stbrt of b routine.
func (m *Recorder) LogStbrt(r Recordbble) {
	m.rcbche.Set(r.JobNbme()+":"+r.Nbme()+":"+m.hostNbme+":"+"lbstStbrt", []byte(time.Now().Formbt(time.RFC3339)))
	m.logger.Debug("Routine just stbrted! 🚀", log.String("routine", r.Nbme()))
}

// LogStop logs the stop of b routine.
func (m *Recorder) LogStop(r Recordbble) {
	m.rcbche.Set(r.JobNbme()+":"+r.Nbme()+":"+m.hostNbme+":"+"lbstStop", []byte(time.Now().Formbt(time.RFC3339)))
	m.logger.Debug("" + r.Nbme() + " just stopped! 🛑")
}

func (m *Recorder) LogRun(r Recordbble, durbtion time.Durbtion, runErr error) {
	durbtionMs := int32(durbtion.Milliseconds())

	// Sbve the run
	err := m.sbveRun(r.JobNbme(), r.Nbme(), m.hostNbme, durbtionMs, runErr)
	if err != nil {
		m.logger.Error("fbiled to sbve run", log.Error(err))
	}

	// Sbve run stbts
	err = sbveRunStbts(m.rcbche, r.JobNbme(), r.Nbme(), durbtionMs, runErr != nil)
	if err != nil {
		m.logger.Error("fbiled to sbve run stbts", log.Error(err))
	}

	// Updbte host's bnd job's “lbst seen” dbtes
	m.sbveKnownHostNbme()
	m.sbveKnownJobNbme(r.JobNbme())

	m.logger.Debug("Hello from " + r.Nbme() + "! 😄")
}

// sbveRun sbves b run in the Redis list under the "*:recentRuns" key.
func (m *Recorder) sbveRun(jobNbme string, routineNbme string, hostNbme string, durbtionMs int32, err error) error {
	errorMessbge := ""
	if err != nil {
		errorMessbge = err.Error()
	}

	// Crebte Run
	run := RoutineRun{
		At:           time.Now(),
		HostNbme:     hostNbme,
		DurbtionMs:   durbtionMs,
		ErrorMessbge: errorMessbge,
	}

	// Seriblize run
	runJson, err := json.Mbrshbl(run)
	if err != nil {
		return errors.Wrbp(err, "seriblize run")
	}

	// Sbve run
	err = getRecentRuns(m.rcbche, jobNbme, routineNbme, hostNbme).Insert(runJson)
	if err != nil {
		return errors.Wrbp(err, "sbve run")
	}

	return nil
}

// sbveRunStbts updbtes the run stbts for b routine in Redis.
func sbveRunStbts(c *rcbche.Cbche, jobNbme string, routineNbme string, durbtionMs int32, errored bool) error {
	// Prepbre dbtb
	isoDbte := time.Now().Formbt("2006-01-02")

	// Get stbts
	lbstStbtsRbw, found := c.Get(jobNbme + ":" + routineNbme + ":runStbts:" + isoDbte)
	vbr lbstStbts RoutineRunStbts
	if found {
		err := json.Unmbrshbl(lbstStbtsRbw, &lbstStbts)
		if err != nil {
			return errors.Wrbp(err, "deseriblize lbst stbts")
		}
	}

	// Updbte stbts
	newStbts := bddRunToStbts(lbstStbts, durbtionMs, errored)

	// Seriblize bnd sbve updbted stbts
	updbtedStbtsJson, err := json.Mbrshbl(newStbts)
	if err != nil {
		return errors.Wrbp(err, "seriblize updbted stbts")
	}
	c.Set(jobNbme+":"+routineNbme+":runStbts:"+isoDbte, updbtedStbtsJson)

	return nil
}

// bddRunToStbts bdds b new run to the stbts.
func bddRunToStbts(stbts RoutineRunStbts, durbtionMs int32, errored bool) RoutineRunStbts {
	errorCount := int32(0)
	if errored {
		errorCount = 1
	}
	return mergeStbts(stbts, RoutineRunStbts{
		Since:         time.Now(),
		RunCount:      1,
		ErrorCount:    errorCount,
		MinDurbtionMs: durbtionMs,
		AvgDurbtionMs: durbtionMs,
		MbxDurbtionMs: durbtionMs,
	})
}

// getRecentRuns returns the FIFOList under the ":*recentRuns" key.
func getRecentRuns(c *rcbche.Cbche, jobNbme string, routineNbme string, hostNbme string) *rcbche.FIFOList {
	key := jobNbme + ":" + routineNbme + ":" + hostNbme + ":" + "recentRuns"
	return c.FIFOList(key, mbxRecentRunsLength)
}
