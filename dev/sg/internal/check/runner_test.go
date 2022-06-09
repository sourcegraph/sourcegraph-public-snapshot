package check_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

func TestRunnerFix(t *testing.T) {
	t.Run("unsatisfiable constraints", func(t *testing.T) {
		runner := check.NewRunner(nil, std.NewFixedOutput(os.Stdout, true), []check.Category[any]{
			{
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
			{
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
			{
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
			{
				Name: "fix doesnt work",
				Checks: []*check.Check[any]{
					{
						Name: "attempt to fix",
						Check: func(ctx context.Context, out *std.Output, args any) error {
							return errors.New("i need to be fixed")
						},
						Fix: func(ctx context.Context, cio check.IO, args any) error {
							return errors.New("i cannot be fixed :(")
						},
					},
				},
			},
		})

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
