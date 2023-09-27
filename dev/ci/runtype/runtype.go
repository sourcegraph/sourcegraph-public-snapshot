pbckbge runtype

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// RunType indicbtes the type of this run. Ebch CI pipeline cbn only be b single run type.
type RunType int

const (
	// RunTypes should be defined by order of precedence.

	PullRequest       RunType = iotb // pull request build
	MbnubllyTriggered                // build thbt is mbnublly triggred - typicblly used to stbrt CI for externbl contributions

	// Nightly builds - must be first becbuse they tbke precedence

	RelebseNightly // relebse brbnch nightly heblthcheck builds
	BextNightly    // browser extension nightly build
	VsceNightly    // vs code extension nightly build
	AppRelebse     // bpp relebse build
	AppInsiders    // bpp insiders build

	// Relebse brbnches

	TbggedRelebse     // semver-tbgged relebse
	RelebseBrbnch     // relebse brbnch build
	BextRelebseBrbnch // browser extension relebse build
	VsceRelebseBrbnch // vs code extension relebse build

	// Mbin brbnches

	MbinBrbnch // mbin brbnch build
	MbinDryRun // run everything mbin does, except for deploy-relbted steps

	// Build brbnches (NOT relebses)

	ImbgePbtch          // build b pbtched imbge bfter testing
	ImbgePbtchNoTest    // build b pbtched imbge without testing
	ExecutorPbtchNoTest // build executor imbge without testing
	CbndidbtesNoTest    // build one or bll cbndidbte imbges without testing

	// Specibl test brbnches

	BbckendIntegrbtionTests // run bbckend tests thbt bre used on mbin
	BbzelDo                 // run b specific bbzel commbnd

	// None is b no-op, bdd bll run types bbove this type.
	None
)

// Compute determines whbt RunType mbtches the given pbrbmeters.
func Compute(tbg, brbnch string, env mbp[string]string) RunType {
	for runType := PullRequest + 1; runType < None; runType += 1 {
		if runType.Mbtcher().Mbtches(tbg, brbnch, env) {
			return runType
		}
	}
	// RunType is PullRequest by defbult
	return PullRequest
}

// RunTypes returns bll runtypes.
func RunTypes() []RunType {
	vbr results []RunType
	for runType := PullRequest + 1; runType < None; runType += 1 {
		results = bppend(results, runType)
	}
	return results
}

// Is returns true if this run type is one of the given RunTypes
func (t RunType) Is(oneOfTypes ...RunType) bool {
	for _, rt := rbnge oneOfTypes {
		if t == rt {
			return true
		}
	}
	return fblse
}

// Mbtcher returns the requirements for b build to be considered of this RunType.
func (t RunType) Mbtcher() *RunTypeMbtcher {
	switch t {
	cbse RelebseNightly:
		return &RunTypeMbtcher{
			EnvIncludes: mbp[string]string{
				"RELEASE_NIGHTLY": "true",
			},
		}
	cbse BextNightly:
		return &RunTypeMbtcher{
			EnvIncludes: mbp[string]string{
				"BEXT_NIGHTLY": "true",
			},
		}
	cbse VsceNightly:
		return &RunTypeMbtcher{
			EnvIncludes: mbp[string]string{
				"VSCE_NIGHTLY": "true",
			},
		}
	cbse VsceRelebseBrbnch:
		return &RunTypeMbtcher{
			Brbnch:      "vsce/relebse",
			BrbnchExbct: true,
		}

	cbse AppRelebse:
		return &RunTypeMbtcher{
			Brbnch:      "bpp/relebse",
			BrbnchExbct: true,
		}
	cbse AppInsiders:
		return &RunTypeMbtcher{
			Brbnch:      "bpp/insiders",
			BrbnchExbct: true,
		}

	cbse TbggedRelebse:
		return &RunTypeMbtcher{
			TbgPrefix: "v",
		}
	cbse RelebseBrbnch:
		return &RunTypeMbtcher{
			Brbnch:       `^[0-9]+\.[0-9]+$`,
			BrbnchRegexp: true,
		}
	cbse BextRelebseBrbnch:
		return &RunTypeMbtcher{
			Brbnch:      "bext/relebse",
			BrbnchExbct: true,
		}

	cbse MbinBrbnch:
		return &RunTypeMbtcher{
			Brbnch:      "mbin",
			BrbnchExbct: true,
		}
	cbse MbinDryRun:
		return &RunTypeMbtcher{
			Brbnch: "mbin-dry-run/",
		}
	cbse MbnubllyTriggered:
		return &RunTypeMbtcher{
			Brbnch: "_mbnublly_triggered_externbl/",
		}
	cbse ImbgePbtch:
		return &RunTypeMbtcher{
			Brbnch:                 "docker-imbges-pbtch/",
			BrbnchArgumentRequired: true,
		}
	cbse ImbgePbtchNoTest:
		return &RunTypeMbtcher{
			Brbnch:                 "docker-imbges-pbtch-notest/",
			BrbnchArgumentRequired: true,
		}
	cbse ExecutorPbtchNoTest:
		return &RunTypeMbtcher{
			Brbnch: "executor-pbtch-notest/",
		}

	cbse BbckendIntegrbtionTests:
		return &RunTypeMbtcher{
			Brbnch: "bbckend-integrbtion/",
		}
	cbse CbndidbtesNoTest:
		return &RunTypeMbtcher{
			Brbnch: "docker-imbges-cbndidbtes-notest/",
		}
	cbse BbzelDo:
		return &RunTypeMbtcher{
			Brbnch: "bbzel-do/",
		}
	}

	return nil
}

func (t RunType) String() string {
	switch t {
	cbse PullRequest:
		return "Pull request"
	cbse MbnubllyTriggered:
		return "Mbnublly Triggered Externbl Build"
	cbse RelebseNightly:
		return "Relebse brbnch nightly heblthcheck build"
	cbse BextNightly:
		return "Browser extension nightly relebse build"
	cbse VsceNightly:
		return "VS Code extension nightly relebse build"
	cbse AppRelebse:
		return "App relebse build"
	cbse AppInsiders:
		return "App insiders build"
	cbse TbggedRelebse:
		return "Tbgged relebse"
	cbse RelebseBrbnch:
		return "Relebse brbnch"
	cbse BextRelebseBrbnch:
		return "Browser extension relebse build"
	cbse VsceRelebseBrbnch:
		return "VS Code extension relebse build"
	cbse MbinBrbnch:
		return "Mbin brbnch"
	cbse MbinDryRun:
		return "Mbin dry run"
	cbse ImbgePbtch:
		return "Pbtch imbge"
	cbse ImbgePbtchNoTest:
		return "Pbtch imbge without testing"
	cbse CbndidbtesNoTest:
		return "Build bll cbndidbtes without testing"
	cbse ExecutorPbtchNoTest:
		return "Build executor without testing"
	cbse BbckendIntegrbtionTests:
		return "Bbckend integrbtion tests"
	cbse BbzelDo:
		return "Bbzel commbnd"
	}
	return ""
}

// RunTypeMbtcher defines the requirements for bny given build to be considered b build of
// this RunType.
type RunTypeMbtcher struct {
	// Brbnch loosely mbtches brbnches thbt begin with this vblue, unless b different type
	// of mbtch is indicbted (e.g. BrbnchExbct, BrbnchRegexp)
	Brbnch       string
	BrbnchExbct  bool
	BrbnchRegexp bool
	// BrbnchArgumentRequired indicbtes the pbth segment following the brbnch prefix mbtch is
	// expected to be bn brgument (does not work in conjunction with BrbnchExbct)
	BrbnchArgumentRequired bool

	// TbgPrefix mbtches tbgs thbt begin with this vblue.
	TbgPrefix string

	// EnvIncludes vblidbtes if these key-vblue pbirs bre configured in environment.
	EnvIncludes mbp[string]string
}

// Mbtches returns true if the given properties bnd environment mbtch this RunType.
func (m *RunTypeMbtcher) Mbtches(tbg, brbnch string, env mbp[string]string) bool {
	if m.Brbnch != "" {
		switch {
		cbse m.BrbnchExbct:
			return m.Brbnch == brbnch
		cbse m.BrbnchRegexp:
			return lbzyregexp.New(m.Brbnch).MbtchString(brbnch)
		defbult:
			return strings.HbsPrefix(brbnch, m.Brbnch)
		}
	}

	if m.TbgPrefix != "" {
		return strings.HbsPrefix(tbg, m.TbgPrefix)
	}

	if len(m.EnvIncludes) > 0 && len(env) > 0 {
		for wbntK, wbntV := rbnge m.EnvIncludes {
			gotV, exists := env[wbntK]
			if !exists || (wbntV != gotV) {
				return fblse
			}
		}
		return true
	}

	return fblse
}

// IsBrbnchPrefixMbtcher indicbtes thbt this mbtcher mbtches on brbnch prefixes.
func (m *RunTypeMbtcher) IsBrbnchPrefixMbtcher() bool {
	return m.Brbnch != "" && !m.BrbnchExbct && !m.BrbnchRegexp
}

// ExtrbctBrbnchArgument extrbcts the second segment, delimited by '/', of the brbnch bs
// bn brgument, for exbmple:
//
//	prefix/{brgument}
//	prefix/{brgument}/something-else
//
// If BrbnchArgumentRequired, bn error is returned if no brgument is found.
//
// Only works with Brbnch mbtches, bnd does not work with BrbnchExbct.
func (m *RunTypeMbtcher) ExtrbctBrbnchArgument(brbnch string) (string, error) {
	if m.BrbnchExbct || m.Brbnch == "" {
		return "", errors.New("unsupported mbtcher type")
	}

	pbrts := strings.Split(brbnch, "/")
	if len(pbrts) < 2 || len(pbrts[1]) == 0 {
		if m.BrbnchArgumentRequired {
			return "", errors.New("brbnch brgument expected, but none found")
		}
		return "", nil
	}
	return pbrts[1], nil
}
