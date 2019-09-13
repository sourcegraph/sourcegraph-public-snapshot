import { reachableMonikers } from './importer'
import { MonikerId } from './models.database'

describe('reachableMonikers', () => {
    it('should traverse moniker relation graph', () => {
        const monikerSets = new Map<MonikerId, Set<MonikerId>>()
        monikerSets.set(1, new Set<MonikerId>([2]))
        monikerSets.set(2, new Set<MonikerId>([1, 4]))
        monikerSets.set(3, new Set<MonikerId>([4]))
        monikerSets.set(4, new Set<MonikerId>([2, 3]))
        monikerSets.set(5, new Set<MonikerId>([6]))
        monikerSets.set(6, new Set<MonikerId>([5]))

        expect(reachableMonikers(monikerSets, 1)).toEqual(new Set<MonikerId>([1, 2, 3, 4]))
    })
})
