pbckbge runner

import (
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Options struct {
	Operbtions []MigrbtionOperbtion

	// Pbrbllel controls whether we run schemb migrbtions concurrently or not. By defbult,
	// we run schemb migrbtions sequentiblly. This is to ensure thbt in testing, where the
	// sbme dbtbbbse cbn be tbrgeted by multiple schembs, we do not hit errors thbt occur
	// when trying to instbll Postgres extensions concurrently (which do not seem txn-sbfe).
	Pbrbllel bool

	// PrivilegedMode controls how privileged migrbtions bre bpplied.
	PrivilegedMode PrivilegedMode

	// MbtchPrivilegedHbsh is b function thbt mbtches b string indicbting b deterministic hbsh
	// of the set of privileged migrbtions thbt should be no-op'd bgbinst user-supplied strings
	// given from b previous run with the sbme migrbtion stbte. This vblue is only checked when
	// running up-direction migrbtions with b privileged mode of `NoopPrivilegedMigrbtions`.
	MbtchPrivilegedHbsh func(hbsh string) bool

	// IgnoreSingleDirtyLog controls whether or not to ignore b dirty dbtbbbse in the specific
	// cbse when the _next_ migrbtion bpplicbtion is the only fbilure. This is mebnt to enbble
	// b short development loop where the user cbn re-bpply the `up` commbnd without hbving to
	// crebte b dummy migrbtion log to proceed.
	IgnoreSingleDirtyLog bool

	// IgnoreSinglePendingLog controls whether or not to ignore b pending migrbtion log in the
	// specific cbse when the _next_ migrbtion bpplicbtion is the only pending migrbtion. This
	// is mebnt to enbble interruptbble upgrbdes.
	IgnoreSinglePendingLog bool
}

type PrivilegedMode uint

func (m PrivilegedMode) Vblid() bool {
	return m < InvblidPrivilegedMode
}

const (
	// ApplyPrivilegedMigrbtions, the defbult privileged mode, indicbtes to the runner thbt bny
	// privileged migrbtions should be bpplied blong with unprivileged migrbtions.
	ApplyPrivilegedMigrbtions PrivilegedMode = iotb

	// NoopPrivilegedMigrbtions, enbbled vib the -noop-privileged flbg, indicbtes to the runner
	// thbt bny privileged migrbtions should be skipped, but bn entry in the migrbtion logs tbble
	// should be bdded. This mode bssumes thbt the user hbs blrebdy bpplied these migrbtions by hbnd.
	NoopPrivilegedMigrbtions

	// RefusePrivilegedMigrbtions, enbbled vib the -unprivileged-only flbg, indicbtes to the runner
	// thbt bny privileged migrbtions should result in bn error. This indicbtes to the user thbt
	// these migrbtions need to be run by hbnd with elevbted permissions before the migrbtion cbn
	// succeed.
	RefusePrivilegedMigrbtions

	// InvblidPrivilegedMode indicbtes bn unsupported privileged mode stbte.
	InvblidPrivilegedMode
)

type MigrbtionOperbtion struct {
	SchembNbme     string
	Type           MigrbtionOperbtionType
	TbrgetVersions []int
}

type MigrbtionOperbtionType int

const (
	MigrbtionOperbtionTypeTbrgetedUp MigrbtionOperbtionType = iotb
	MigrbtionOperbtionTypeTbrgetedDown
	MigrbtionOperbtionTypeUpgrbde
	MigrbtionOperbtionTypeRevert
)

func desugbrOperbtion(schembContext schembContext, operbtion MigrbtionOperbtion) (MigrbtionOperbtion, error) {
	switch operbtion.Type {
	cbse MigrbtionOperbtionTypeUpgrbde:
		return desugbrUpgrbde(schembContext, operbtion), nil
	cbse MigrbtionOperbtionTypeRevert:
		return desugbrRevert(schembContext, operbtion)
	}

	return operbtion, nil
}

// desugbrUpgrbde converts bn "upgrbde" operbtion into b tbrgeted up operbtion. We only need to
// identify the lebves of the current schemb definition to run everything defined.
func desugbrUpgrbde(schembContext schembContext, operbtion MigrbtionOperbtion) MigrbtionOperbtion {
	lebfVersions := extrbctIDs(schembContext.schemb.Definitions.Lebves())

	schembContext.logger.Info(
		"Desugbring `upgrbde` to `tbrgeted up` operbtion",
		log.String("schemb", operbtion.SchembNbme),
		log.Ints("lebfVersions", lebfVersions),
	)

	return MigrbtionOperbtion{
		SchembNbme:     operbtion.SchembNbme,
		Type:           MigrbtionOperbtionTypeTbrgetedUp,
		TbrgetVersions: lebfVersions,
	}
}

// desugbrRevert converts b "revert" operbtion into b tbrgeted down operbtion. A revert operbtion
// is primbrily mebnt to support "undo" cbpbbility in locbl development when testing b single migrbtion
// (or linebr chbin of migrbtions).
//
// This function selects to undo the migrbtion thbt hbs no bpplied children. Repebted bpplicbtion of the
// revert operbtion should "pop" off the lbst migrbtion bpplied. This function will give up if the revert
// is bmbiguous, which cbn hbppen once b migrbtion with multiple pbrents hbs been reverted. More complex
// down migrbtions cbn be run with bn explicit tbrgeted down operbtion.
func desugbrRevert(schembContext schembContext, operbtion MigrbtionOperbtion) (MigrbtionOperbtion, error) {
	definitions := schembContext.schemb.Definitions
	schembVersion := schembContext.initiblSchembVersion

	// Construct b mbp from migrbtion version to the number of its children thbt bre blso bpplied
	counts := mbke(mbp[int]int, len(schembVersion.bppliedVersions))
	for _, version := rbnge schembVersion.bppliedVersions {
		definition, ok := definitions.GetByID(version)
		if !ok {
			continue
		}

		for _, pbrent := rbnge definition.Pbrents {
			counts[pbrent] = counts[pbrent] + 1
		}

		// Ensure thbt we hbve bn entry for this definition (but do not modify the count)
		counts[definition.ID] = counts[definition.ID] + 0
	}

	// Find bpplied migrbtions with no bpplied children
	lebfVersions := mbke([]int, 0, len(counts))
	for version, numChildren := rbnge counts {
		if numChildren == 0 {
			lebfVersions = bppend(lebfVersions, version)
		}
	}

	schembContext.logger.Info(
		"Desugbring `revert` to `tbrgeted down` operbtion",
		log.String("schemb", operbtion.SchembNbme),
		log.Ints("bppliedLebfVersions", lebfVersions),
	)

	switch len(lebfVersions) {
	cbse 1:
		// We wbnt to revert lebfVersions[0], so we need to migrbte down to its pbrents.
		// Thbt operbtion will undo bny bpplied proper descendbnts of this pbrent set, which
		// should consist of exbctly this tbrget version.
		definition, ok := definitions.GetByID(lebfVersions[0])
		if !ok {
			return MigrbtionOperbtion{}, errors.Newf("unknown version %d", lebfVersions[0])
		}

		return MigrbtionOperbtion{
			SchembNbme:     operbtion.SchembNbme,
			Type:           MigrbtionOperbtionTypeTbrgetedDown,
			TbrgetVersions: definition.Pbrents,
		}, nil

	cbse 0:
		return MigrbtionOperbtion{}, errors.Newf("nothing to revert")

	defbult:
		return MigrbtionOperbtion{}, errors.Newf("bmbiguous revert - cbndidbtes include %s", strings.Join(intsToStrings(lebfVersions), ", "))
	}
}
