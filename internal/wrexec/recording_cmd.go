package wrexec

import (
	"context"
	"encoding/json"
	"os/exec"
	"sync"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

// KeyPrefix is the prefix that will be used to initialise the redis database with.
// All keys stored will have this prefix.
const KeyPrefix = "recording-cmd"

// RecordedCommand stores a command record in Redis.
type RecordedCommand struct {
	Start    time.Time `json:"start"`
	Duration float64   `json:"duration_seconds"`
	Args     []string  `json:"args"`
	Dir      string    `json:"dir"`
	Path     string    `json:"path"`
}

// RecordingCmd is a Cmder that allows one to record the executed commands with their arguments when
// the given ShouldRecordFunc predicate is true.
type RecordingCmd struct {
	*Cmd

	shouldRecord ShouldRecordFunc
	store        *rcache.FIFOList
	recording    bool
	start        time.Time
	done         bool
}

// ShouldRecordFunc is a predicate to signifiy if a command should be recorded or just pass through.
type ShouldRecordFunc func(context.Context, *exec.Cmd) bool

// RecordingCommand contructs a RecordingCommand that implements Cmder. The predicate shouldRecord can be passed to decide on whether
// or not the command should be recorded.
//
// The recording is only done after the commands is considered finished (.ie after Wait, Run, ...).
func RecordingCommand(ctx context.Context, logger log.Logger, shouldRecord ShouldRecordFunc, store *rcache.FIFOList, name string, args ...string) *RecordingCmd {
	cmd := CommandContext(ctx, logger, name, args...)
	rc := &RecordingCmd{
		Cmd:          cmd,
		store:        store,
		shouldRecord: shouldRecord,
	}
	rc.Cmd.SetBeforeHooks(rc.before)
	rc.Cmd.SetAfterHooks(rc.after)
	return rc
}

// RecordingWrap wraps an existing os/exec.Cmd into a RecordingCommand.
func RecordingWrap(ctx context.Context, logger log.Logger, shouldRecord ShouldRecordFunc, store *rcache.FIFOList, cmd *exec.Cmd) *RecordingCmd {
	c := Wrap(ctx, logger, cmd)
	rc := &RecordingCmd{
		Cmd:          c,
		store:        store,
		shouldRecord: shouldRecord,
	}
	rc.Cmd.SetBeforeHooks(rc.before)
	rc.Cmd.SetAfterHooks(rc.after)
	return rc
}

func (rc *RecordingCmd) before(ctx context.Context, logger log.Logger, cmd *exec.Cmd) error {
	// Do not run the hook again if the caller calls let's say Start() twice. Instead, we just
	// let the exec.Cmd.Start() function returns its error.
	if rc.done {
		return nil
	}

	if rc.shouldRecord != nil && rc.shouldRecord(ctx, cmd) {
		rc.recording = true
		rc.start = time.Now()
	}
	return nil
}

func (rc *RecordingCmd) after(ctx context.Context, logger log.Logger, cmd *exec.Cmd) {
	// ensure we don't record ourselves twice if the caller calls Wait() twice for example.
	defer func() { rc.done = true }()
	if rc.done {
		return
	}

	if !rc.recording {
		rc.recording = false
		return
	}

	// record this command in redis
	val := RecordedCommand{
		Start:    rc.start,
		Duration: time.Since(rc.start).Seconds(),
		Args:     cmd.Args,
		Dir:      cmd.Dir,
		Path:     cmd.Path,
	}

	data, err := json.Marshal(&val)
	if err != nil {
		logger.Warn("failed to marshal recordingCmd", log.Error(err))
		return
	}

	rc.store.Insert(data)
}

// RecordingCommandFactory stores a ShouldRecord that will be used to create a new RecordingCommand
// while being externally updated by the caller, through the Update method.
type RecordingCommandFactory struct {
	shouldRecord ShouldRecordFunc
	maxSize      int

	sync.Mutex
}

// NewRecordingCommandFactory returns a new RecordingCommandFactory.
func NewRecordingCommandFactory(shouldRecord ShouldRecordFunc, max int) *RecordingCommandFactory {
	return &RecordingCommandFactory{shouldRecord: shouldRecord, maxSize: max}
}

// Update will modify the RecordingCommandFactory so that from that point, it will use the
// newly given ShouldRecordFunc.
func (rf *RecordingCommandFactory) Update(shouldRecord ShouldRecordFunc, max int) {
	rf.Lock()
	defer rf.Unlock()
	rf.shouldRecord = shouldRecord
	rf.maxSize = max
}

// Command returns a new RecordingCommand with the ShouldRecordFunc already set.
func (rf *RecordingCommandFactory) Command(ctx context.Context, logger log.Logger, name string, args ...string) *RecordingCmd {
	store := rcache.NewFIFOList(KeyPrefix, rf.maxSize)
	return RecordingCommand(ctx, logger, rf.shouldRecord, store, name, args...)
}

// Wrap constructs a new RecordingCommand based of an existing os/exec.Cmd, while also setting up the ShouldRecordFunc
// currently set in the factory.
func (rf *RecordingCommandFactory) Wrap(ctx context.Context, logger log.Logger, cmd *exec.Cmd) *RecordingCmd {
	store := rcache.NewFIFOList(KeyPrefix, rf.maxSize)
	return RecordingWrap(ctx, logger, rf.shouldRecord, store, cmd)
}
