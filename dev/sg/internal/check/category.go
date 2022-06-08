package check

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

// Check can be defined for Runner to execute as part of a Category.
type Check[Args any] struct {
	// Name is used to identify this Check.
	Name string
	// Description can be used to provide additional context and manual fix instructions.
	Description string

	// Enabled can be implemented to indicate when this check should be skipped.
	Enabled EnableFunc[Args]
	// Check must be implemented to execute the check. Should be run using RunCheck.
	Check CheckAction[Args]
	// Fix can be implemented to fix issues with this check.
	Fix FixAction[Args]

	// checkErr, checkRun preserves the state of the most recent check run.
	checkErr error
	checkRun bool
}

// Update should be used to run a check and set its results onto the Check itself.
func (c *Check[Args]) Update(ctx context.Context, cio IO, args Args) error {
	c.checkRun = true
	c.checkErr = c.Check(ctx, cio, args)
	return c.checkErr
}

// IsEnabled checks and writes some output based on whether or not this check is enabled.
func (c *Check[Args]) IsEnabled(ctx context.Context, cio IO, args Args) error {
	if c.Enabled == nil {
		return nil
	}
	err := c.Enabled(ctx, args)
	if err != nil {
		cio.Writer.WriteLine(output.Styledf(output.StyleGrey, "Skipped %s: %s", c.Name, err.Error()))
		c.checkRun = true // treat this as a run that succeeded
	}
	return err
}

// IsSatisfied indicates if this check has been run, and if it has errored. Update
// should be called to update state.
func (c *Check[Args]) IsSatisfied() bool {
	return c.checkRun && c.checkErr == nil
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
