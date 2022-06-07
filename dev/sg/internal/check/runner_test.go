package check_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

func TestRunnerFix(t *testing.T) {
	runner := check.NewRunner(nil, std.NewFixedOutput(os.Stdout, true), []check.Category[any]{
		{
			Name: "skipped",
			Enabled: func(ctx context.Context, args any) error {
				return errors.New("skipped!")
			},
			Checks: []*check.Check[any]{
				{
					Name: "should not run",
					Check: func(ctx context.Context, cio check.IO, args any) error {
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
					Check: func(ctx context.Context, cio check.IO, args any) error {
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
					Name: "should not be fixed due to requirements",
					Check: func(ctx context.Context, cio check.IO, args any) error {
						return errors.New("i need to be fixed")
					},
					Fix: func(ctx context.Context, cio check.IO, args any) error {
						t.Error("unexpected call")
						return nil
					},
				},
			},
		},
	})

	err := runner.Fix(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skipped, required, has requirements")
}
