import { OrderedLocationSet } from './location'

const location1 = {
    dumpId: 1,
    path: '1.ts',
    range: { start: { line: 1, character: 10 }, end: { line: 1, character: 15 } },
}

const location2 = {
    dumpId: 2,
    path: '2.ts',
    range: { start: { line: 2, character: 20 }, end: { line: 2, character: 25 } },
}

const location3 = {
    dumpId: 3,
    path: '3.ts',
    range: { start: { line: 3, character: 30 }, end: { line: 3, character: 35 } },
}

const location4 = {
    dumpId: 4,
    path: '4.ts',
    range: { start: { line: 4, character: 40 }, end: { line: 4, character: 45 } },
}

describe('OrderedLocationSet', () => {
    it('should not contain duplicates', () => {
        const locations = new OrderedLocationSet()
        locations.push(location1)
        locations.push(location1)
        locations.push(location2)
        locations.push(location2)

        expect(locations.locations).toEqual([location1, location2])
    })

    it('should retain insertion order', () => {
        const locations = new OrderedLocationSet()
        locations.push(location4)
        locations.push(location3)
        locations.push(location1)
        locations.push(location2)
        locations.push(location2)
        locations.push(location3)
        locations.push(location1)
        locations.push(location4)

        expect(locations.locations).toEqual([location4, location3, location1, location2])
    })
})
