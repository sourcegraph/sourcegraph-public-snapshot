import { isDefined } from '@sourcegraph/common'

// Strip leading/trailing slashes and add a single leading slash
export function sanitizePath(root: string): string {
    return `/${root.replaceAll(/(^\/+)|(\/+$)/g, '')}`
}

// Group values together based on the given function
export function groupBy<V, K>(values: V[], keyFn: (value: V) => K): Map<K, V[]> {
    return values.reduce(
        (acc, val) => acc.set(keyFn(val), (acc.get(keyFn(val)) || []).concat([val])),
        new Map<K, V[]>()
    )
}

// Compare two flattened Map<string, T> entries by key.
export function byKey<T>(tup1: [string, T], tup2: [string, T]): number {
    return tup1[0].localeCompare(tup2[0])
}

// Return the list of keys for the associated values for which the given predicate returned true.
export function keysMatchingPredicate<K, V>(map: Map<K, V>, predFn: (value: V) => boolean): K[] {
    return [...map.entries()].map(([key, value]) => (predFn(value) ? key : undefined)).filter(isDefined)
}

// Return true if the given slices for a proper (pairwise) subset < superset relation
export function checkSubset(subset: string[], superset: string[]): boolean {
    return subset.length < superset.length && subset.filter((value, index) => value !== superset[index]).length === 0
}
