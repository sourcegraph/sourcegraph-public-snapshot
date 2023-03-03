package notify

import (
	"testing"

	"github.com/hexops/autogold"
)

type dumpJobLine struct {
	title string
	url   string
}

func (d *dumpJobLine) Title() string {
	return d.title
}

func (d *dumpJobLine) LogURL() string {
	return d.url
}

func TestGenerateHeader(t *testing.T) {
	jobLine := dumpJobLine{title: "this is a line", url: "www.example.com"}
	for _, tc := range []struct {
		build *BuildNotification
		want  autogold.Value // use 'go test -update' to update
	}{
		{
			build: &BuildNotification{
				BuildNumber:        100,
				ConsecutiveFailure: 0,
				Failed:             []JobLine{&jobLine},
			},
			want: autogold.Want("first failure", ":red_circle: Build 100 failed"),
		},
		{
			build: &BuildNotification{
				BuildNumber:        100,
				ConsecutiveFailure: 1,
				Failed:             []JobLine{&jobLine},
			},
			want: autogold.Want("second failure", ":red_circle: Build 100 failed"),
		},
		{
			build: &BuildNotification{
				BuildNumber:        100,
				ConsecutiveFailure: 4,
				Failed:             []JobLine{&jobLine},
			},
			want: autogold.Want("fifth failure", ":red_circle: Build 100 failed (:bangbang: 4th failure)"),
		},
		{
			build: &BuildNotification{
				BuildNumber:        100,
				ConsecutiveFailure: 4,
				Failed:             []JobLine{&jobLine},
				Fixed:              []JobLine{&jobLine},
			},
			want: autogold.Want("fifth failure", ":red_circle: Build 100 failed (:bangbang: 4th failure)"),
		},
		{
			build: &BuildNotification{
				BuildNumber:        100,
				ConsecutiveFailure: 4,
				Fixed:              []JobLine{&jobLine},
			},
			want: autogold.Want("fifth failure", ":large_green_circle: Build 100 fixed"),
		},
	} {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := generateSlackHeader(tc.build)
			tc.want.Equal(t, got)
		})
	}
}
