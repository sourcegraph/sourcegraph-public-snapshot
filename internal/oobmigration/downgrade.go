package oobmigration

import (
	"sort"
)

func scheduleDowngrade(from, to Version, migrations []yamlMigration) ([]MigrationInterrupt, error) {
	// First, extract the intervals on which the given out of band migrations are defined. We
	// need to undo each migration before we downgrade to a version prior to its introduction.
	// We skip the out of band migrations introduced before or after the given interval.

	intervals := make([]migrationInterval, 0, len(migrations))
	for _, m := range migrations {
		introduced := Version{m.IntroducedVersionMajor, m.IntroducedVersionMinor}
		if CompareVersions(to, introduced) == VersionOrderAfter || CompareVersions(introduced, from) != VersionOrderBefore {
			// Skip migrations not introduced within the the instance downgrade interval
			continue
		}

		interval := migrationInterval{
			id:         m.ID,
			introduced: introduced,

			// Just assume it's deprecated after the current version prior to a downgrade.
			// This value doesn't matter, but not having a pointer type here makes the
			// following code a bit more uniform.
			deprecated: to.Next(),
		}
		if m.DeprecatedVersionMajor != nil {
			interval.deprecated = Version{*m.DeprecatedVersionMajor, *m.DeprecatedVersionMinor}
		}

		intervals = append(intervals, interval)
	}

	// Choose a minimal set of versions that intersect all migration intervals. These will be the
	// points in the downgrade where we need to wait for an out of band migration to finish before
	// proceeding to earlier versions.
	//
	// The following greedy algorithm chooses the optimal number of versions with a single scan
	// over the intervals:
	//
	//   (1) Order intervals by decreasing lower bound
	//   (2) For each interval, choose a new version equal to the interval's lower bound if
	//       no previously chosen version falls within the interval.

	sort.Slice(intervals, func(i, j int) bool {
		return CompareVersions(intervals[j].introduced, intervals[i].introduced) == VersionOrderBefore
	})

	points := make([]Version, 0, len(intervals))
	for _, interval := range intervals {
		if len(points) == 0 || CompareVersions(interval.deprecated, points[len(points)-1]) != VersionOrderAfter {
			points = append(points, interval.introduced)
		}
	}

	// Finally, we reconstruct the return value, which pairs each of our chosen versions with the
	// set of migrations that need to finish prior to continuing the downgrade process.

	interrupts, err := makeCoveringSet(intervals, points)
	if err != nil {
		return nil, err
	}

	// Sort descending
	sort.Slice(interrupts, func(i, j int) bool {
		return CompareVersions(interrupts[j].Version, interrupts[i].Version) == VersionOrderBefore
	})
	return interrupts, nil
}
