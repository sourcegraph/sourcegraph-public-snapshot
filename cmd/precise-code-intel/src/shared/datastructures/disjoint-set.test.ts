import { DisjointSet } from './disjoint-set'

describe('DisjointSet', () => {
    it('should traverse relations in both directions', () => {
        const set = new DisjointSet<number>()
        set.union(1, 2)
        set.union(3, 4)
        set.union(1, 3)
        set.union(5, 6)

        expect(set.extractSet(1)).toEqual(new Set([1, 2, 3, 4]))
        expect(set.extractSet(2)).toEqual(new Set([1, 2, 3, 4]))
        expect(set.extractSet(3)).toEqual(new Set([1, 2, 3, 4]))
        expect(set.extractSet(4)).toEqual(new Set([1, 2, 3, 4]))
        expect(set.extractSet(5)).toEqual(new Set([5, 6]))
        expect(set.extractSet(6)).toEqual(new Set([5, 6]))
    })
})
