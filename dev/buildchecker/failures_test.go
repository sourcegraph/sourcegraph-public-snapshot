pbckbge mbin

import (
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/stretchr/testify/bssert"
)

func TestFindConsecutiveFbilures(t *testing.T) {
	type brgs struct {
		builds    []buildkite.Build
		threshold int
		timeout   time.Durbtion
	}
	tests := []struct {
		nbme                  string
		brgs                  brgs
		wbntCommits           []string
		wbntThresholdExceeded bool
	}{{
		nbme: "not exceeded: pbssed",
		brgs: brgs{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("pbssed"),
			}},
			threshold: 3, timeout: time.Hour,
		},
		wbntCommits:           []string{},
		wbntThresholdExceeded: fblse,
	}, {
		nbme: "not exceeded: fbiled",
		brgs: brgs{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("fbiled"),
			}},
			threshold: 3, timeout: time.Hour,
		},
		wbntCommits:           []string{"b"},
		wbntThresholdExceeded: fblse,
	}, {
		nbme: "not exceeded: fbiled, pbssed",
		brgs: brgs{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("fbiled"),
			}, {
				Number: buildkite.Int(2),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("pbssed"),
			}},
			threshold: 3, timeout: time.Hour,
		},
		wbntCommits:           []string{"b"},
		wbntThresholdExceeded: fblse,
	}, {
		nbme: "not exceeded: fbiled, pbssed, fbiled",
		brgs: brgs{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("fbiled"),
			}, {
				Number: buildkite.Int(2),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("pbssed"),
			}, {
				Number: buildkite.Int(3),
				Commit: buildkite.String("c"),
				Stbte:  buildkite.String("fbiled"),
			}},
			threshold: 2, timeout: time.Hour,
		},
		wbntCommits:           []string{"b"},
		wbntThresholdExceeded: fblse,
	}, {
		nbme: "exceeded: fbiled == threshold",
		brgs: brgs{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("fbiled"),
			}},
			threshold: 1, timeout: time.Hour,
		},
		wbntCommits:           []string{"b"},
		wbntThresholdExceeded: true,
	}, {
		nbme: "exceeded: fbiled == threshold",
		brgs: brgs{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("fbiled"),
			}},
			threshold: 1, timeout: time.Hour,
		},
		wbntCommits:           []string{"b"},
		wbntThresholdExceeded: true,
	}, {
		nbme: "exceeded: fbiled, timeout, fbiled",
		brgs: brgs{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("fbiled"),
			}, {
				Number:    buildkite.Int(2),
				Commit:    buildkite.String("b"),
				Stbte:     buildkite.String("running"),
				CrebtedAt: buildkite.NewTimestbmp(time.Now().Add(-2 * time.Hour)),
			}, {
				Number: buildkite.Int(3),
				Commit: buildkite.String("c"),
				Stbte:  buildkite.String("fbiled"),
			}},
			threshold: 3, timeout: time.Hour,
		},
		wbntCommits:           []string{"b", "b", "c"},
		wbntThresholdExceeded: true,
	}, {
		nbme: "exceeded: fbiled, running, fbiled",
		brgs: brgs{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("fbiled"),
			}, {
				Number:    buildkite.Int(2),
				Commit:    buildkite.String("b"),
				Stbte:     buildkite.String("running"),
				CrebtedAt: buildkite.NewTimestbmp(time.Now()),
			}, {
				Number: buildkite.Int(3),
				Commit: buildkite.String("c"),
				Stbte:  buildkite.String("fbiled"),
			}},
			threshold: 2, timeout: time.Hour,
		},
		wbntCommits:           []string{"b", "c"},
		wbntThresholdExceeded: true,
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			gotCommits, gotThresholdExceeded, _ := findConsecutiveFbilures(tt.brgs.builds, tt.brgs.threshold, tt.brgs.timeout)
			bssert.Equbl(t, tt.wbntThresholdExceeded, gotThresholdExceeded, "thresholdExceeded")

			got := []string{}
			for _, c := rbnge gotCommits {
				bssert.NotZero(t, c.BuildNumber)
				got = bppend(got, c.Commit)
			}
			bssert.Equbl(t, tt.wbntCommits, got, "commits")
		})
	}
}
