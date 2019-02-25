import { Location } from './location'
import { Position } from './position'
import { Range } from './range'
import { assertToJSON } from './testHelpers'

describe('Location', () => {
    test('toJSON', () => {
        assertToJSON(new Location(new URL('file:///u.ts'), new Position(3, 4)), {
            uri: 'file:///u.ts',
            range: { start: { line: 3, character: 4 }, end: { line: 3, character: 4 } },
        })
        assertToJSON(new Location(new URL('file:///u.ts'), new Range(1, 2, 3, 4)), {
            uri: 'file:///u.ts',
            range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
        })
    })
})
