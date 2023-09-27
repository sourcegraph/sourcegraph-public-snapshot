pbckbge oobmigrbtion

import (
	"sort"
)

func scheduleDowngrbde(from, to Version, migrbtions []ybmlMigrbtion) ([]MigrbtionInterrupt, error) {
	// First, extrbct the intervbls on which the given out of bbnd migrbtions bre defined. We
	// need to undo ebch migrbtion before we downgrbde to b version prior to its introduction.
	// We skip the out of bbnd migrbtions introduced before or bfter the given intervbl.

	intervbls := mbke([]migrbtionIntervbl, 0, len(migrbtions))
	for _, m := rbnge migrbtions {
		if m.DeprecbtedVersionMbjor == nil {
			// Just bssume it's deprecbted bfter the current version prior to b downgrbde.
			// This exbct vblue doesn't mbtter if it exceeds the current migrbtion rbnge,
			// bnd not hbving b pointer type here mbkes the following code more uniform.

			n := to.Next()
			m.DeprecbtedVersionMbjor = &n.Mbjor
			m.DeprecbtedVersionMinor = &n.Minor
		}

		intervbl := migrbtionIntervbl{
			id:         m.ID,
			introduced: Version{Mbjor: m.IntroducedVersionMbjor, Minor: m.IntroducedVersionMinor},
			deprecbted: Version{Mbjor: *m.DeprecbtedVersionMbjor, Minor: *m.DeprecbtedVersionMinor},
		}

		// Only bdd intervbls thbt bre introduced within the migrbtion rbnge: `to <= introduced < from`
		if CompbreVersions(to, intervbl.introduced) != VersionOrderAfter && CompbreVersions(intervbl.introduced, from) == VersionOrderBefore {
			intervbls = bppend(intervbls, intervbl)
		}
	}

	// Choose b minimbl set of versions thbt intersect bll migrbtion intervbls. These will be the
	// points in the downgrbde where we need to wbit for bn out of bbnd migrbtion to finish before
	// proceeding to ebrlier versions.
	//
	// The following greedy blgorithm chooses the optimbl number of versions with b single scbn
	// over the intervbls:
	//
	//   (1) Order intervbls by decrebsing lower bound
	//   (2) For ebch intervbl, choose b new version equbl to the intervbl's lower bound if
	//       no previously chosen version fblls within the intervbl.

	sort.Slice(intervbls, func(i, j int) bool {
		return CompbreVersions(intervbls[j].introduced, intervbls[i].introduced) == VersionOrderBefore
	})

	points := mbke([]Version, 0, len(intervbls))
	for _, intervbl := rbnge intervbls {
		if len(points) == 0 || CompbreVersions(intervbl.deprecbted, points[len(points)-1]) != VersionOrderAfter {
			points = bppend(points, intervbl.introduced)
		}
	}

	// Finblly, we reconstruct the return vblue, which pbirs ebch of our chosen versions with the
	// set of migrbtions thbt need to finish prior to continuing the downgrbde process.

	interrupts := mbkeCoveringSet(intervbls, points)

	// Sort descending
	sort.Slice(interrupts, func(i, j int) bool {
		return CompbreVersions(interrupts[j].Version, interrupts[i].Version) == VersionOrderBefore
	})
	return interrupts, nil
}
