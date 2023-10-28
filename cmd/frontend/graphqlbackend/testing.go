package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sort"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func mustParseGraphQLSchema(t *testing.T, db database.DB) *graphql.Schema {
	return mustParseGraphQLSchemaWithClient(t, db, gitserver.NewClient("graphql.test"))
}

func mustParseGraphQLSchemaWithClient(t *testing.T, db database.DB, gitserverClient gitserver.Client) *graphql.Schema {
	t.Helper()

	parsedSchema, parseSchemaErr := NewSchema(
		db,
		gitserverClient,
		[]OptionalResolver{},
		graphql.PanicHandler(printStackTrace{&gqlerrors.DefaultPanicHandler{}}),
	)
	if parseSchemaErr != nil {
		t.Fatal(parseSchemaErr)
	}

	return parsedSchema
}

// Code below copied from graph-gophers and has been modified to improve
// error messages

// Test is a GraphQL test case to be used with RunTest(s).
type Test struct {
	Context        context.Context
	Schema         *graphql.Schema
	Query          string
	OperationName  string
	Variables      map[string]any
	ExpectedResult string
	ExpectedErrors []*gqlerrors.QueryError
	Label          string
}

// RunTests runs the given GraphQL test cases as subtests.
func RunTests(t *testing.T, tests []*Test) {
	t.Helper()

	if len(tests) == 1 {
		RunTest(t, tests[0])
		return
	}

	for i, test := range tests {
		testName := strconv.Itoa(i + 1)
		if test.Label != "" {
			testName = fmt.Sprintf("%s/%s", testName, test.Label)
		}
		t.Run(testName, func(t *testing.T) {
			t.Helper()
			RunTest(t, test)
		})
	}
}

// RunTest runs a single GraphQL test case.
func RunTest(t *testing.T, test *Test) {
	t.Helper()

	if test.Context == nil {
		test.Context = context.Background()
	}
	result := test.Schema.Exec(test.Context, test.Query, test.OperationName, test.Variables)

	checkErrors(t, test.ExpectedErrors, result.Errors)

	if test.ExpectedResult == "" {
		if result.Data != nil {
			t.Logf("%v", test)
			t.Errorf("got: %s", result.Data)
			t.Fatal("want: null")
		}
		return
	}

	// Verify JSON to avoid red herring errors.
	got, err := formatJSON(result.Data)
	if err != nil {
		t.Fatalf("got: invalid JSON: %s\n\n%v", err, result.Data)
	}
	want, err := formatJSON([]byte(test.ExpectedResult))
	if err != nil {
		t.Fatalf("want: invalid JSON: %s\n\n%s", err, test.ExpectedResult)
	}

	require.JSONEq(t, string(want), string(got))
}

func formatJSON(data []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	formatted, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return formatted, nil
}

func checkErrors(t *testing.T, want, got []*gqlerrors.QueryError) {
	t.Helper()

	sortErrors(want)
	sortErrors(got)

	// Compare without caring about the concrete type of the error returned
	if diff := cmp.Diff(want, got, cmpopts.IgnoreFields(gqlerrors.QueryError{}, "ResolverError", "Err")); diff != "" {
		t.Fatal(diff)
	}
}

func sortErrors(errs []*gqlerrors.QueryError) {
	if len(errs) <= 1 {
		return
	}
	sort.Slice(errs, func(i, j int) bool {
		return fmt.Sprintf("%s", errs[i].Path) < fmt.Sprintf("%s", errs[j].Path)
	})
}

// printStackTrace wraps panic recovery from given Handler and prints the stack trace.
type printStackTrace struct {
	Handler gqlerrors.PanicHandler
}

func (t printStackTrace) MakePanicError(ctx context.Context, value interface{}) *gqlerrors.QueryError {
	debug.PrintStack()
	return t.Handler.MakePanicError(ctx, value)
}
