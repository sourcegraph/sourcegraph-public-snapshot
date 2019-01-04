import { Location } from '../types/location'
import { Position } from '../types/position'
import { Range } from '../types/range'
import { URI } from '../types/uri'
import { fromLocation } from './types'

describe('fromLocation', () => {
    test('converts to location', () =>
        expect(
            fromLocation(new Location(URI.parse('x'), new Range(new Position(1, 2), new Position(3, 4)), { a: 1 }))
        ).toEqual({
            uri: 'x',
            range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
            context: { a: 1 },
        }))
})
