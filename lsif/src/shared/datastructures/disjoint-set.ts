import { DefaultMap } from './default-map'

/**
 * A modified disjoint set or union-find data structure. Allows linking
 * items together and retrieving the set of all items for a given set.
 */
export class DisjointSet<T> {
    /**
     * For every linked value `v1` and `v2`, `v2` in `links[v1]` and `v1` in `links[v2]`.
     */
    private links = new DefaultMap<T, Set<T>>(() => new Set())

    /**
     * Return an iterator of all elements int he set.
     */
    public keys(): IterableIterator<T> {
        return this.links.keys()
    }

    /**
     * Link two values into the same set. If one or the other value is
     * already in the set, then the sets of the two values will merge.
     *
     * @param v1 One linked value.
     * @param v2 The other linked value.
     */
    public union(v1: T, v2: T): void {
        this.links.getOrDefault(v1).add(v2)
        this.links.getOrDefault(v2).add(v1)
    }

    /**
     * Return the values in the same set as the given source value.
     *
     * @param v The source value.
     */
    public extractSet(v: T): Set<T> {
        const set = new Set<T>()

        let frontier = [v]
        while (frontier.length > 0) {
            const val = frontier.pop()
            if (val === undefined) {
                // hol up, frontier was non-empty!
                throw new Error('Impossible condition.')
            }

            if (!set.has(val)) {
                set.add(val)
                frontier = frontier.concat(Array.from(this.links.get(val) || []))
            }
        }

        return set
    }
}
