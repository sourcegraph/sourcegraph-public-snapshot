package gqltesting

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/errors"
)

// Test is a GraphQL test case to be used with RunTest(s).
type Test struct {
	Context        context.Context
	Schema         *graphql.Schema
	Query          string
	OperationName  string
	Variables      map[string]interface{}
	ExpectedResult string
	ExpectedErrors []*errors.QueryError
}

// RunTests runs the given GraphQL test cases as subtests.
func RunTests(t *testing.T, tests []*Test) {
	if len(tests) == 1 {
		RunTest(t, tests[0])
		return
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			RunTest(t, test)
		})
	}
}

// RunTest runs a single GraphQL test case.
func RunTest(t *testing.T, test *Test) {
	if test.Context == nil {
		test.Context = context.Background()
	}
	result := test.Schema.Exec(test.Context, test.Query, test.OperationName, test.Variables)
	// Verify JSON to avoid red herring errors.
	got, err := formatJSON(result.Data)
	if err != nil {
		t.Fatalf("got: invalid JSON: %s", err)
	}
	want, err := formatJSON([]byte(test.ExpectedResult))
	if err != nil {
		t.Fatalf("want: invalid JSON: %s", err)
	}

	checkErrors(t, test.ExpectedErrors, result.Errors)

	if !bytes.Equal(got, want) {
		t.Logf("got:  %s", got)
		t.Logf("want: %s", want)
		t.Fail()
	}
}

func formatJSON(data []byte) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	formatted, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return formatted, nil
}

func checkErrors(t *testing.T, expected, actual []*errors.QueryError) {
	expectedCount, actualCount := len(expected), len(actual)

	if expectedCount != actualCount {
		t.Fatalf("unexpected number of errors: got %d, want %d", actualCount, expectedCount)
	}

	if expectedCount > 0 {
		for i, want := range expected {
			got := actual[i]

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("unexpected error: got %+v, want %+v", got, want)
			}
		}

		// Return because we're done checking.
		return
	}

	for _, err := range actual {
		t.Errorf("unexpected error: '%s'", err)
	}
}
