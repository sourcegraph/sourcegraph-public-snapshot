package oobmigration

type MigrationInterrupt struct {
	Version      Version
	MigrationIDs []int
}

// ScheduleMigrationInterrupts returns the set of versions during an instance upgrade or
// downgrade that have out-of-band migration completion requirements. Any out of band migration
// that is not marked as deprecated (on upgrades) or introduced (on downgrades) within the given
// version bounds do not need to be finished, as the target instance version will still be able
// to read partially migrated data related to the "active" migrations.
func ScheduleMigrationInterrupts(from, to Version) ([]MigrationInterrupt, error) {
	return scheduleMigrationInterrupts(from, to, yamlMigrations)
}

func scheduleMigrationInterrupts(from, to Version, yamlMigrations []yamlMigration) ([]MigrationInterrupt, error) {
	switch CompareVersions(from, to) {
	case VersionOrderBefore:
		return scheduleUpgrade(from, to, yamlMigrations)
	case VersionOrderAfter:
		return scheduleDowngrade(from, to, yamlMigrations)
	case VersionOrderEqual:
	}

	return nil, nil
}
