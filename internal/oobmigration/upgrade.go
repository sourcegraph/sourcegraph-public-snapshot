package oobmigration

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func scheduleUpgrade(from, to Version, migrations []yamlMigration) ([]MigrationInterrupt, error) {
	// First, extract the intervals on which the given out of band migrations are defined. If
	// the interval hasn't been deprecated, it's still "open" and does not need to complete for
	// the instance upgrade operation to be successful.

	intervals := make([]migrationInterval, 0, len(migrations))
	for _, m := range migrations {
		if m.DeprecatedVersionMajor == nil {
			continue
		}

		interval := migrationInterval{
			id:         m.ID,
			introduced: Version{Major: m.IntroducedVersionMajor, Minor: m.IntroducedVersionMinor},
			deprecated: Version{Major: *m.DeprecatedVersionMajor, Minor: *m.DeprecatedVersionMinor},
		}

		// Only add intervals that are deprecated within the migration range: `from < deprecated <= to`
		if CompareVersions(from, interval.deprecated) == VersionOrderBefore && CompareVersions(interval.deprecated, to) != VersionOrderAfter {
			intervals = append(intervals, interval)
		}
	}

	// Choose a minimal set of versions that intersect all migration intervals. These will be the
	// points in the upgrade where we need to wait for an out of band migration to finish before
	// proceeding to subsequent versions.
	//
	// The following greedy algorithm chooses the optimal number of versions with a single scan
	// over the intervals:
	//
	//   (1) Order intervals by increasing upper bound
	//   (2) For each interval, choose a new version equal to one version prior to the interval's
	//       upper bound (the last version prior to its deprecation) if no previously chosen version
	//       falls within the interval.

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
	// set of migrations that need to finish prior to continuing the upgrade process.

	interrupts := makeCoveringSet(intervals, points)

	// Sort ascending
	sort.Slice(interrupts, func(i, j int) bool {
		return CompareVersions(interrupts[i].Version, interrupts[j].Version) == VersionOrderBefore
	})
	return interrupts, nil
}

type migrationInterval struct {
	id         int
	introduced Version
	deprecated Version
}

// makeCoveringSet returns a slice of migration interrupts each represeting a target instance version
// and the set of out of band migrations that must complete before migrating away from that version.
// We assume that the given points are ordered in the direction of migration (e.g., asc for upgrades).
func makeCoveringSet(intervals []migrationInterval, points []Version) []MigrationInterrupt {
	coveringSet := make(map[Version][]int, len(intervals))

	// Flip the order of points to delay the oob migration runs as late as possible. This allows
	// us to make maximal upgrade/downgrade process when we encounter a data error that needs a
	// manual fix.
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

	interrupts := make([]MigrationInterrupt, 0, len(coveringSet))
	for version, ids := range coveringSet {
		sort.Ints(ids)
		interrupts = append(interrupts, MigrationInterrupt{version, ids})
	}

	return interrupts
}
