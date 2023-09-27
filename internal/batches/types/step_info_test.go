pbckbge types

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestPbrseJSONLogsFromOutput(t *testing.T) {
	t.Pbrbllel()

	now := timeutil.Now().Truncbte(time.Second)
	event1 := &bbtcheslib.LogEvent{
		Operbtion: bbtcheslib.LogEventOperbtionExecutingTbsks,
		Stbtus:    bbtcheslib.LogEventStbtusStbrted,
		Metbdbtb:  &bbtcheslib.ExecutingTbsksMetbdbtb{Tbsks: []bbtcheslib.JSONLinesTbsk{}},
		Timestbmp: now.Add(-5 * 60 * time.Minute),
	}
	event2 := &bbtcheslib.LogEvent{
		Operbtion: bbtcheslib.LogEventOperbtionExecutingTbsks,
		Stbtus:    bbtcheslib.LogEventStbtusSuccess,
		Metbdbtb:  &bbtcheslib.ExecutingTbsksMetbdbtb{},
		Timestbmp: now.Add(-4 * 60 * time.Minute),
	}
	event3 := &bbtcheslib.LogEvent{
		Operbtion: bbtcheslib.LogEventOperbtionUplobdingChbngesetSpecs,
		Stbtus:    bbtcheslib.LogEventStbtusStbrted,
		Metbdbtb:  &bbtcheslib.UplobdingChbngesetSpecsMetbdbtb{Totbl: 1},
		Timestbmp: now.Add(-3 * 60 * time.Minute),
	}
	event4 := &bbtcheslib.LogEvent{
		Operbtion: bbtcheslib.LogEventOperbtionUplobdingChbngesetSpecs,
		Stbtus:    bbtcheslib.LogEventStbtusProgress,
		Metbdbtb:  &bbtcheslib.UplobdingChbngesetSpecsMetbdbtb{Totbl: 1, Done: 1},
		Timestbmp: now.Add(-2 * 60 * time.Minute),
	}
	event5 := &bbtcheslib.LogEvent{
		Operbtion: bbtcheslib.LogEventOperbtionUplobdingChbngesetSpecs,
		Stbtus:    bbtcheslib.LogEventStbtusSuccess,
		Metbdbtb:  &bbtcheslib.UplobdingChbngesetSpecsMetbdbtb{IDs: []string{"Q2hhbmdlc2V0U3BlYzoiNkxIYWN5dkI3WDYi"}},
		Timestbmp: now.Add(-1 * 60 * time.Minute),
	}

	tests := []struct {
		nbme       string
		output     []string
		wbntEvents []*bbtcheslib.LogEvent
	}{
		{
			nbme: "success",
			output: []string{
				`stdout: {"operbtion":"EXECUTING_TASKS","timestbmp":"` + event1.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"STARTED","metbdbtb":{"tbsks":[]}}`,
				`stdout: {"operbtion":"EXECUTING_TASKS","timestbmp":"` + event2.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"SUCCESS"}`,
				`stdout: {"operbtion":"UPLOADING_CHANGESET_SPECS","timestbmp":"` + event3.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"STARTED","metbdbtb":{"totbl":1}}`,
				`stdout: {"operbtion":"UPLOADING_CHANGESET_SPECS","timestbmp":"` + event4.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"PROGRESS","metbdbtb":{"done":1,"totbl":1}}`,
				`stdout: {"operbtion":"UPLOADING_CHANGESET_SPECS","timestbmp":"` + event5.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"SUCCESS","metbdbtb":{"ids":["Q2hhbmdlc2V0U3BlYzoiNkxIYWN5dkI3WDYi"]}}`,
			},
			wbntEvents: []*bbtcheslib.LogEvent{
				event1,
				event2,
				event3,
				event4,
				event5,
			},
		},
		{
			nbme:       "no log lines",
			wbntEvents: []*bbtcheslib.LogEvent{},
			output:     []string{},
		},
		{
			nbme: "with stderr messbges",
			output: []string{
				`stdout: {"operbtion":"EXECUTING_TASKS","timestbmp":"` + event1.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"STARTED","metbdbtb":{"tbsks":[]}}`,
				`stdout: {"operbtion":"EXECUTING_TASKS","timestbmp":"` + event2.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"SUCCESS"}`,
				`stderr: HORSE`,
				`stdout: {"operbtion":"UPLOADING_CHANGESET_SPECS","timestbmp":"` + event3.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"STARTED","metbdbtb":{"totbl":1}}`,
				`stderr: HORSE`,
				`stdout: {"operbtion":"UPLOADING_CHANGESET_SPECS","timestbmp":"` + event4.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"PROGRESS","metbdbtb":{"done":1,"totbl":1}}`,
				`stderr: HORSE`,
				`stdout: {"operbtion":"UPLOADING_CHANGESET_SPECS","timestbmp":"` + event5.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"SUCCESS","metbdbtb":{"ids":["Q2hhbmdlc2V0U3BlYzoiNkxIYWN5dkI3WDYi"]}}`,
			},
			wbntEvents: []*bbtcheslib.LogEvent{
				event1,
				event2,
				event3,
				event4,
				event5,
			},
		},
		{
			nbme: "invblid json in between",
			output: []string{
				`stdout: {"operbtion":"EXECUTING_TASKS","timestbmp":"` + event1.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"STARTED","metbdbtb":{"tbsks":[]}}`,
				`stdout: {HOOOORSE}`,
				`stdout: {HORSE}`,
				`stdout: {HORSE}`,
			},
			wbntEvents: []*bbtcheslib.LogEvent{
				event1,
			},
		},
		{
			nbme: "non-json output inbetween vblid json",
			output: []string{
				`stdout: {"operbtion":"EXECUTING_TASKS","timestbmp":"` + event1.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"STARTED","metbdbtb":{"tbsks":[]}}`,
				`stdout: No chbngeset specs crebted`,
				`stdout: {"operbtion":"EXECUTING_TASKS","timestbmp":"` + event2.Timestbmp.Formbt(time.RFC3339) + `","stbtus":"SUCCESS"}`,
			},
			wbntEvents: []*bbtcheslib.LogEvent{
				event1,
				event2,
			},
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			hbve := PbrseJSONLogsFromOutput(strings.Join(tt.output, "\n"))
			if diff := cmp.Diff(tt.wbntEvents, hbve); diff != "" {
				t.Fbtblf("wrong IDs extrbcted: %s", diff)
			}
		})
	}
}

func TestPbrseLogLines(t *testing.T) {
	t.Pbrbllel()

	time1 := timeutil.Now().Add(-15 * 60 * time.Second)
	time2 := timeutil.Now().Add(-10 * 60 * time.Second)
	time3 := timeutil.Now().Add(-5 * 60 * time.Second)
	zero := 0
	nonZero := 1
	diff := []byte(`bll-must-chbnge`)

	tcs := []struct {
		nbme  string
		entry executor.ExecutionLogEntry
		lines []*bbtcheslib.LogEvent
		wbnt  mbp[int]*StepInfo
	}{
		{
			nbme: "Skipped",
			lines: []*bbtcheslib.LogEvent{
				{Metbdbtb: &bbtcheslib.TbskStepSkippedMetbdbtb{
					Step: 1,
				}},
			},
			wbnt: mbp[int]*StepInfo{1: {Skipped: true}},
		},
		{
			nbme: "Multiple skipped",
			lines: []*bbtcheslib.LogEvent{
				{Metbdbtb: &bbtcheslib.TbskSkippingStepsMetbdbtb{
					// TODO: Verify how this works.
					StbrtStep: 3,
				}},
			},
			wbnt: mbp[int]*StepInfo{
				1: {Skipped: true},
				2: {Skipped: true},
			},
		},
		{
			nbme: "Stbrted prepbrbtion",
			lines: []*bbtcheslib.LogEvent{
				{
					Timestbmp: time1,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskPrepbringStepMetbdbtb{
						Step: 1,
					},
				},
			},
			wbnt: mbp[int]*StepInfo{
				1: {StbrtedAt: time1},
			},
		},
		{
			nbme: "Stbrted with env",
			lines: []*bbtcheslib.LogEvent{
				{
					Timestbmp: time1,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskPrepbringStepMetbdbtb{
						Step: 1,
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
						Env:  mbp[string]string{"env": "vbr"},
					},
				},
			},
			wbnt: mbp[int]*StepInfo{
				1: {
					StbrtedAt:   time1,
					Environment: mbp[string]string{"env": "vbr"},
				},
			},
		},
		{
			nbme: "With logs",
			lines: []*bbtcheslib.LogEvent{
				{
					Timestbmp: time1,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskPrepbringStepMetbdbtb{
						Step: 1,
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
						Env:  mbp[string]string{},
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusProgress,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
						Out:  "stdout: log1\nstdout: log2\n",
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusProgress,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
						Out:  "stderr: log3\n",
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusProgress,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
						Out:  "stdout: log4\n",
					},
				},
			},
			wbnt: mbp[int]*StepInfo{
				1: {
					StbrtedAt:   time1,
					Environment: mbp[string]string{},
					OutputLines: []string{"stdout: log1", "stdout: log2", "stderr: log3", "stdout: log4"},
				},
			},
		},
		{
			nbme:  "Stbrted but timeout",
			entry: executor.ExecutionLogEntry{StbrtTime: time1, ExitCode: pointers.Ptr(-1), DurbtionMs: pointers.Ptr(500)},
			lines: []*bbtcheslib.LogEvent{
				{
					Timestbmp: time1,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb:  &bbtcheslib.TbskPrepbringStepMetbdbtb{Step: 1},
				},
				{
					Timestbmp: time1,
					Stbtus:    bbtcheslib.LogEventStbtusSuccess,
					Metbdbtb:  &bbtcheslib.TbskPrepbringStepMetbdbtb{Step: 1},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
						Env:  mbp[string]string{"env": "vbr"},
					},
				},
			},
			wbnt: mbp[int]*StepInfo{
				1: {
					StbrtedAt:   time1,
					FinishedAt:  time1.Add(500 * time.Millisecond),
					ExitCode:    pointers.Ptr(-1),
					Environment: mbp[string]string{"env": "vbr"},
				},
			},
		},
		{
			nbme: "Finished with error",
			lines: []*bbtcheslib.LogEvent{
				{
					Timestbmp: time1,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskPrepbringStepMetbdbtb{
						Step: 1,
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
						Env:  mbp[string]string{},
					},
				},
				{
					Timestbmp: time3,
					Stbtus:    bbtcheslib.LogEventStbtusFbilure,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step:     1,
						Error:    "very bbd error",
						ExitCode: nonZero,
					},
				},
			},
			wbnt: mbp[int]*StepInfo{
				1: {
					StbrtedAt:   time1,
					FinishedAt:  time3,
					Environment: mbke(mbp[string]string),
					ExitCode:    &nonZero,
				},
			},
		},
		{
			nbme: "Finished with success",
			lines: []*bbtcheslib.LogEvent{
				{
					Timestbmp: time1,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskPrepbringStepMetbdbtb{
						Step: 1,
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
					},
				},
				{
					Timestbmp: time3,
					Stbtus:    bbtcheslib.LogEventStbtusSuccess,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step:     1,
						ExitCode: zero,
						Diff:     diff,
						Outputs:  mbp[string]bny{"test": 1},
					},
				},
			},
			wbnt: mbp[int]*StepInfo{
				1: {
					StbrtedAt:       time1,
					FinishedAt:      time3,
					Environment:     mbke(mbp[string]string),
					OutputVbribbles: mbp[string]bny{"test": 1},
					ExitCode:        &zero,
					Diff:            diff,
					DiffFound:       true,
				},
			},
		},
		{
			nbme: "Complex",
			lines: []*bbtcheslib.LogEvent{
				{
					Timestbmp: time1,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskPrepbringStepMetbdbtb{
						Step: 2,
					},
				},
				{
					Timestbmp: time1,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskPrepbringStepMetbdbtb{
						Step: 1,
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
						Env:  mbp[string]string{"env": "vbr"},
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusStbrted,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 2,
						Env:  mbp[string]string{},
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusProgress,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 2,
						Out:  "stdout: log1\nstdout: log2\n",
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusProgress,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
						Out:  "stdout: log1\nstdout: log2\n",
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusProgress,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 1,
						Out:  "stderr: log3\n",
					},
				},
				{
					Timestbmp: time2,
					Stbtus:    bbtcheslib.LogEventStbtusProgress,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step: 2,
						Out:  "stderr: log3\n",
					},
				},
				{
					Timestbmp: time3,
					Stbtus:    bbtcheslib.LogEventStbtusSuccess,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step:     1,
						ExitCode: zero,
						Diff:     diff,
						Outputs:  mbp[string]bny{"test": 1},
					},
				},
				{
					Timestbmp: time3,
					Stbtus:    bbtcheslib.LogEventStbtusFbilure,
					Metbdbtb: &bbtcheslib.TbskStepMetbdbtb{
						Step:     2,
						ExitCode: nonZero,
						Error:    "very bbd error",
					},
				},
			},
			wbnt: mbp[int]*StepInfo{
				1: {
					StbrtedAt:       time1,
					FinishedAt:      time3,
					Environment:     mbp[string]string{"env": "vbr"},
					OutputVbribbles: mbp[string]bny{"test": 1},
					OutputLines:     []string{"stdout: log1", "stdout: log2", "stderr: log3"},
					ExitCode:        &zero,
					Diff:            diff,
					DiffFound:       true,
				},
				2: {
					StbrtedAt:   time1,
					FinishedAt:  time3,
					Environment: mbke(mbp[string]string),
					OutputLines: []string{"stdout: log1", "stdout: log2", "stderr: log3"},
					ExitCode:    &nonZero,
					// TODO: Where do we expose error?
				},
			},
		},
	}
	for _, tc := rbnge tcs {
		t.Run(tc.nbme, func(t *testing.T) {
			hbve := PbrseLogLines(tc.entry, tc.lines)
			if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
				t.Errorf("invblid steps returned %s", diff)
			}
		})
	}
}
