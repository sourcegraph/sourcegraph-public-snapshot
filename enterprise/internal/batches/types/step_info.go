package types

import (
	"encoding/json"
	"strings"
	"time"

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
	OutputVariables map[string]interface{}
	Diff            *string
	ExitCode        *int
}

// ParseLogLines looks at all given log lines and determines the derived *StepInfo
// for each step it could find logs for.
func ParseLogLines(lines []*batcheslib.LogEvent) map[int]*StepInfo {
	infoByStep := make(map[int]*StepInfo)

	setSafe := func(step int, cb func(*StepInfo)) {
		if info, ok := infoByStep[step]; !ok {
			info := &StepInfo{}
			infoByStep[step] = info
			cb(info)
		} else {
			cb(info)
		}
	}

	for _, l := range lines {
		switch m := l.Metadata.(type) {
		case *batcheslib.TaskSkippingStepsMetadata:
			// Set all steps up until i as skipped.
			for i := 1; i < m.StartStep; i++ {
				setSafe(i, func(si *StepInfo) {
					si.Skipped = true
				})
			}
		case *batcheslib.TaskStepSkippedMetadata:
			setSafe(m.Step, func(si *StepInfo) {
				si.Skipped = true
			})
		case *batcheslib.TaskPreparingStepMetadata:
			if l.Status == batcheslib.LogEventStatusStarted {
				setSafe(m.Step, func(si *StepInfo) {
					si.StartedAt = l.Timestamp
				})
			}
		case *batcheslib.TaskStepMetadata:
			if l.Status == batcheslib.LogEventStatusSuccess || l.Status == batcheslib.LogEventStatusFailure {
				setSafe(m.Step, func(si *StepInfo) {
					si.FinishedAt = l.Timestamp
					si.ExitCode = &m.ExitCode
					if l.Status == batcheslib.LogEventStatusSuccess {
						outputs := m.Outputs
						if outputs == nil {
							outputs = map[string]interface{}{}
						}
						si.OutputVariables = outputs
						si.Diff = &m.Diff
					}
				})
			} else if l.Status == batcheslib.LogEventStatusStarted {
				setSafe(m.Step, func(si *StepInfo) {
					env := m.Env
					if env == nil {
						env = make(map[string]string)
					}
					si.Environment = env
				})
			} else if l.Status == batcheslib.LogEventStatusProgress {
				if m.Out != "" {
					setSafe(m.Step, func(si *StepInfo) {
						lines := strings.Split(strings.TrimSuffix(m.Out, "\n"), "\n")
						si.OutputLines = append(si.OutputLines, lines...)
					})
				}
			}
		}
	}

	return infoByStep
}
