package template

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var partialEvalStepCtx = &StepContext{
	BatchChange: BatchChangeAttributes{
		Name:        "test-batch-change",
		Description: "test-description",
	},
	// Step is not set when evalStepCondition is called
	Repository: Repository{
		Name: "github.com/sourcegraph/src-cli",
		FileMatches: []string{
			"main.go", "README.md",
		},
	},
}

func runParseAndPartialTest(t *testing.T, in, want string) {
	t.Helper()

	tmpl, err := parseAndPartialEval(in, partialEvalStepCtx)
	if err != nil {
		t.Fatal(err)
	}

	tmplStr := tmpl.Tree.Root.String()
	if tmplStr != want {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, tmplStr))
	}
}

func TestParseAndPartialEval(t *testing.T) {
	t.Run("evaluated", func(t *testing.T) {
		for _, tt := range []struct{ input, want string }{
			{
				// Literal constants:
				`this is my template ${{ "hardcoded string" }}`,
				`this is my template hardcoded string`,
			},
			{
				`${{ 1234 }}`,
				`1234`,
			},
			{
				`${{ true }} ${{ false }}`,
				`true false`,
			},
			{
				// Evaluated, since they're static values:
				`${{ repository.name }} ${{ batch_change.name }} ${{ batch_change.description }}`,
				`github.com/sourcegraph/src-cli test-batch-change test-description`,
			},
			{
				`AAA${{ repository.name }}BBB${{ batch_change.name }}CCC${{ batch_change.description }}DDD`,
				`AAAgithub.com/sourcegraph/src-cliBBBtest-batch-changeCCCtest-descriptionDDD`,
			},
			{
				// Function call with static value and runtime value:
				`${{ eq repository.name outputs.repo.name }}`,
				// Aborts, since one of them is runtime value
				`{{eq repository.name outputs.repo.name}}`,
			},
			{
				// "eq" call with 2 static values:
				`${{ eq repository.name "github.com/sourcegraph/src-cli" }}`,
				`true`,
			},
			{
				// "eq" call with 2 literal values:
				`${{ eq 5 5 }}`,
				`true`,
			},
			{
				// "not" call:
				`${{ not (eq repository.name "bitbucket-repo") }}`,
				`true`,
			},
			{
				// "not" call:
				`${{ not 1234 }} ${{ not false }} ${{ not true }}`,
				`false true false`,
			},
			{
				// "ne" call with 2 static values:
				`${{ ne repository.name "github.com/sourcegraph/src-cli" }}`,
				`false`,
			},
			{
				// "ne" call with 2 literal values:
				`${{ ne 5 5 }}`,
				`false`,
			},
			{
				// Function call with builtin function and static values:
				`${{ matches repository.name "github.com*" }}`,
				`true`,
			},
			{
				// Nested function call with builtin function and static values:
				`${{ eq false (matches repository.name "github.com*") }}`,
				`false`,
			},
			{
				// Nested nested function call with builtin function and static values:
				`${{ eq false (eq false (matches repository.name "github.com*")) }}`,
				`true`,
			},
		} {
			runParseAndPartialTest(t, tt.input, tt.want)
		}
	})

	t.Run("aborted", func(t *testing.T) {
		for _, tt := range []struct{ input, want string }{
			{
				// Field that doesn't exist
				`${{ repository.secretlocation }}`,
				`{{repository.secretlocation}}`,
			},
			{
				// Field access that's too deep
				`${{ repository.name.doesnotexist }}`,
				`{{repository.name.doesnotexist}}`,
			},
			{
				// Complex value
				`${{ repository.search_result_paths }}`,
				// String representation of templates uses standard delimiters
				`{{repository.search_result_paths}}`,
			},
			{
				// Runtime value
				`${{ outputs.runtime.value }}`,
				`{{outputs.runtime.value}}`,
			},
			{
				// Runtime value
				`${{ step.modified_files }}`,
				`{{step.modified_files}}`,
			},
			{
				// Runtime value
				`${{ previous_step.modified_files }}`,
				`{{previous_step.modified_files}}`,
			},
			{
				// "eq" call with static value and runtime value:
				`${{ eq repository.name outputs.repo.name }}`,
				// Aborts, since one of them is runtime value
				`{{eq repository.name outputs.repo.name}}`,
			},
			{
				// "eq" call with more than 2 arguments:
				`${{ eq repository.name "github.com/sourcegraph/src-cli" "github.com/sourcegraph/sourcegraph" }}`,
				`{{eq repository.name "github.com/sourcegraph/src-cli" "github.com/sourcegraph/sourcegraph"}}`,
			},
			{
				// Nested nested function call with builtin function but runtime values:
				`${{ eq false (eq false (matches outputs.runtime.value "github.com*")) }}`,
				`{{eq false (eq false (matches outputs.runtime.value "github.com*"))}}`,
			},
		} {
			runParseAndPartialTest(t, tt.input, tt.want)
		}
	})
}

func TestParseAndPartialEval_BuiltinFunctions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for _, tt := range []struct{ input, want string }{
			{
				`${{ join (split repository.name "/") "-" }}`,
				`github.com-sourcegraph-src-cli`,
			},
			{
				`${{ split repository.name "/" "-" }}`,
				`{{split repository.name "/" "-"}}`,
			},
			{
				`${{ replace repository.name "github" "foobar" }}`,
				`foobar.com/sourcegraph/src-cli`,
			},
			{
				`${{ join_if "SEP" repository.name "postfix" }}`,
				`github.com/sourcegraph/src-cliSEPpostfix`,
			},
			{
				`${{ matches repository.name "github.com*" }}`,
				`true`,
			},
		} {
			runParseAndPartialTest(t, tt.input, tt.want)
		}
	})

	t.Run("aborted", func(t *testing.T) {
		for _, tt := range []struct{ input, want string }{
			{
				// Wrong argument type
				`${{ join "foobar" "-" }}`,
				`{{join "foobar" "-"}}`,
			},
			{
				// Wrong argument count
				`${{ join (split repository.name "/") "-" "too" "many" "args" }}`,
				`{{join (split repository.name "/") "-" "too" "many" "args"}}`,
			},
		} {
			runParseAndPartialTest(t, tt.input, tt.want)
		}
	})
}

func TestIsStaticBool(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		wantIsStatic bool
		wantBoolVal  bool
	}{

		{
			name:         "true literal",
			template:     `true`,
			wantIsStatic: true,
			wantBoolVal:  true,
		},
		{
			name:         "false literal",
			template:     `false`,
			wantIsStatic: true,
			wantBoolVal:  false,
		},
		{
			name:         "static non bool value",
			template:     `${{ repository.name }}`,
			wantIsStatic: true,
			wantBoolVal:  false,
		},
		{
			name:         "function call true val",
			template:     `${{ eq repository.name "github.com/sourcegraph/src-cli" }}`,
			wantIsStatic: true,
			wantBoolVal:  true,
		},
		{
			name:         "function call false val",
			template:     `${{ eq repository.name "hans wurst" }}`,
			wantIsStatic: true,
			wantBoolVal:  false,
		},
		{
			name:         "nested function call and whitespace",
			template:     `   ${{ eq false (eq false (matches repository.name "github.com*")) }}   `,
			wantIsStatic: true,
			wantBoolVal:  true,
		},
		{
			name:         "nested function call with runtime value",
			template:     `${{ eq false (eq false (matches outputs.repo.name "github.com*")) }}`,
			wantIsStatic: false,
			wantBoolVal:  false,
		},
		{
			name:         "random string",
			template:     `adfadsfasdfadsfasdfasdfadsf`,
			wantIsStatic: true,
			wantBoolVal:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			isStatic, boolVal, err := IsStaticBool(tt.template, partialEvalStepCtx)
			if err != nil {
				t.Fatal(err)
			}

			if isStatic != tt.wantIsStatic {
				t.Fatalf("wrong isStatic value. want=%t, got=%t", tt.wantIsStatic, isStatic)
			}
			if boolVal != tt.wantBoolVal {
				t.Fatalf("wrong boolVal value. want=%t, got=%t", tt.wantBoolVal, boolVal)
			}
		})
	}
}
