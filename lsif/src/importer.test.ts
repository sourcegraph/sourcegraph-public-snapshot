import { reachableItems } from './importer'
import { MonikerId } from './models.database'

describe('reachableItems', () => {
    it('should traverse moniker relation graph', () => {
        const linkedItems = new Map<MonikerId, Set<MonikerId>>()
        linkedItems.set(1, new Set<MonikerId>([2]))
        linkedItems.set(2, new Set<MonikerId>([1, 4]))
        linkedItems.set(3, new Set<MonikerId>([4]))
        linkedItems.set(4, new Set<MonikerId>([2, 3]))
        linkedItems.set(5, new Set<MonikerId>([6]))
        linkedItems.set(6, new Set<MonikerId>([5]))

        expect(reachableItems(1, linkedItems)).toEqual(new Set<MonikerId>([1, 2, 3, 4]))
    })
})
