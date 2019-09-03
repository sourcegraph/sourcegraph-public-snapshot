import { findRange, asLocations, makeRemoteUri, comparePosition } from './database'
import { RangeData } from './entities'
import * as lsp from 'vscode-languageserver-protocol'
import { Id } from 'lsif-protocol'

describe('findRange', () => {
    it('should find all ranges in list', () => {
        // Generate starting characters for each range. Thse neds to be
        // spread wide enough so that the ranges on each line don't touch.
        const characters = Array.from(Array(2000).keys()).map(i => i * 5)

        const ranges: RangeData[] = []
        for (let i = 1; i <= 1000; i++) {
            const c1 = characters[(i - 1) * 2]
            const c2 = characters[(i - 1) * 2 + 1]

            // Generate two ranges on each line
            ranges.push({ startLine: i, startCharacter: c1, endLine: i, endCharacter: c1 + 3, monikers: [] })
            ranges.push({ startLine: i, startCharacter: c2, endLine: i, endCharacter: c2 + 3, monikers: [] })
        }

        for (const range of ranges) {
            // search for midpoint of each range
            const c = (range.startCharacter + range.endCharacter) / 2
            expect(findRange(ranges, { line: range.startLine, character: c })).toEqual(range)
        }

        for (let i = 1; i <= 1000; i++) {
            // search for the empty space between ranges on a line
            const c = characters[(i - 1) * 2 + 1] - 1
            expect(findRange(ranges, { line: i, character: c })).toBeUndefined()
        }
    })
})

describe('comparePosition', () => {
    it('should return the relative order to a range', () => {
        const range = {
            startLine: 5,
            startCharacter: 11,
            endLine: 5,
            endCharacter: 13,
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

describe('makeRemoteUri', () => {
    it('should generate a URI to another project', () => {
        const pkg = {
            id: 0,
            scheme: '',
            name: '',
            version: '',
            repository: 'github.com/sourcegraph/codeintellify',
            commit: 'deadbeef',
        }

        const uri = makeRemoteUri(pkg, 'src/position.ts')
        expect(uri).toEqual('git://github.com/sourcegraph/codeintellify?deadbeef#src/position.ts')
    })
})

describe('asLocations', () => {
    it('should convert ranges to locations', () => {
        const ranges = new Map<Id, number>()
        ranges.set(1, 0)
        ranges.set(2, 2)
        ranges.set(4, 1)

        const orderedRanges = [
            { startLine: 1, startCharacter: 1, endLine: 1, endCharacter: 2, monikers: [] },
            { startLine: 2, startCharacter: 1, endLine: 2, endCharacter: 2, monikers: [] },
            { startLine: 3, startCharacter: 1, endLine: 3, endCharacter: 2, monikers: [] },
        ]

        expect(asLocations(ranges, orderedRanges, 'src/position.ts', [1, 2, 4])).toEqual([
            lsp.Location.create('src/position.ts', {
                start: { line: 1, character: 1 },
                end: { line: 1, character: 2 },
            }),
            lsp.Location.create('src/position.ts', {
                start: { line: 3, character: 1 },
                end: { line: 3, character: 2 },
            }),
            lsp.Location.create('src/position.ts', {
                start: { line: 2, character: 1 },
                end: { line: 2, character: 2 },
            }),
        ])
    })
})
