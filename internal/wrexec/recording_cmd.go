package wrexec

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

// KeyPrefix is the prefix that will be used to initialise the redis database with.
// All keys stored will have this prefix.
const KeyPrefix = "recording-cmd"

// RecordedCommand stores a command record in Redis.
type RecordedCommand struct {
	Start     time.Time `json:"start"`
	Duration  float64   `json:"duration_seconds"`
	Args      []string  `json:"args"`
	Dir       string    `json:"dir"`
	Path      string    `json:"path"`
	Output    string    `json:"output"`
	IsSuccess bool      `json:"success"`
}

func UnmarshalCommand(rawCommand []byte) (RecordedCommand, error) {
	var command RecordedCommand
	if err := json.Unmarshal(rawCommand, &command); err != nil {
		return RecordedCommand{}, err
	}
	return command, nil
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
	redactorFunc RedactorFunc
}

type RedactorFunc func(string) string

// ShouldRecordFunc is a predicate to signify if a command should be recorded or just pass through.
type ShouldRecordFunc func(context.Context, *exec.Cmd) bool

// RecordingCommand constructs a RecordingCommand that implements Cmder. The
// predicate shouldRecord can be passed to decide on whether the command should
// be recorded.
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

func (rc *RecordingCmd) before(ctx context.Context, _ log.Logger, cmd *exec.Cmd) error {
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

// WithRedactorFunc sets a redaction function f that will be called to redact  the command's arguments
// and output before recording.
//
// The redaction function f accepts the raw argument or output string as input and returns the
// redacted string.
//
// This allows sensitive arguments or output to be redacted before recording.
// Returns the RecordingCmd to allow chaining.
func (rc *RecordingCmd) WithRedactorFunc(f RedactorFunc) *RecordingCmd {
	rc.redactorFunc = f
	return rc
}

func (rc *RecordingCmd) after(_ context.Context, logger log.Logger, cmd *exec.Cmd) {
	// ensure we don't record ourselves twice if the caller calls Wait() twice for example.
	defer func() { rc.done = true }()
	if rc.done {
		return
	}

	if !rc.recording {
		rc.recording = false
		return
	}

	commandArgs := cmd.Args
	commandOutput := rc.Cmd.GetExecutionOutput()

	if rc.redactorFunc != nil {
		commandOutput = rc.redactorFunc(commandOutput)

		redactedArgs := make([]string, len(commandArgs))
		for i, arg := range commandArgs {
			redactedArgs[i] = rc.redactorFunc(arg)
		}
		// We don't directly modify the commandArgs above because we want to avoid
		// overwriting the original args (cmd.Args).
		commandArgs = redactedArgs
	}

	isSuccess := false
	if cmd.ProcessState != nil {
		isSuccess = cmd.ProcessState.Success()
	}

	// record this command in redis
	val := RecordedCommand{
		Start:    rc.start,
		Duration: time.Since(rc.start).Seconds(),
		Args:     commandArgs,
		Dir:      cmd.Dir,
		Path:     cmd.Path,

		IsSuccess: isSuccess,
		Output:    commandOutput,
	}

	data, err := json.Marshal(&val)
	if err != nil {
		logger.Warn("failed to marshal recordingCmd", log.Error(err))
		return
	}

	_ = rc.store.Insert(data)
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

// Disable will modify the RecordingCommandFactory so that from that point, it
// will not record. This is a convenience around Update.
func (rf *RecordingCommandFactory) Disable() {
	rf.Update(nil, 0)
}

// Command returns a new RecordingCommand with the ShouldRecordFunc already set.
func (rf *RecordingCommandFactory) Command(ctx context.Context, logger log.Logger, repoName, cmdName string, args ...string) *RecordingCmd {
	store := rcache.NewFIFOList(redispool.Cache, GetFIFOListKey(repoName), rf.maxSize)
	return RecordingCommand(ctx, logger, rf.shouldRecord, store, cmdName, args...)
}

// Wrap constructs a new RecordingCommand based of an existing os/exec.Cmd, while also setting up the ShouldRecordFunc
// currently set in the factory.
func (rf *RecordingCommandFactory) Wrap(ctx context.Context, logger log.Logger, cmd *exec.Cmd) *RecordingCmd {
	store := rcache.NewFIFOList(redispool.Cache, KeyPrefix, rf.maxSize)
	return RecordingWrap(ctx, logger, rf.shouldRecord, store, cmd)
}

// WrapWithRepoName constructs a new RecordingCommand based of an existing
// os/exec.Cmd, while also setting up the ShouldRecordFunc currently set in the
// factory. It uses repoName to create a new Redis list using it.
func (rf *RecordingCommandFactory) WrapWithRepoName(ctx context.Context, logger log.Logger, repoName api.RepoName, cmd *exec.Cmd) *RecordingCmd {
	store := rcache.NewFIFOList(redispool.Cache, GetFIFOListKey(string(repoName)), rf.maxSize)
	return RecordingWrap(ctx, logger, rf.shouldRecord, store, cmd)
}

// NewNoOpRecordingCommandFactory is a recording command factory that is intialised with a nil shouldRecord and maxSize 0. This is a helper for use in tests.
func NewNoOpRecordingCommandFactory() *RecordingCommandFactory {
	return &RecordingCommandFactory{shouldRecord: nil, maxSize: 0}
}

// GetFIFOListKey returns the name of FIFO list in Redis for a given repo name.
func GetFIFOListKey(repoName string) string {
	return fmt.Sprintf("%s-%s", KeyPrefix, repoName)
}
