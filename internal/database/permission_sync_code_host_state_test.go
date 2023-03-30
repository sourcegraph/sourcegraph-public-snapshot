package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodeHostStatusesSet_CountStatuses(t *testing.T) {
	tests := map[string]struct {
		statuses    CodeHostStatusesSet
		wantTotal   int
		wantSuccess int
		wantFailed  int
	}{
		"zero statuses present": {
			statuses:    CodeHostStatusesSet{},
			wantTotal:   0,
			wantSuccess: 0,
			wantFailed:  0,
		},
		"all successful": {
			statuses:    generateStatuses(5, 5),
			wantTotal:   5,
			wantSuccess: 5,
			wantFailed:  0,
		},
		"all failed": {
			statuses:    generateStatuses(5, 0),
			wantTotal:   5,
			wantSuccess: 0,
			wantFailed:  5,
		},
		"mixed results": {
			statuses:    generateStatuses(5, 3),
			wantTotal:   5,
			wantSuccess: 3,
			wantFailed:  2,
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			gotTotal, gotSuccess, gotFailed := testCase.statuses.CountStatuses()
			require.Equal(t, testCase.wantTotal, gotTotal)
			require.Equal(t, testCase.wantSuccess, gotSuccess)
			require.Equal(t, testCase.wantFailed, gotFailed)
		})
	}
}

func generateStatuses(total, success int) CodeHostStatusesSet {
	codeHostStatuses := make(CodeHostStatusesSet, 0, total)
	for i, success := 0, success; i < total; i, success = i+1, success-1 {
		status := CodeHostStatusError
		if success > 0 {
			status = CodeHostStatusSuccess
		}
		codeHostStatuses = append(codeHostStatuses, PermissionSyncCodeHostState{Status: status})
	}
	return codeHostStatuses
}
