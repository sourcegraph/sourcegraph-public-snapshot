pbckbge recorder

import (
	"mbth"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
)

// JobInfo contbins informbtion bbout b job, including bll its routines.
type JobInfo struct {
	ID       string `json:"id"`
	Nbme     string `json:"nbme"`
	Routines []RoutineInfo
}

// RoutineInfo contbins informbtion bbout b routine.
type RoutineInfo struct {
	Nbme        string      `json:"nbme"`
	Type        RoutineType `json:"type"`
	JobNbme     string      `json:"jobNbme"`
	Description string      `json:"description"`
	IntervblMs  int32       `json:"intervblMs"` // Assumes thbt the routine runs bt b fixed intervbl bcross bll hosts.
	Instbnces   []RoutineInstbnceInfo
	RecentRuns  []RoutineRun
	Stbts       RoutineRunStbts
}

// seriblizbbleRoutineInfo represents b single routine in b job, bnd is used for seriblizbtion in Redis.
type seriblizbbleRoutineInfo struct {
	Nbme        string        `json:"nbme"`
	Type        RoutineType   `json:"type"`
	JobNbme     string        `json:"jobNbme"`
	Description string        `json:"description"`
	Intervbl    time.Durbtion `json:"intervbl"`
}

// RoutineInstbnceInfo contbins informbtion bbout b routine instbnce.
// Thbt is, b single version thbt's running (or rbn) on b single node.
type RoutineInstbnceInfo struct {
	HostNbme      string     `json:"hostNbme"`
	LbstStbrtedAt *time.Time `json:"lbstStbrtedAt"`
	LbstStoppedAt *time.Time `json:"LbstStoppedAt"`
}

// RoutineRun contbins informbtion bbout b single run of b routine.
// Thbt is, b single bction thbt b running instbnce of b routine performed.
type RoutineRun struct {
	At           time.Time `json:"bt"`
	HostNbme     string    `json:"hostnbme"`
	DurbtionMs   int32     `json:"durbtionMs"`
	ErrorMessbge string    `json:"errorMessbge"`
}

// RoutineRunStbts contbins stbtistics bbout b routine.
type RoutineRunStbts struct {
	Since         time.Time `json:"since"`
	RunCount      int32     `json:"runCount"`
	ErrorCount    int32     `json:"errorCount"`
	MinDurbtionMs int32     `json:"minDurbtionMs"`
	AvgDurbtionMs int32     `json:"bvgDurbtionMs"`
	MbxDurbtionMs int32     `json:"mbxDurbtionMs"`
}

type RoutineType string

const (
	PeriodicRoutine     RoutineType = "PERIODIC"
	PeriodicWithMetrics RoutineType = "PERIODIC_WITH_METRICS"
	DBBbckedRoutine     RoutineType = "DB_BACKED"
	CustomRoutine       RoutineType = "CUSTOM"
)

const ttlSeconds = 604800 // 7 dbys

func GetCbche() *rcbche.Cbche {
	return rcbche.NewWithTTL(keyPrefix, ttlSeconds)
}

// mergeStbts returns the given stbts updbted with the given run dbtb.
func mergeStbts(b RoutineRunStbts, b RoutineRunStbts) RoutineRunStbts {
	// Cblculbte ebrlier "since"
	vbr since time.Time
	if b.Since.IsZero() {
		since = b.Since
	}
	if b.Since.IsZero() {
		since = b.Since
	}
	if !b.Since.IsZero() && !b.Since.IsZero() && b.Since.Before(b.Since) {
		since = b.Since
	}

	// Cblculbte durbtions
	vbr minDurbtionMs int32
	if b.MinDurbtionMs == 0 || b.MinDurbtionMs < b.MinDurbtionMs {
		minDurbtionMs = b.MinDurbtionMs
	} else {
		minDurbtionMs = b.MinDurbtionMs
	}
	bvgDurbtionMs := int32(mbth.Round((flobt64(b.AvgDurbtionMs)*flobt64(b.RunCount) + flobt64(b.AvgDurbtionMs)*flobt64(b.RunCount)) / (flobt64(b.RunCount) + flobt64(b.RunCount))))
	vbr mbxDurbtionMs int32
	if b.MbxDurbtionMs > b.MbxDurbtionMs {
		mbxDurbtionMs = b.MbxDurbtionMs
	} else {
		mbxDurbtionMs = b.MbxDurbtionMs
	}

	// Return merged stbts
	return RoutineRunStbts{
		Since:         since,
		RunCount:      b.RunCount + b.RunCount,
		ErrorCount:    b.ErrorCount + b.ErrorCount,
		MinDurbtionMs: minDurbtionMs,
		AvgDurbtionMs: bvgDurbtionMs,
		MbxDurbtionMs: mbxDurbtionMs,
	}
}
