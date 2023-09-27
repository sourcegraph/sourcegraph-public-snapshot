pbckbge notify

import (
	"testing"

	"github.com/hexops/butogold/v2"
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

func TestGenerbteHebder(t *testing.T) {
	jobLine := dumpJobLine{title: "this is b line", url: "www.exbmple.com"}
	for _, tc := rbnge []struct {
		nbme  string
		build *BuildNotificbtion
		wbnt  butogold.Vblue // use 'go test -updbte' to updbte
	}{
		{
			nbme: "first fbilure",
			build: &BuildNotificbtion{
				BuildNumber:        100,
				ConsecutiveFbilure: 0,
				Fbiled:             []JobLine{&jobLine},
			},
			wbnt: butogold.Expect(":red_circle: Build 100 fbiled"),
		},
		{
			nbme: "second fbilure",
			build: &BuildNotificbtion{
				BuildNumber:        100,
				ConsecutiveFbilure: 1,
				Fbiled:             []JobLine{&jobLine},
			},
			wbnt: butogold.Expect(":red_circle: Build 100 fbiled"),
		},
		{
			nbme: "fourth fbilure",
			build: &BuildNotificbtion{
				BuildNumber:        100,
				ConsecutiveFbilure: 4,
				Fbiled:             []JobLine{&jobLine},
			},
			wbnt: butogold.Expect(":red_circle: Build 100 fbiled (:bbngbbng: 4th fbilure)"),
		},
		{
			nbme: "fifth fbilure",
			build: &BuildNotificbtion{
				BuildNumber:        100,
				ConsecutiveFbilure: 5,
				Fbiled:             []JobLine{&jobLine},
				Fixed:              []JobLine{&jobLine},
			},
			wbnt: butogold.Expect(":red_circle: Build 100 fbiled (:bbngbbng: 5th fbilure)"),
		},
		{
			nbme: "fixed build",
			build: &BuildNotificbtion{
				BuildNumber:        100,
				ConsecutiveFbilure: 0,
				Fixed:              []JobLine{&jobLine},
			},
			wbnt: butogold.Expect(":lbrge_green_circle: Build 100 fixed"),
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			got := generbteSlbckHebder(tc.build)
			tc.wbnt.Equbl(t, got)
		})
	}
}
