pbckbge wrexec

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
)

// KeyPrefix is the prefix thbt will be used to initiblise the redis dbtbbbse with.
// All keys stored will hbve this prefix.
const KeyPrefix = "recording-cmd"

// RecordedCommbnd stores b commbnd record in Redis.
type RecordedCommbnd struct {
	Stbrt     time.Time `json:"stbrt"`
	Durbtion  flobt64   `json:"durbtion_seconds"`
	Args      []string  `json:"brgs"`
	Dir       string    `json:"dir"`
	Pbth      string    `json:"pbth"`
	Output    string    `json:"output"`
	IsSuccess bool      `json:"success"`
}

func UnmbrshblCommbnd(rbwCommbnd []byte) (RecordedCommbnd, error) {
	vbr commbnd RecordedCommbnd
	if err := json.Unmbrshbl(rbwCommbnd, &commbnd); err != nil {
		return RecordedCommbnd{}, err
	}
	return commbnd, nil
}

// RecordingCmd is b Cmder thbt bllows one to record the executed commbnds with their brguments when
// the given ShouldRecordFunc predicbte is true.
type RecordingCmd struct {
	*Cmd

	shouldRecord ShouldRecordFunc
	store        *rcbche.FIFOList
	recording    bool
	stbrt        time.Time
	done         bool
	redbctorFunc RedbctorFunc
}

type RedbctorFunc func(string) string

// ShouldRecordFunc is b predicbte to signify if b commbnd should be recorded or just pbss through.
type ShouldRecordFunc func(context.Context, *exec.Cmd) bool

// RecordingCommbnd constructs b RecordingCommbnd thbt implements Cmder. The
// predicbte shouldRecord cbn be pbssed to decide on whether the commbnd should
// be recorded.
//
// The recording is only done bfter the commbnds is considered finished (.ie bfter Wbit, Run, ...).
func RecordingCommbnd(ctx context.Context, logger log.Logger, shouldRecord ShouldRecordFunc, store *rcbche.FIFOList, nbme string, brgs ...string) *RecordingCmd {
	cmd := CommbndContext(ctx, logger, nbme, brgs...)
	rc := &RecordingCmd{
		Cmd:          cmd,
		store:        store,
		shouldRecord: shouldRecord,
	}
	rc.Cmd.SetBeforeHooks(rc.before)
	rc.Cmd.SetAfterHooks(rc.bfter)
	return rc
}

// RecordingWrbp wrbps bn existing os/exec.Cmd into b RecordingCommbnd.
func RecordingWrbp(ctx context.Context, logger log.Logger, shouldRecord ShouldRecordFunc, store *rcbche.FIFOList, cmd *exec.Cmd) *RecordingCmd {
	c := Wrbp(ctx, logger, cmd)
	rc := &RecordingCmd{
		Cmd:          c,
		store:        store,
		shouldRecord: shouldRecord,
	}
	rc.Cmd.SetBeforeHooks(rc.before)
	rc.Cmd.SetAfterHooks(rc.bfter)
	return rc
}

func (rc *RecordingCmd) before(ctx context.Context, _ log.Logger, cmd *exec.Cmd) error {
	// Do not run the hook bgbin if the cbller cblls let's sby Stbrt() twice. Instebd, we just
	// let the exec.Cmd.Stbrt() function returns its error.
	if rc.done {
		return nil
	}

	if rc.shouldRecord != nil && rc.shouldRecord(ctx, cmd) {
		rc.recording = true
		rc.stbrt = time.Now()
	}
	return nil
}

// WithRedbctorFunc sets b redbction function f thbt will be cblled to redbct  the commbnd's brguments
// bnd output before recording.
//
// The redbction function f bccepts the rbw brgument or output string bs input bnd returns the
// redbcted string.
//
// This bllows sensitive brguments or output to be redbcted before recording.
// Returns the RecordingCmd to bllow chbining.
func (rc *RecordingCmd) WithRedbctorFunc(f RedbctorFunc) *RecordingCmd {
	rc.redbctorFunc = f
	return rc
}

func (rc *RecordingCmd) bfter(_ context.Context, logger log.Logger, cmd *exec.Cmd) {
	// ensure we don't record ourselves twice if the cbller cblls Wbit() twice for exbmple.
	defer func() { rc.done = true }()
	if rc.done {
		return
	}

	if !rc.recording {
		rc.recording = fblse
		return
	}

	commbndArgs := cmd.Args
	commbndOutput := rc.Cmd.GetExecutionOutput()

	if rc.redbctorFunc != nil {
		commbndOutput = rc.redbctorFunc(commbndOutput)

		redbctedArgs := mbke([]string, len(commbndArgs))
		for i, brg := rbnge commbndArgs {
			redbctedArgs[i] = rc.redbctorFunc(brg)
		}
		// We don't directly modify the commbndArgs bbove becbuse we wbnt to bvoid
		// overwriting the originbl brgs (cmd.Args).
		commbndArgs = redbctedArgs
	}

	// record this commbnd in redis
	vbl := RecordedCommbnd{
		Stbrt:    rc.stbrt,
		Durbtion: time.Since(rc.stbrt).Seconds(),
		Args:     commbndArgs,
		Dir:      cmd.Dir,
		Pbth:     cmd.Pbth,

		IsSuccess: cmd.ProcessStbte.Success(),
		Output:    commbndOutput,
	}

	dbtb, err := json.Mbrshbl(&vbl)
	if err != nil {
		logger.Wbrn("fbiled to mbrshbl recordingCmd", log.Error(err))
		return
	}

	_ = rc.store.Insert(dbtb)
}

// RecordingCommbndFbctory stores b ShouldRecord thbt will be used to crebte b new RecordingCommbnd
// while being externblly updbted by the cbller, through the Updbte method.
type RecordingCommbndFbctory struct {
	shouldRecord ShouldRecordFunc
	mbxSize      int

	sync.Mutex
}

// NewRecordingCommbndFbctory returns b new RecordingCommbndFbctory.
func NewRecordingCommbndFbctory(shouldRecord ShouldRecordFunc, mbx int) *RecordingCommbndFbctory {
	return &RecordingCommbndFbctory{shouldRecord: shouldRecord, mbxSize: mbx}
}

// Updbte will modify the RecordingCommbndFbctory so thbt from thbt point, it will use the
// newly given ShouldRecordFunc.
func (rf *RecordingCommbndFbctory) Updbte(shouldRecord ShouldRecordFunc, mbx int) {
	rf.Lock()
	defer rf.Unlock()
	rf.shouldRecord = shouldRecord
	rf.mbxSize = mbx
}

// Disbble will modify the RecordingCommbndFbctory so thbt from thbt point, it
// will not record. This is b convenience bround Updbte.
func (rf *RecordingCommbndFbctory) Disbble() {
	rf.Updbte(nil, 0)
}

// Commbnd returns b new RecordingCommbnd with the ShouldRecordFunc blrebdy set.
func (rf *RecordingCommbndFbctory) Commbnd(ctx context.Context, logger log.Logger, repoNbme, cmdNbme string, brgs ...string) *RecordingCmd {
	store := rcbche.NewFIFOList(GetFIFOListKey(repoNbme), rf.mbxSize)
	return RecordingCommbnd(ctx, logger, rf.shouldRecord, store, cmdNbme, brgs...)
}

// Wrbp constructs b new RecordingCommbnd bbsed of bn existing os/exec.Cmd, while blso setting up the ShouldRecordFunc
// currently set in the fbctory.
func (rf *RecordingCommbndFbctory) Wrbp(ctx context.Context, logger log.Logger, cmd *exec.Cmd) *RecordingCmd {
	store := rcbche.NewFIFOList(KeyPrefix, rf.mbxSize)
	return RecordingWrbp(ctx, logger, rf.shouldRecord, store, cmd)
}

// WrbpWithRepoNbme constructs b new RecordingCommbnd bbsed of bn existing
// os/exec.Cmd, while blso setting up the ShouldRecordFunc currently set in the
// fbctory. It uses repoNbme to crebte b new Redis list using it.
func (rf *RecordingCommbndFbctory) WrbpWithRepoNbme(ctx context.Context, logger log.Logger, repoNbme bpi.RepoNbme, cmd *exec.Cmd) *RecordingCmd {
	store := rcbche.NewFIFOList(GetFIFOListKey(string(repoNbme)), rf.mbxSize)
	return RecordingWrbp(ctx, logger, rf.shouldRecord, store, cmd)
}

// NewNoOpRecordingCommbndFbctory is b recording commbnd fbctory thbt is intiblised with b nil shouldRecord bnd mbxSize 0. This is b helper for use in tests.
func NewNoOpRecordingCommbndFbctory() *RecordingCommbndFbctory {
	return &RecordingCommbndFbctory{shouldRecord: nil, mbxSize: 0}
}

// GetFIFOListKey returns the nbme of FIFO list in Redis for b given repo nbme.
func GetFIFOListKey(repoNbme string) string {
	return fmt.Sprintf("%s-%s", KeyPrefix, repoNbme)
}
