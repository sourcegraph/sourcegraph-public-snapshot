package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

const TTL = time.Hour * 24 * 7

type RecordingCmd struct {
	*Cmd

	shouldRecord ShouldRecordFunc
	r            *rcache.Cache
	recording    bool
	start        time.Time
	done         bool
}

type ShouldRecordFunc func(context.Context, *exec.Cmd) bool

func RecordingCommand(ctx context.Context, logger log.Logger, shouldRecord ShouldRecordFunc, name string, args ...string) *RecordingCmd {
	cmd := Command(ctx, logger, name, args...)
	rc := &RecordingCmd{
		Cmd:          cmd,
		r:            rcache.New("recording-cmd"),
		shouldRecord: shouldRecord,
	}
	rc.Cmd.SetBeforeHooks(rc.before)
	rc.Cmd.SetAfterHooks(rc.after)
	return rc
}

func RecordingWrap(ctx context.Context, logger log.Logger, shouldRecord ShouldRecordFunc, cmd *exec.Cmd) *RecordingCmd {
	c := Wrap(ctx, logger, cmd)
	rc := &RecordingCmd{
		Cmd:          c,
		r:            rcache.New("recording-cmd"),
		shouldRecord: shouldRecord,
	}
	rc.Cmd.SetBeforeHooks(rc.before)
	rc.Cmd.SetAfterHooks(rc.after)
	return rc
}

func (rc *RecordingCmd) before(ctx context.Context, logger log.Logger, cmd *exec.Cmd) error {
	println("👏")
	// Do not run the hook again if the caller calls let's say Start() twice. Instead, we just
	// let the exec.Cmd.Start() function returns its error.
	if rc.done {
		return nil
	}

	println("👏", rc.shouldRecord == nil)
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
	val := struct {
		Start    time.Time `json:"start"`
		Duration float64   `json:"duration_seconds"`
		Args     []string  `json:"args"`
		Dir      string    `json:"dir"`
		Path     string    `json:"path"`
	}{
		Start:    rc.start,
		Duration: time.Now().Sub(rc.start).Seconds(),
		Args:     cmd.Args,
		Dir:      cmd.Dir,
		Path:     cmd.Path,
	}

	data, err := json.Marshal(&val)
	if err != nil {
		logger.Warn("failed to marshal recordingCmd", log.Error(err))
	}

	// Using %p here, because timestamp + cmd address makes it unique. Your own command can't be
	// ran multiple time.
	rc.r.SetWithTTL(fmt.Sprintf("%v:%p", time.Now().Unix(), cmd), data, int(TTL.Seconds())) // TODO
}

type RecordingCommandFactory struct {
	shouldRecord ShouldRecordFunc
	sync.Mutex
}

func NewRecordingCommandFactory(shouldRecord ShouldRecordFunc) *RecordingCommandFactory {
	return &RecordingCommandFactory{shouldRecord: shouldRecord}
}

func (rf *RecordingCommandFactory) Update(shouldRecord ShouldRecordFunc) {
	rf.Lock()
	defer rf.Unlock()
	rf.shouldRecord = shouldRecord
}

func (rf *RecordingCommandFactory) Command(ctx context.Context, logger log.Logger, name string, args ...string) *RecordingCmd {
	return RecordingCommand(ctx, logger, rf.shouldRecord, name, args...)
}

func (rf *RecordingCommandFactory) Wrap(ctx context.Context, logger log.Logger, cmd *exec.Cmd) *RecordingCmd {
	return RecordingWrap(ctx, logger, rf.shouldRecord, cmd)
}
