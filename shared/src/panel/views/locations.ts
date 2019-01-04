import { Location } from '@sourcegraph/extension-api-types'
import { isEqual, uniqWith } from 'lodash'

/**
 * Grouped locations returned by {@link groupLocations}.
 *
 * There are multiple levels of grouping. For example, the first level might be the repository, and the second
 * level might be the file. Typically the levels are displayed as columns, with
 *
 * @template L The location type.
 * @template G The type that describes a grouping level.
 */
export interface GroupedLocations<L, G> {
    /**
     * The groups to show at each level.
     */
    groups: { key: G; count: number }[][]

    /**
     * The selected group at each level.
     */
    selectedGroups: G[]

    /**
     * The locations to display (based on the selected group at each level).
     */
    visibleLocations: L[]
}

/**
 * @template L The location type.
 * @template G The type that describes a grouping level.
 */
export function groupLocations<L = Location, G = string>(
    locations: L[],
    selectedGroups: G[] | null,
    groupKeys: ((location: L) => G | undefined)[],
    locationForDefaultSelection: L
): GroupedLocations<L, G> {
    locations = uniqWith<L>(locations, (a, b) => isEqual(a, b))

    const groups: GroupedLocations<L, G>['groups'] = []

    if (!selectedGroups) {
        // Set default selection.
        selectedGroups = []
        for (const groupKey of groupKeys) {
            const group = groupKey(locationForDefaultSelection)
            if (group === undefined) {
                break
            }
            selectedGroups.push(group)
        }
    }

    const visibleLocations: L[] = []
    for (const loc of locations) {
        for (const [i, groupKey] of groupKeys.entries()) {
            const group = groupKey(loc)
            if (group === undefined) {
                break
            }
            if (!groups[i]) {
                groups[i] = []
            }
            const groupEntry = groups[i].find(g => g.key === group)
            if (groupEntry) {
                groupEntry.count++
            } else {
                groups[i].push({ key: group, count: 1 })
            }
            if (selectedGroups[i] === undefined) {
                selectedGroups[i] = group
            }
            if (selectedGroups[i] !== group) {
                // This location won't be visible and won't contribute to any more groups, so stop processing it.
                break
            }

            // If this location is the rightmost selected group, it is visible.
            if (i === groupKeys.length - 1) {
                visibleLocations.push(loc)
            }
        }
    }

    return {
        groups,
        selectedGroups,
        visibleLocations,
    }
}
