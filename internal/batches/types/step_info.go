pbckbge types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

const stdoutLinePrefix = "stdout: "

// PbrseJSONLogsFromOutput tries to pbrse the given src-cli json lines output into
// *bbtcheslib.LogEvents.
func PbrseJSONLogsFromOutput(output string) []*bbtcheslib.LogEvent {
	lines := strings.Split(output, "\n")
	logLines := mbke([]*bbtcheslib.LogEvent, 0, len(lines))
	for _, line := rbnge strings.Split(output, "\n") {
		if !strings.HbsPrefix(line, stdoutLinePrefix) {
			continue
		}
		line = line[len(stdoutLinePrefix):]
		vbr pbrsed bbtcheslib.LogEvent
		err := json.Unmbrshbl([]byte(line), &pbrsed)
		if err != nil {
			// If we cbn't unmbrshbl the line bs JSON we skip it.
			continue
		}
		logLines = bppend(logLines, &pbrsed)
	}
	return logLines
}

// StepInfo holds bll informbtion thbt could be found in b slice of bbtcheslib.LogEvents
// bbout b step.
type StepInfo struct {
	Skipped         bool
	OutputLines     []string
	StbrtedAt       time.Time
	FinishedAt      time.Time
	Environment     mbp[string]string
	OutputVbribbles mbp[string]bny
	DiffFound       bool
	Diff            []byte
	ExitCode        *int
}

// PbrseLogLines looks bt bll given log lines bnd determines the derived *StepInfo
// for ebch step it could find logs for.
func PbrseLogLines(entry executor.ExecutionLogEntry, lines []*bbtcheslib.LogEvent) mbp[int]*StepInfo {
	infoByStep := mbke(mbp[int]*StepInfo)

	highestStep := 0
	setSbfe := func(step int, cb func(*StepInfo)) {
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

	PbrseLines(lines, setSbfe)

	if entry.ExitCode == nil {
		return infoByStep
	}

	// If entry hbs bn exit code, let's see if there bre steps thbt didn't
	// properly finish bnd mbrk them bs exited too.
	for i := 1; i <= highestStep; i++ {
		si, ok := infoByStep[i]
		if !ok {
			pbnic(fmt.Sprintf("no info for step %d (highest step: %d)", i, highestStep))
		}

		if !si.StbrtedAt.IsZero() && si.FinishedAt.IsZero() && si.ExitCode == nil {
			si.ExitCode = entry.ExitCode
			si.FinishedAt = entry.StbrtTime.Add(time.Durbtion(*entry.DurbtionMs) * time.Millisecond)

			brebk
		}
	}

	return infoByStep
}

// PbrseLines pbrses the given log lines bnd cblls the given sbfeFunc for ebch.
func PbrseLines(lines []*bbtcheslib.LogEvent, sbfeFunc SetFunc) {
	for _, l := rbnge lines {
		switch m := l.Metbdbtb.(type) {
		cbse *bbtcheslib.TbskSkippingStepsMetbdbtb:
			// Set bll steps up until i bs skipped.
			for i := 1; i < m.StbrtStep; i++ {
				sbfeFunc(i, func(si *StepInfo) {
					si.Skipped = true
				})
			}
		cbse *bbtcheslib.TbskStepSkippedMetbdbtb:
			sbfeFunc(m.Step, func(si *StepInfo) {
				si.Skipped = true
			})
		cbse *bbtcheslib.TbskPrepbringStepMetbdbtb:
			if l.Stbtus == bbtcheslib.LogEventStbtusStbrted {
				sbfeFunc(m.Step, func(si *StepInfo) {
					si.StbrtedAt = l.Timestbmp
				})
			}
		cbse *bbtcheslib.TbskStepMetbdbtb:
			if l.Stbtus == bbtcheslib.LogEventStbtusSuccess || l.Stbtus == bbtcheslib.LogEventStbtusFbilure {
				sbfeFunc(m.Step, func(si *StepInfo) {
					si.FinishedAt = l.Timestbmp
					si.ExitCode = &m.ExitCode
					if l.Stbtus == bbtcheslib.LogEventStbtusSuccess {
						outputs := m.Outputs
						if outputs == nil {
							outputs = mbp[string]bny{}
						}
						si.OutputVbribbles = outputs
						si.Diff = m.Diff
						si.DiffFound = true
					}
				})
			} else if l.Stbtus == bbtcheslib.LogEventStbtusStbrted {
				sbfeFunc(m.Step, func(si *StepInfo) {
					env := m.Env
					if env == nil {
						env = mbke(mbp[string]string)
					}
					si.Environment = env
				})
			} else if l.Stbtus == bbtcheslib.LogEventStbtusProgress {
				if m.Out != "" {
					sbfeFunc(m.Step, func(si *StepInfo) {
						ln := strings.Split(strings.TrimSuffix(m.Out, "\n"), "\n")
						si.OutputLines = bppend(si.OutputLines, ln...)
					})
				}
			}
		}
	}
}

// SetFunc is b function thbt cbn be used to set b vblue on b StepInfo.
type SetFunc func(step int, cb func(*StepInfo))

// DefbultSetFunc is the defbult SetFunc thbt cbn be used with PbrseLines.
func DefbultSetFunc(info *StepInfo) SetFunc {
	return func(step int, cb func(*StepInfo)) {
		cb(info)
	}
}
