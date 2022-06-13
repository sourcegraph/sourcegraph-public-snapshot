package check_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

func getUnsatisfiableChecks(t *testing.T) []check.Category[any] {
	return []check.Category[any]{
		{ // 1
			Name: "skipped",
			Enabled: func(ctx context.Context, args any) error {
				return errors.New("skipped!")
			},
			Checks: []*check.Check[any]{
				{
					Name: "should not run",
					Check: func(ctx context.Context, out *std.Output, args any) error {
						t.Error("unexpected call")
						return nil
					},
				},
			},
		},
		{ // 2
			Name: "required",
			Checks: []*check.Check[any]{
				{
					Name: "not satisfied",
					Check: func(ctx context.Context, out *std.Output, args any) error {
						return errors.New("check not satisfied")
					},
				},
			},
		},
		{ // 3
			Name:      "has requirements",
			DependsOn: []string{"required"},
			Checks: []*check.Check[any]{
				{
					Name: "should not be fixed due to requirements that cannot be satisfied",
					Check: func(ctx context.Context, out *std.Output, args any) error {
						return errors.New("i need to be fixed")
					},
					Fix: func(ctx context.Context, cio check.IO, args any) error {
						t.Error("unexpected call")
						return nil
					},
				},
			},
		},
		{ // 4
			Name: "fix doesnt work",
			Checks: []*check.Check[any]{
				{
					Name:        "attempt to fix",
					Description: "how to fix manually",
					Check: func(ctx context.Context, out *std.Output, args any) error {
						return errors.New("i need to be fixed")
					},
					Fix: func(ctx context.Context, cio check.IO, args any) error {
						return errors.New("4 cannot be fixed :(")
					},
				},
			},
		},
	}
}

func TestRunnerCheck(t *testing.T) {
	t.Run("unfixed checks", func(t *testing.T) {
		runner := check.NewRunner(nil, std.NewFixedOutput(os.Stdout, true), getUnsatisfiableChecks(t))

		err := runner.Check(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "3 checks failed (1 skipped)")
	})

	t.Run("okay checks", func(t *testing.T) {
		runner := check.NewRunner(nil, std.NewFixedOutput(os.Stdout, true), []check.Category[any]{
			{
				Name: "I'm okay!",
				Checks: []*check.Check[any]{
					{
						Name:  "OKAY",
						Check: func(ctx context.Context, out *std.Output, args any) error { return nil },
					},
				},
			},
		})

		err := runner.Check(context.Background(), nil)
		assert.NoError(t, err)
	})
}

func TestRunnerFix(t *testing.T) {
	t.Run("unsatisfiable constraints", func(t *testing.T) {
		runner := check.NewRunner(nil, std.NewFixedOutput(os.Stdout, true), getUnsatisfiableChecks(t))

		err := runner.Fix(context.Background(), nil)
		require.Error(t, err)
		for _, c := range []string{
			"Some categories are still unsatisfied",
			// Categories that should be failing
			"required",
			"has requirements",
		} {
			assert.Contains(t, err.Error(), c)
		}
	})

	t.Run("fix all in order", func(t *testing.T) {
		var fixedMap sync.Map
		runner := check.NewRunner(nil, std.NewFixedOutput(os.Stdout, true), []check.Category[any]{
			{
				Name: "broken but can be fixed",
				Checks: []*check.Check[any]{
					{
						Name: "fixable",
						Check: func(ctx context.Context, out *std.Output, args any) error {
							if _, ok := fixedMap.Load("1"); ok {
								return nil
							}
							return errors.New("needs fixing!")
						},
						Fix: func(ctx context.Context, cio check.IO, args any) error {
							fixedMap.Store("1", true)
							return nil
						},
					},
				},
			},
			{
				Name:      "depends on fixable",
				DependsOn: []string{"broken but can be fixed"},
				Checks: []*check.Check[any]{
					{
						Name: "also fixable",
						Check: func(ctx context.Context, out *std.Output, args any) error {
							if _, ok := fixedMap.Load("2"); ok {
								return nil
							}
							return errors.New("needs fixing!")
						},
						Fix: func(ctx context.Context, cio check.IO, args any) error {
							fixedMap.Store("2", true)
							return nil
						},
					},
					{
						Name: "no action needed",
						Check: func(ctx context.Context, out *std.Output, args any) error {
							return nil
						},
					},
					{
						Name: "disabled",
						Enabled: func(ctx context.Context, args any) error {
							return errors.New("disabled")
						},
					},
				},
			},
		})

		err := runner.Fix(context.Background(), nil)
		assert.NoError(t, err)
	})
}

func TestRunnerInteractive(t *testing.T) {
	t.Run("bad input", func(t *testing.T) {
		inputs := []string{
			"1",  // skipped, so unfixable
			"12", // not an option
		}
		var output strings.Builder
		runner := check.NewRunner(
			strings.NewReader(strings.Join(inputs, "\n")),
			std.NewSimpleOutput(&output, true),
			getUnsatisfiableChecks(t))

		runner.Interactive(context.Background(), nil)

		got := output.String()
		for _, c := range []string{
			"Some checks failed. Which one do you want to fix?",
			// Our choice was invalid
			"1 is an invalid choice :(",
			// Our second choice was invalid
			"12 is an invalid choice :(",
		} {
			assert.Contains(t, got, c)
		}
	})

	t.Run("auto fix", func(t *testing.T) {
		inputs := []string{
			"4", // fixable
			"3", // go back
			"4", // fixable
			"1", // automatically fix this for me
		}
		var output strings.Builder
		runner := check.NewRunner(
			strings.NewReader(strings.Join(inputs, "\n")),
			std.NewSimpleOutput(&output, true),
			getUnsatisfiableChecks(t))

		runner.Interactive(context.Background(), nil)

		got := output.String()
		for _, c := range []string{
			"Some checks failed. Which one do you want to fix?",
			// Unfixable error
			"4 cannot be fixed",
		} {
			assert.Contains(t, got, c)
		}
	})

	t.Run("manual fix", func(t *testing.T) {
		inputs := []string{
			"4", // fixable
			"2", // manual fix
			"1", // fix the first
			// "",  // say I'm done
			// "3",
		}
		var output strings.Builder
		runner := check.NewRunner(
			strings.NewReader(strings.Join(inputs, "\n")),
			std.NewSimpleOutput(&output, true),
			getUnsatisfiableChecks(t))

		// Fix did not work
		err := runner.Interactive(context.Background(), nil)
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "i need to be fixed")

		got := output.String()
		for _, c := range []string{
			"Some checks failed. Which one do you want to fix?",
			// description
			"how to fix manually",
		} {
			assert.Contains(t, got, c)
		}
	})
}
