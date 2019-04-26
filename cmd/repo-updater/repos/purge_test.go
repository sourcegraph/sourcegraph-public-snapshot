package repos

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

func Test_shouldKeep(t *testing.T) {
	type args struct {
		info *protocol.RepoInfo
	}
	tests := []struct {
		name  string
		args  args
		wantKeep  bool
		wantAge time.Duration
	}{
		{
			name: "nil info",
			args: args { info: nil },
			// If we don't know anything about it then the answer will never change, so if we keep it then
			// we will have to keep it forever.
			wantKeep: false,
		},
		{
			name: "default fields",
			args: args { info: &protocol.RepoInfo{} },
			// We shouldn't keep in this case for the same reason as "nil info" above.
			wantKeep: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := shouldKeep(tt.args.info)
			if got != tt.wantKeep {
				t.Errorf("shouldKeep() got = %v, want %v", got, tt.wantKeep)
			}
			if got1 != tt.wantAge {
				t.Errorf("shouldKeep() got1 = %v, want %v", got1, tt.wantAge)
			}
		})
	}
}
