package loki

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
)

func TestNewStreamFromJobLogs(t *testing.T) {
	type args struct {
		log string
	}
	tests := []struct {
		name    string
		args    args
		want    [][2]string
		wantErr bool
	}{
		{
			name: "parse line",
			args: args{
				log: `_bk;t=1633575941106~~~ Preparing working directory`,
			},
			want: [][2]string{
				{"1633575941106000000", "~~~ Preparing working directory"},
			},
		},
		{
			name: "merge timestamps",
			args: args{
				log: `_bk;t=1633575941106~~~ Preparing working directory
_bk;t=1633575941106$ cd /buildkite/builds/buildkite-agent-77bfc969fc-4zfqc-1/sourcegraph/sourcegraph
_bk;t=1633575941112$ git remote set-url origin git@github.com:sourcegraph/sourcegraph.git
_bk;t=1633575946276remote: Enumerating objects: 25, done._bk;t=1633575947202

_bk;t=1633575947202remote: Counting objects:   4% (1/25)_bk;t=1633575947202
remote: Counting objects:   8% (2/25)_bk;t=1633575947202`,
			},
			want: [][2]string{
				{"1633575941106000000", "~~~ Preparing working directory\n$ cd /buildkite/builds/buildkite-agent-77bfc969fc-4zfqc-1/sourcegraph/sourcegraph"},
				{"1633575941112000000", "$ git remote set-url origin git@github.com:sourcegraph/sourcegraph.git"},
				{"1633575946276000000", "remote: Enumerating objects: 25, done."},
				{"1633575947202000000", "remote: Counting objects:   4% (1/25)\nremote: Counting objects:   8% (2/25)"},
			},
		},
		{
			name: "weird ansi things",
			args: args{
				log: `_bk;t=1633575951822[38;5;48m2021-10-07 03:05:51 INFO  [0m [0mUpdating BUILDKITE_COMMIT to "d4b6e13eab2216ea2a934607df5c97a25e920207"[0m

_bk;t=1633575951838[38;5;48m2021-10-07 03:05:54 INFO  [0m [0mSuccessfully uploaded and parsed pipeline config[0m`,
			},
			want: [][2]string{
				{"1633575951822000000", "2021-10-07 03:05:51 INFO   Updating BUILDKITE_COMMIT to \"d4b6e13eab2216ea2a934607df5c97a25e920207\""},
				{"1633575951838000000", "2021-10-07 03:05:54 INFO   Successfully uploaded and parsed pipeline config"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStreamFromJobLogs(&bk.JobLogs{
				JobMeta: bk.JobMeta{Job: tt.name},
				Content: &tt.args.log,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStreamFromJobLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got.Values); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
		})
	}
}
