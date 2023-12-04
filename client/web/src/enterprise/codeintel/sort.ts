import { sortBy } from 'lodash'

import type { Location } from '../../codeintel/location'

/**
 * Sort the locations by their url field's (which is a relative path, e.g.
 * `/github.com/golang/go/-/blob/src/cmd/trace/trace.go`) similarity to the
 * current text document URI path. This is done by applying a similarity
 * coefficient to the segments of each file path. Paths with more segments in
 * common will have a higher similarity coefficient.
 *
 * @param locations A list of locations to sort.
 * @param currentPath The pathname of the current text document.
 */
export function sortByProximity(locations: Location[], currentPath: string): Location[] {
    return sortBy(
        locations,
        ({ url }) => -jaccardIndex(new Set(url.slice(1).split('/')), new Set(currentPath.slice(1).split('/')))
    )
}

/**
 * Calculate the jaccard index, or the Intersection over Union of two sets.
 *
 * @param a The first set.
 * @param b The second set.
 */
function jaccardIndex<T>(a: Set<T>, b: Set<T>): number {
    const aArray = [...a]
    const bArray = [...b]

    return (
        // Get the size of the intersection
        new Set(aArray.filter(value => b.has(value))).size /
        // Get the size of the union
        new Set(aArray.concat(bArray)).size
    )
}
