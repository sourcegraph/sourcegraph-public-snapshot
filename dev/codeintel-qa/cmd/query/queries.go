package main

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// buildQueries returns a channel that is fed all of the test functions that should be invoked
// as part of the test. This function depends on the flags provided by the user to alter the
// behavior of the testing functions.
func buildQueries() <-chan queryFunc {
	fns := make(chan queryFunc)

	go func() {
		defer close(fns)

		for _, generator := range testCaseGenerators {
			for _, testCase := range generator() {
				fns <- testCase
			}
		}
	}()

	return fns
}

type testFunc func(ctx context.Context, location Location) ([]Location, error)

// makeTestFunc returns a test function that invokes the given function f with the given
// source, then compares the result against the set of expected locations. This function
// depends on the flags provided by the user to alter the behavior of the testing
// functions.
func makeTestFunc(name string, f testFunc, source Location, expectedLocations []Location) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		locations, err := f(ctx, source)
		if err != nil {
			return err
		}

		if checkQueryResult {
			sortLocations(locations)
			sortLocations(expectedLocations)

			if allowDirtyInstance {
				// We allow other upload records to exist on the instance, so we might have
				// additional locations. Here, we trim down the set of returned locations
				// to only include the expected values, and check only that the instance gave
				// us a superset of the expected output.

				filteredLocations := locations[:0]
			outer:
				for _, location := range locations {
					for _, expectedLocation := range expectedLocations {
						if expectedLocation == location {
							filteredLocations = append(filteredLocations, location)
							continue outer
						}
					}
				}

				locations = filteredLocations
			}

			if diff := cmp.Diff(expectedLocations, locations); diff != "" {
				collectRepositoryToResults := func(locations []Location) map[string]int {
					repositoryToResults := map[string]int{}
					for _, location := range locations {
						if _, ok := repositoryToResults[location.Repo]; !ok {
							repositoryToResults[location.Repo] = 0
						}
						repositoryToResults[location.Repo] += 1
					}
					return repositoryToResults
				}

				e := ""
				e += fmt.Sprintf("%s: unexpected results\n\n", name)
				e += fmt.Sprintf("started at location:\n\n    %+v\n\n", source)
				e += "results by repository:\n\n"

				allRepos := map[string]struct{}{}
				for _, location := range append(locations, expectedLocations...) {
					allRepos[location.Repo] = struct{}{}
				}
				repositoryToGottenResults := collectRepositoryToResults(locations)
				repositoryToWantedResults := collectRepositoryToResults(expectedLocations)
				for repo := range allRepos {
					e += fmt.Sprintf("    - %s: want %d locations, got %d locations\n", repo, repositoryToWantedResults[repo], repositoryToGottenResults[repo])
				}
				e += "\n"

				e += "raw diff (-want +got):\n\n" + diff

				return errors.Errorf(e)
			}
		}

		return nil
	}
}
