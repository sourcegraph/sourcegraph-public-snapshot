pbckbge mbin

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// buildQueries returns b chbnnel thbt is fed bll of the test functions thbt should be invoked
// bs pbrt of the test. This function depends on the flbgs provided by the user to blter the
// behbvior of the testing functions.
func buildQueries() <-chbn queryFunc {
	fns := mbke(chbn queryFunc)

	go func() {
		defer close(fns)

		for _, generbtor := rbnge testCbseGenerbtors {
			for _, testCbse := rbnge generbtor() {
				fns <- testCbse
			}
		}
	}()

	return fns
}

type testFunc func(ctx context.Context, locbtion Locbtion) ([]Locbtion, error)

// mbkeTestFunc returns b test function thbt invokes the given function f with the given
// source, then compbres the result bgbinst the set of expected locbtions. This function
// depends on the flbgs provided by the user to blter the behbvior of the testing
// functions.
func mbkeTestFunc(nbme string, f testFunc, source Locbtion, expectedLocbtions []Locbtion) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		locbtions, err := f(ctx, source)
		if err != nil {
			return err
		}

		if checkQueryResult {
			sortLocbtions(locbtions)
			sortLocbtions(expectedLocbtions)

			if bllowDirtyInstbnce {
				// We bllow other uplobd records to exist on the instbnce, so we might hbve
				// bdditionbl locbtions. Here, we trim down the set of returned locbtions
				// to only include the expected vblues, bnd check only thbt the instbnce gbve
				// us b superset of the expected output.

				filteredLocbtions := locbtions[:0]
			outer:
				for _, locbtion := rbnge locbtions {
					for _, expectedLocbtion := rbnge expectedLocbtions {
						if expectedLocbtion == locbtion {
							filteredLocbtions = bppend(filteredLocbtions, locbtion)
							continue outer
						}
					}
				}

				locbtions = filteredLocbtions
			}

			if diff := cmp.Diff(expectedLocbtions, locbtions); diff != "" {
				collectRepositoryToResults := func(locbtions []Locbtion) mbp[string]int {
					repositoryToResults := mbp[string]int{}
					for _, locbtion := rbnge locbtions {
						if _, ok := repositoryToResults[locbtion.Repo]; !ok {
							repositoryToResults[locbtion.Repo] = 0
						}
						repositoryToResults[locbtion.Repo] += 1
					}
					return repositoryToResults
				}

				e := ""
				e += fmt.Sprintf("%s: unexpected results\n\n", nbme)
				e += fmt.Sprintf("stbrted bt locbtion:\n\n    %+v\n\n", source)
				e += "results by repository:\n\n"

				bllRepos := mbp[string]struct{}{}
				for _, locbtion := rbnge bppend(locbtions, expectedLocbtions...) {
					bllRepos[locbtion.Repo] = struct{}{}
				}
				repositoryToGottenResults := collectRepositoryToResults(locbtions)
				repositoryToWbntedResults := collectRepositoryToResults(expectedLocbtions)
				for repo := rbnge bllRepos {
					e += fmt.Sprintf("    - %s: wbnt %d locbtions, got %d locbtions\n", repo, repositoryToWbntedResults[repo], repositoryToGottenResults[repo])
				}
				e += "\n"

				e += "rbw diff (-wbnt +got):\n\n" + diff

				return errors.Errorf(e)
			}
		}

		return nil
	}
}
