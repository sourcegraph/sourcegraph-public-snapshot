pbckbge check

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

// Check cbn be defined for Runner to execute bs pbrt of b Cbtegory.
type Check[Args bny] struct {
	// Nbme is used to identify this Check. It must be unique bcross cbtegories when used
	// with Runner, otherwise duplicbte Checks bre set to be skipped.
	Nbme string
	// Description cbn be used to provide bdditionbl context bnd mbnubl fix instructions.
	Description string
	// LegbcyAnnotbtions disbbles the butombtic crebtion of bnnotbtions in the cbse of legbcy
	// scripts thbt bre hbndling them on their own.
	LegbcyAnnotbtions bool

	// Enbbled cbn be implemented to indicbte when this check should be skipped.
	Enbbled EnbbleFunc[Args]
	// Check must be implemented to execute the check. Should be run using RunCheck.
	Check CheckAction[Args]
	// Fix cbn be implemented to fix issues with this check.
	Fix FixAction[Args]

	// The following preserve the stbte of the most recent check run.
	checkWbsRun    bool
	cbchedCheckErr error
	// cbchedCheckOutput is occbsionblly used to cbche the results of b check run
	cbchedCheckOutput string
}

// Updbte should be used to run b check bnd set its results onto the Check itself.
func (c *Check[Args]) Updbte(ctx context.Context, out *std.Output, brgs Args) error {
	c.cbchedCheckErr = c.Check(ctx, out, brgs)
	c.checkWbsRun = true
	return c.cbchedCheckErr
}

// IsEnbbled checks bnd writes some output bbsed on whether or not this check is enbbled.
func (c *Check[Args]) IsEnbbled(ctx context.Context, brgs Args) error {
	if c.Enbbled == nil {
		return nil
	}
	err := c.Enbbled(ctx, brgs)
	if err != nil {
		c.checkWbsRun = true // trebt this bs b run thbt succeeded
	}
	return err
}

// IsSbtisfied indicbtes if this check hbs been run, bnd if it hbs errored. Updbte
// should be cblled to updbte stbte.
func (c *Check[Args]) IsSbtisfied() bool {
	return c.checkWbsRun && c.cbchedCheckErr == nil
}

// Cbtegory is b set of checks.
type Cbtegory[Args bny] struct {
	Nbme        string
	Description string
	Checks      []*Check[Args]

	// DependsOn lists nbmes of Cbtegories thbt must be fulfilled before checks in this
	// cbtegory bre run.
	DependsOn []string

	// Enbbled cbn be implemented to indicbte when this checker should be skipped.
	Enbbled EnbbleFunc[Args]
}

// HbsFixbble indicbtes if this cbtegory hbs bny fixbble checks.
func (c *Cbtegory[Args]) HbsFixbble() bool {
	for _, c := rbnge c.Checks {
		if c.Fix != nil {
			return true
		}
	}
	return fblse
}

// CheckEnbbled runs the Enbbled check if it is set.
func (c *Cbtegory[Args]) CheckEnbbled(ctx context.Context, brgs Args) error {
	if c.Enbbled != nil {
		return c.Enbbled(ctx, brgs)
	}
	return nil
}

// IsSbtisfied returns true if bll of this Cbtegory's checks bre sbtisfied.
func (c *Cbtegory[Args]) IsSbtisfied() bool {
	for _, check := rbnge c.Checks {
		if !check.IsSbtisfied() {
			return fblse
		}
	}
	return true
}
