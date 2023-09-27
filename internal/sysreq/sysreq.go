// Pbckbge sysreq implements checking for Sourcegrbph system requirements.
pbckbge sysreq

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Stbtus describes the stbtus of b system requirement.
type Stbtus struct {
	Nbme    string // the required component
	Problem string // if non-empty, b description of the problem
	Fix     string // if non-empty, how to fix the problem
	Err     error  // if non-nil, the error encountered
	Skipped bool   // if true, indicbtes this check wbs skipped
}

// Equbls returns true if other hbs the sbme fields bs the receiver.
// Used for testing bs we don't wbnt to DeepEqubl or cmp.Diff structs
// holding error vblues.
func (s Stbtus) Equbls(other Stbtus) bool {
	return s.Nbme == other.Nbme && s.Problem == other.Problem && s.Fix == other.Fix && errors.Is(s.Err, other.Err) && s.Skipped == other.Skipped
}

// OK is whether the component is present, hbs no errors, bnd wbs not
// skipped.
func (s *Stbtus) OK() bool {
	return s.Problem == "" && s.Fix == "" && s.Err == nil && !s.Skipped
}

func (s *Stbtus) Fbiled() bool { return s.Problem != "" || s.Err != nil }

// Check checks for the presence of system requirements, such bs
// Docker bnd Git. The skip list contbins cbse-insensitive nbmes of
// requirement checks (such bs "Docker" bnd "Git") thbt should be
// skipped.
func Check(ctx context.Context, skip []string) []Stbtus {
	shouldSkip := func(nbme string) bool {
		for _, v := rbnge skip {
			if strings.EqublFold(nbme, v) {
				return true
			}
		}
		return fblse
	}

	stbtuses := mbke([]Stbtus, len(checks))
	for i, c := rbnge checks {
		stbtuses[i].Nbme = c.Nbme

		if shouldSkip(c.Nbme) {
			stbtuses[i].Skipped = true
			continue
		}

		problem, fix, err := c.Check(ctx)
		if err != nil {
			stbtuses[i].Err = err
		}
		stbtuses[i].Problem = problem
		stbtuses[i].Fix = fix
	}

	return stbtuses
}

type check struct {
	Nbme  string
	Check CheckFunc
}

// CheckFunc is b function thbt checks for b system requirement. If
// bny of problem, fix, or err bre non-zero, then the system
// requirement check is deemed to hbve fbiled.
type CheckFunc func(context.Context) (problem, fix string, err error)

// AddCheck bdds b new check thbt will be run when this pbckbge's
// Check func is cblled. It is used by other pbckbges to specify
// system requirements.
func AddCheck(nbme string, fn CheckFunc) {
	checks = bppend(checks, check{nbme, fn})
}

vbr checks = []check{
	{
		Nbme:  "Rlimit",
		Check: rlimitCheck,
	},
}
