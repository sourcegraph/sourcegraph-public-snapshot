package main

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
)

// buildQueries returns a channel that is fed all of the test functions that should be invoked
// as part of the test. This function depends on the flags provided by the user to alter the
// behavior of the testing functions.
func buildQueries() <-chan queryFunc {
	fns := make(chan queryFunc)

	go func() {
		defer close(fns)

		for _, testCase := range testCases {
			// Definition returns defintion
			fns <- makeTestFunc(queryDefinitions, testCase.Definition, []Location{testCase.Definition})

			// References return definition
			for _, reference := range testCase.References {
				fns <- makeTestFunc(queryDefinitions, reference, []Location{testCase.Definition})
			}

			// Definition returns references
			fns <- makeTestFunc(queryReferences, testCase.Definition, testCase.References)

			// References return references
			if queryReferencesOfReferences {
				for _, reference := range testCase.References {
					fns <- makeTestFunc(queryReferences, reference, testCase.References)
				}
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
func makeTestFunc(f testFunc, source Location, expectedLocations []Location) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		locations, err := f(ctx, source)
		if err != nil {
			return err
		}

		if checkQueryResult {
			sortLocations(locations)

			if diff := cmp.Diff(expectedLocations, locations); diff != "" {
				return errors.Errorf("unexpected locations (-want +got):\n%s", diff)
			}
		}

		return nil
	}
}
