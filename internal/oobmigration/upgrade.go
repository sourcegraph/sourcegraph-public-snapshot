package oobmigration

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MigrationInterrupt struct {
	Version      Version
	MigrationIDs []int
}

// ScheduleMigrationInterrupts returns the set of versions during an instance upgrade that
// have out-of-band migration completion requirements. Any out of band migrations that do
// not become deprecated within the given version bounds do not need to be completed, as
// the target instance version will still be able to read partially migrated data from
// non (or not-yet-)-deprecated out of band migrations.
func ScheduleMigrationInterrupts(from, to Version) ([]MigrationInterrupt, error) {
	return scheduleMigrationInterrupts(from, to, yamlMigrations)
}

func scheduleMigrationInterrupts(from, to Version, migrations []yamlMigration) ([]MigrationInterrupt, error) {
	type migrationInterval struct {
		id         int
		introduced Version
		deprecated Version
	}

	// First, extract the intervals on which the given out of band migrations are defined. If
	// the interval hasn't been deprecated, it's still "open" and does not need to complete for
	// the instance upgrade operation to be successful.

	intervals := make([]migrationInterval, 0, len(migrations))
	for _, m := range migrations {
		if m.DeprecatedVersionMajor == nil {
			continue
		}

		introduced := Version{m.IntroducedVersionMajor, m.IntroducedVersionMinor}
		if CompareVersions(introduced, to) == VersionOrderAfter {
			// Skip migrations introduced after the target instance version
			continue
		}

		deprecated := Version{*m.DeprecatedVersionMajor, *m.DeprecatedVersionMinor}
		if !(CompareVersions(from, deprecated) == VersionOrderBefore && CompareVersions(deprecated, to) != VersionOrderAfter) {
			// Skip migrations not deprecated within the the instance upgrade interval
			continue
		}

		intervals = append(intervals, migrationInterval{m.ID, introduced, deprecated})
	}

	// Choose a minimal set of versions that intersect all migration intervals. These will be the
	// points in the upgrade where we need to wait for an out of band migration to finish before
	// proceeding to subsequent versions.
	//
	// The following greedy algorithm chooses the optimal number of versions with a single scan
	// over the intervals:
	//
	//   (1) Order intervals by increasing upper bound
	//   (2) For each interval, choose a new version equal to the interval's upper bound if
	//       no previously chosen version falls within the interval.

	sort.Slice(intervals, func(i, j int) bool {
		return CompareVersions(intervals[i].deprecated, intervals[j].deprecated) == VersionOrderBefore
	})

	points := make([]Version, 0, len(intervals))
	for _, interval := range intervals {
		if len(points) == 0 || CompareVersions(points[len(points)-1], interval.introduced) == VersionOrderBefore {
			v, ok := interval.deprecated.Previous()
			if !ok {
				return nil, errors.Newf("cannot determine version prior to %s", interval.deprecated.String())
			}
			points = append(points, v)
		}
	}

	// Finally, we reconstruct the return value, which pairs each of our chosen versions with the
	// set of migrations that need to finish prior to continuing the upgrade process. When an interval
	// contains multiple chosen versions, we add it only to the largest version so that we delay
	// completion as long as possible (hence the reversal of the points slice).

	coveringSet := make(map[Version][]int, len(intervals))

	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

outer:
	for _, interval := range intervals {
		for _, point := range points {
			// check for intersection
			if pointIntersectsInterval(interval.introduced, interval.deprecated, point) {
				coveringSet[point] = append(coveringSet[point], interval.id)
				continue outer
			}
		}

		panic("unreachable: input interval not covered in output")
	}

	interupts := make([]MigrationInterrupt, 0, len(coveringSet))
	for version, ids := range coveringSet {
		sort.Ints(ids)
		interupts = append(interupts, MigrationInterrupt{version, ids})
	}
	sort.Slice(interupts, func(i, j int) bool {
		return CompareVersions(interupts[i].Version, interupts[j].Version) == VersionOrderBefore
	})

	return interupts, nil
}
