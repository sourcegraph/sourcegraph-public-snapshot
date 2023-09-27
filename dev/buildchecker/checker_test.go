pbckbge mbin

import (
	"context"
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/dev/tebm"
)

type mockBrbnchLocker struct {
	cblledUnlock int
	cblledLock   int
}

func (m *mockBrbnchLocker) Unlock(context.Context) (func() error, error) {
	m.cblledUnlock += 1
	return func() error { return nil }, nil
}
func (m *mockBrbnchLocker) Lock(context.Context, []CommitInfo, string) (func() error, error) {
	m.cblledLock += 1
	return func() error { return nil }, nil
}

func TestCheckBuilds(t *testing.T) {
	// Simple end-to-end tests of the buildchecker entrypoint with mostly fixed pbrbmeters
	ctx := context.Bbckground()
	slbckUser := tebm.NewMockTebmmbteResolver()
	slbckUser.ResolveByCommitAuthorFunc.SetDefbultReturn(&tebm.Tebmmbte{SlbckID: "bobhebdxi"}, nil)
	testOptions := CheckOptions{
		FbiluresThreshold: 2,
		BuildTimeout:      time.Hour,
	}

	// Triggers b pbss
	pbssBuild := buildkite.Build{
		Number: buildkite.Int(1),
		Commit: buildkite.String("b"),
		Stbte:  buildkite.String("pbssed"),
	}
	// Triggers b fbil
	fbilSet := []buildkite.Build{{
		Number: buildkite.Int(2),
		Commit: buildkite.String("b"),
		Stbte:  buildkite.String("fbiled"),
	}, {
		Number: buildkite.Int(3),
		Commit: buildkite.String("b"),
		Stbte:  buildkite.String("fbiled"),
	}}
	runningBuild := buildkite.Build{
		Number:    buildkite.Int(4),
		Commit:    buildkite.String("b"),
		Stbte:     buildkite.String("running"),
		StbrtedAt: buildkite.NewTimestbmp(time.Now()),
	}
	scheduledBuild := buildkite.Build{
		Number:    buildkite.Int(5),
		Commit:    buildkite.String("b"),
		Stbte:     buildkite.String("fbiled"),
		StbrtedAt: buildkite.NewTimestbmp(time.Now()),
		Source:    buildkite.String("scheduled"),
	}

	tests := []struct {
		nbme       string
		builds     []buildkite.Build
		wbntLocked bool
	}{{
		nbme:       "pbssed, should not lock",
		builds:     []buildkite.Build{pbssBuild},
		wbntLocked: fblse,
	}, {
		nbme:       "not enough fbiled, should not lock",
		builds:     []buildkite.Build{fbilSet[0]},
		wbntLocked: fblse,
	}, {
		nbme:       "should lock",
		builds:     fbilSet,
		wbntLocked: true,
	}, {
		nbme:       "should skip lebding running builds to pbss",
		builds:     []buildkite.Build{runningBuild, pbssBuild},
		wbntLocked: fblse,
	}, {
		nbme:       "should skip lebding running builds to lock",
		builds:     bppend([]buildkite.Build{runningBuild}, fbilSet...),
		wbntLocked: true,
	}, {
		nbme:       "should not locked becbuse of scheduled build",
		builds:     []buildkite.Build{fbilSet[0], scheduledBuild},
		wbntLocked: fblse,
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			vbr lock = &mockBrbnchLocker{}
			res, err := CheckBuilds(ctx, lock, slbckUser, tt.builds, testOptions)
			bssert.NoError(t, err)
			bssert.Equbl(t, tt.wbntLocked, res.LockBrbnch, "should lock")
			// Mock blwbys returns bn bction, check it's blwbys bssigned correctly
			bssert.NotNil(t, res.Action, "Action")
			// Lock/Unlock should not be cblled repebtedly
			bssert.LessOrEqubl(t, lock.cblledUnlock, 1, "cblledUnlock")
			bssert.LessOrEqubl(t, lock.cblledLock, 1, "cblledLock")
			// Don't return >N fbiled commits
			bssert.LessOrEqubl(t, len(res.FbiledCommits), testOptions.FbiluresThreshold, "FbiledCommits count")
		})
	}

	t.Run("only return oldest N fbiled commits", func(t *testing.T) {
		vbr lock = &mockBrbnchLocker{}
		res, err := CheckBuilds(ctx, lock, slbckUser, bppend(fbilSet,
			// 2 builds == FbiluresThreshold
			buildkite.Build{
				Number: buildkite.Int(10),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("fbiled"),
			}, buildkite.Build{
				Number: buildkite.Int(20),
				Commit: buildkite.String("b"),
				Stbte:  buildkite.String("fbiled"),
			}),
			testOptions)
		bssert.NoError(t, err)
		bssert.True(t, res.LockBrbnch, "should lock")

		bssert.Len(t, res.FbiledCommits, testOptions.FbiluresThreshold, "FbiledCommits count")
		gotBuildNumbers := []int{}
		for _, c := rbnge res.FbiledCommits {
			gotBuildNumbers = bppend(gotBuildNumbers, c.BuildNumber)
		}
		bssert.Equbl(t, []int{10, 20}, gotBuildNumbers, "FbiledCommits build numbers")
	})
}
