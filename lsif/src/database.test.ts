import * as lsp from 'vscode-languageserver-protocol'
import { comparePosition, createRemoteUri, mapRangesToLocations } from './database'
import { MonikerId, RangeData, RangeId } from './models.database'

describe('comparePosition', () => {
    it('should return the relative order to a range', () => {
        const range = {
            startLine: 5,
            startCharacter: 11,
            endLine: 5,
            endCharacter: 13,
            monikerIds: new Set<MonikerId>(),
        }

        expect(comparePosition(range, { line: 5, character: 11 })).toEqual(0)
        expect(comparePosition(range, { line: 5, character: 12 })).toEqual(0)
        expect(comparePosition(range, { line: 5, character: 13 })).toEqual(0)
        expect(comparePosition(range, { line: 4, character: 12 })).toEqual(+1)
        expect(comparePosition(range, { line: 5, character: 10 })).toEqual(+1)
        expect(comparePosition(range, { line: 5, character: 14 })).toEqual(-1)
        expect(comparePosition(range, { line: 6, character: 12 })).toEqual(-1)
    })
})

describe('createRemoteUri', () => {
    it('should generate a URI to another project', () => {
        const pkg = {
            id: 0,
            scheme: '',
            name: '',
            version: '',
            repository: 'github.com/sourcegraph/codeintellify',
            commit: 'deadbeef',
        }

        const uri = createRemoteUri(pkg, 'src/position.ts')
        expect(uri).toEqual('git://github.com/sourcegraph/codeintellify?deadbeef#src/position.ts')
    })
})

describe('mapRangesToLocations', () => {
    it('should map ranges to locations', () => {
        const ranges = new Map<RangeId, RangeData>()
        ranges.set(1, {
            startLine: 1,
            startCharacter: 1,
            endLine: 1,
            endCharacter: 2,
            monikerIds: new Set<MonikerId>(),
        })
        ranges.set(2, {
            startLine: 3,
            startCharacter: 1,
            endLine: 3,
            endCharacter: 2,
            monikerIds: new Set<MonikerId>(),
        })
        ranges.set(4, {
            startLine: 2,
            startCharacter: 1,
            endLine: 2,
            endCharacter: 2,
            monikerIds: new Set<MonikerId>(),
        })

        const locations = mapRangesToLocations(ranges, 'src/position.ts', new Set([1, 2, 4]))
        expect(locations).toContainEqual(
            lsp.Location.create('src/position.ts', { start: { line: 1, character: 1 }, end: { line: 1, character: 2 } })
        )
        expect(locations).toContainEqual(
            lsp.Location.create('src/position.ts', { start: { line: 3, character: 1 }, end: { line: 3, character: 2 } })
        )
        expect(locations).toContainEqual(
            lsp.Location.create('src/position.ts', { start: { line: 2, character: 1 }, end: { line: 2, character: 2 } })
        )
        expect(locations).toHaveLength(3)
    })
})
