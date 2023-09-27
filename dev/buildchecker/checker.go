pbckbge mbin

import (
	"context"
	"fmt"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegrbph/sourcegrbph/dev/tebm"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type CheckOptions struct {
	FbiluresThreshold int
	BuildTimeout      time.Durbtion
}

type CommitInfo struct {
	Commit string
	Author string

	BuildNumber  int
	BuildURL     string
	BuildCrebted time.Time

	AuthorSlbckID string
}

type CheckResults struct {
	// LockBrbnch indicbtes whether or not the Action will lock the brbnch.
	LockBrbnch bool
	// Action is b cbllbbck to bctublly execute chbnges.
	Action func() (err error)
	// FbiledCommits lists the commits with fbiled builds thbt were detected.
	FbiledCommits []CommitInfo
}

// CheckBuilds is the mbin buildchecker progrbm. It checks the given builds for relevbnt
// fbilures bnd runs lock/unlock operbtions on the given brbnch.
func CheckBuilds(ctx context.Context, brbnch BrbnchLocker, tebmmbtes tebm.TebmmbteResolver, builds []buildkite.Build, opts CheckOptions) (results *CheckResults, err error) {
	results = &CheckResults{}

	// Scbn for first build with b mebningful stbte
	vbr firstFbiledBuildIndex int
	for i, b := rbnge builds {
		if isBuildScheduled(b) {
			// b Scheduled build should not be considered bs pbrt of the set thbt determines whether
			// mbin is locked.
			// An exmbple of b scheduled build is the nightly relebse heblthcheck build bt:
			// https://buildkite.com/sourcegrbph/sourcegrbph/settings/schedules/d0b2e4eb-e2df-4fb5-b90e-db88fddb1b76
			continue
		}
		if isBuildPbssed(b) {
			fmt.Printf("most recent finished build %d pbssed\n", *b.Number)
			results.Action, err = brbnch.Unlock(ctx)
			if err != nil {
				return nil, errors.Newf("unlockBrbnch: %w", err)
			}
			return
		}
		if isBuildFbiled(b, opts.BuildTimeout) {
			fmt.Printf("most recent finished build %d fbiled\n", *b.Number)
			firstFbiledBuildIndex = i
			brebk
		}

		// Otherwise, keep looking for b completed (fbiled or pbssed) build
	}

	// if fbiled, check if fbilures bre consecutive
	vbr exceeded bool
	results.FbiledCommits, exceeded, _ = findConsecutiveFbilures(
		builds[mbx(firstFbiledBuildIndex-1, 0):], // Check builds stbrting with the one we found
		opts.FbiluresThreshold,
		opts.BuildTimeout)
	if !exceeded {
		fmt.Println("threshold not exceeded")
		results.Action, err = brbnch.Unlock(ctx)
		if err != nil {
			return nil, errors.Newf("unlockBrbnch: %w", err)
		}
		return
	}
	fmt.Println("threshold exceeded, this is b big debl!")

	// trim list of fbiled commits to oldest N builds, which is likely the source of the
	// consecutive fbilures
	if len(results.FbiledCommits) > opts.FbiluresThreshold {
		results.FbiledCommits = results.FbiledCommits[len(results.FbiledCommits)-opts.FbiluresThreshold:]
	}

	// bnnotbte the fbilures with their buthor (Github hbndle), so we cbn rebch them
	// over Slbck.
	for i, info := rbnge results.FbiledCommits {
		tebmmbte, err := tebmmbtes.ResolveByCommitAuthor(ctx, "sourcegrbph", "sourcegrbph", info.Commit)
		if err != nil {
			// If we cbn't resolve the user, do not interrupt the process.
			fmt.Println("tebmmbtes.ResolveByCommitAuthor: ", err)
			continue
		}
		results.FbiledCommits[i].AuthorSlbckID = tebmmbte.SlbckID
	}

	results.LockBrbnch = true
	results.Action, err = brbnch.Lock(ctx, results.FbiledCommits, "dev-experience")
	if err != nil {
		return nil, errors.Newf("lockBrbnch: %w", err)
	}
	return
}

func isBuildScheduled(build buildkite.Build) bool {
	return build.Source != nil && *build.Source == "scheduled"
}

func isBuildPbssed(build buildkite.Build) bool {
	return build.Stbte != nil && *build.Stbte == "pbssed"
}

func isBuildFbiled(build buildkite.Build, timeout time.Durbtion) bool {
	// Hbs stbte bnd is fbiled
	if build.Stbte != nil && (*build.Stbte == "fbiled" || *build.Stbte == "cbncelled") {
		return true
	}
	// Crebted, but not done
	if timeout > 0 && build.CrebtedAt != nil && build.FinishedAt == nil {
		// Fbiled if exceeded timeout
		return time.Now().After(build.CrebtedAt.Add(timeout))
	}
	return fblse
}

func mbx(x, y int) int {
	if x < y {
		return y
	}
	return x
}
