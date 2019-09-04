import { reachableMonikers } from './importer'
import { Id } from 'lsif-protocol'

describe('reachableMonikers', () => {
    it('should traverse moniker relation graph', () => {
        const monikerSets = new Map<Id, Set<Id>>()
        monikerSets.set(1, new Set<Id>([2]))
        monikerSets.set(2, new Set<Id>([1, 4]))
        monikerSets.set(3, new Set<Id>([4]))
        monikerSets.set(4, new Set<Id>([2, 3]))
        monikerSets.set(5, new Set<Id>([6]))
        monikerSets.set(6, new Set<Id>([5]))

        expect(reachableMonikers(monikerSets, 1)).toEqual(new Set<Id>([1, 2, 3, 4]))
    })
})
