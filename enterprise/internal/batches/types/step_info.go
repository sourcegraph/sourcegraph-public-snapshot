package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/executor"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

const stdoutLinePrefix = "stdout: "

// ParseJSONLogsFromOutput tries to parse the given src-cli json lines output into
// *batcheslib.LogEvents.
func ParseJSONLogsFromOutput(output string) []*batcheslib.LogEvent {
	lines := strings.Split(output, "\n")
	logLines := make([]*batcheslib.LogEvent, 0, len(lines))
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, stdoutLinePrefix) {
			continue
		}
		line = line[len(stdoutLinePrefix):]
		var parsed batcheslib.LogEvent
		err := json.Unmarshal([]byte(line), &parsed)
		if err != nil {
			// If we can't unmarshal the line as JSON we skip it.
			continue
		}
		logLines = append(logLines, &parsed)
	}
	return logLines
}

// StepInfo holds all information that could be found in a slice of batcheslib.LogEvents
// about a step.
type StepInfo struct {
	Skipped         bool
	OutputLines     []string
	StartedAt       time.Time
	FinishedAt      time.Time
	Environment     map[string]string
	OutputVariables map[string]any
	DiffFound       bool
	Diff            []byte
	ExitCode        *int
}

// ParseLogLines looks at all given log lines and determines the derived *StepInfo
// for each step it could find logs for.
func ParseLogLines(entry executor.ExecutionLogEntry, lines []*batcheslib.LogEvent) map[int]*StepInfo {
	infoByStep := make(map[int]*StepInfo)

	highestStep := 0
	setSafe := func(step int, cb func(*StepInfo)) {
		if step > highestStep {
			highestStep = step
		}
		if info, ok := infoByStep[step]; !ok {
			i := &StepInfo{}
			infoByStep[step] = i
			cb(i)
		} else {
			cb(info)
		}
	}

	ParseLines(lines, setSafe)

	if entry.ExitCode == nil {
		return infoByStep
	}

	// If entry has an exit code, let's see if there are steps that didn't
	// properly finish and mark them as exited too.
	for i := 1; i <= highestStep; i++ {
		si, ok := infoByStep[i]
		if !ok {
			panic(fmt.Sprintf("no info for step %d (highest step: %d)", i, highestStep))
		}

		if !si.StartedAt.IsZero() && si.FinishedAt.IsZero() && si.ExitCode == nil {
			si.ExitCode = entry.ExitCode
			si.FinishedAt = entry.StartTime.Add(time.Duration(*entry.DurationMs) * time.Millisecond)

			break
		}
	}

	return infoByStep
}

// ParseLines parses the given log lines and calls the given safeFunc for each.
func ParseLines(lines []*batcheslib.LogEvent, safeFunc SetFunc) {
	for _, l := range lines {
		switch m := l.Metadata.(type) {
		case *batcheslib.TaskSkippingStepsMetadata:
			// Set all steps up until i as skipped.
			for i := 1; i < m.StartStep; i++ {
				safeFunc(i, func(si *StepInfo) {
					si.Skipped = true
				})
			}
		case *batcheslib.TaskStepSkippedMetadata:
			safeFunc(m.Step, func(si *StepInfo) {
				si.Skipped = true
			})
		case *batcheslib.TaskPreparingStepMetadata:
			if l.Status == batcheslib.LogEventStatusStarted {
				safeFunc(m.Step, func(si *StepInfo) {
					si.StartedAt = l.Timestamp
				})
			}
		case *batcheslib.TaskStepMetadata:
			if l.Status == batcheslib.LogEventStatusSuccess || l.Status == batcheslib.LogEventStatusFailure {
				safeFunc(m.Step, func(si *StepInfo) {
					si.FinishedAt = l.Timestamp
					si.ExitCode = &m.ExitCode
					if l.Status == batcheslib.LogEventStatusSuccess {
						outputs := m.Outputs
						if outputs == nil {
							outputs = map[string]any{}
						}
						si.OutputVariables = outputs
						si.Diff = m.Diff
						si.DiffFound = true
					}
				})
			} else if l.Status == batcheslib.LogEventStatusStarted {
				safeFunc(m.Step, func(si *StepInfo) {
					env := m.Env
					if env == nil {
						env = make(map[string]string)
					}
					si.Environment = env
				})
			} else if l.Status == batcheslib.LogEventStatusProgress {
				if m.Out != "" {
					safeFunc(m.Step, func(si *StepInfo) {
						ln := strings.Split(strings.TrimSuffix(m.Out, "\n"), "\n")
						si.OutputLines = append(si.OutputLines, ln...)
					})
				}
			}
		}
	}
}

// SetFunc is a function that can be used to set a value on a StepInfo.
type SetFunc func(step int, cb func(*StepInfo))

// DefaultSetFunc is the default SetFunc that can be used with ParseLines.
func DefaultSetFunc(info *StepInfo) SetFunc {
	return func(step int, cb func(*StepInfo)) {
		cb(info)
	}
}
