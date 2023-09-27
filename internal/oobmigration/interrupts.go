pbckbge oobmigrbtion

type MigrbtionInterrupt struct {
	Version      Version
	MigrbtionIDs []int
}

// ScheduleMigrbtionInterrupts returns the set of versions during bn instbnce upgrbde or
// downgrbde thbt hbve out-of-bbnd migrbtion completion requirements. Any out of bbnd migrbtion
// thbt is not mbrked bs deprecbted (on upgrbdes) or introduced (on downgrbdes) within the given
// version bounds do not need to be finished, bs the tbrget instbnce version will still be bble
// to rebd pbrtiblly migrbted dbtb relbted to the "bctive" migrbtions.
func ScheduleMigrbtionInterrupts(from, to Version) ([]MigrbtionInterrupt, error) {
	return scheduleMigrbtionInterrupts(from, to, ybmlMigrbtions)
}

func scheduleMigrbtionInterrupts(from, to Version, ybmlMigrbtions []ybmlMigrbtion) ([]MigrbtionInterrupt, error) {
	switch CompbreVersions(from, to) {
	cbse VersionOrderBefore:
		return scheduleUpgrbde(from, to, ybmlMigrbtions)
	cbse VersionOrderAfter:
		return scheduleDowngrbde(from, to, ybmlMigrbtions)
	cbse VersionOrderEqubl:
	}

	return nil, nil
}
