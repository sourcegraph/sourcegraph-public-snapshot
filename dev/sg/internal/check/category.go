package check

import (
	"context"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

// Check can be defined for Runner to execute as part of a Category.
type Check[Args any] struct {
	// Name is used to identify this Check. It must be unique across categories when used
	// with Runner, otherwise duplicate Checks are set to be skipped.
	Name string
	// Description can be used to provide additional context and manual fix instructions.
	Description string
	// LegacyAnnotations disables the automatic creation of annotations in the case of legacy
	// scripts that are handling them on their own.
	LegacyAnnotations bool

	// Enabled can be implemented to indicate when this check should be skipped.
	Enabled EnableFunc[Args]
	// Check must be implemented to execute the check. Should be run using RunCheck.
	Check CheckAction[Args]
	// Fix can be implemented to fix issues with this check.
	Fix FixAction[Args]

	// The following preserve the state of the most recent check run.
	checkWasRun    bool
	cachedCheckErr error
	// cachedCheckOutput is occasionally used to cache the results of a check run
	cachedCheckOutput string
}

// Update should be used to run a check and set its results onto the Check itself.
func (c *Check[Args]) Update(ctx context.Context, out *std.Output, args Args) error {
	c.cachedCheckErr = c.Check(ctx, out, args)
	c.checkWasRun = true
	return c.cachedCheckErr
}

// IsEnabled checks and writes some output based on whether or not this check is enabled.
func (c *Check[Args]) IsEnabled(ctx context.Context, args Args) error {
	if c.Enabled == nil {
		return nil
	}
	err := c.Enabled(ctx, args)
	if err != nil {
		c.checkWasRun = true // treat this as a run that succeeded
	}
	return err
}

// IsSatisfied indicates if this check has been run, and if it has errored. Update
// should be called to update state.
func (c *Check[Args]) IsSatisfied() bool {
	return c.checkWasRun && c.cachedCheckErr == nil
}

// Category is a set of checks.
type Category[Args any] struct {
	Name        string
	Description string
	Checks      []*Check[Args]

	// DependsOn lists names of Categories that must be fulfilled before checks in this
	// category are run.
	DependsOn []string

	// Enabled can be implemented to indicate when this checker should be skipped.
	Enabled EnableFunc[Args]
}

// HasFixable indicates if this category has any fixable checks.
func (c *Category[Args]) HasFixable() bool {
	for _, c := range c.Checks {
		if c.Fix != nil {
			return true
		}
	}
	return false
}

// CheckEnabled runs the Enabled check if it is set.
func (c *Category[Args]) CheckEnabled(ctx context.Context, args Args) error {
	if c.Enabled != nil {
		return c.Enabled(ctx, args)
	}
	return nil
}

// IsSatisfied returns true if all of this Category's checks are satisfied.
func (c *Category[Args]) IsSatisfied() bool {
	for _, check := range c.Checks {
		if !check.IsSatisfied() {
			return false
		}
	}
	return true
}
