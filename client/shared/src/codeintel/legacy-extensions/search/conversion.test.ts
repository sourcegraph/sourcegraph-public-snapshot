import * as assert from 'assert'

import * as scip from '../../scip'

import { resultToLocation, searchResultToResults } from './conversion'

describe('resultToLocation', () => {
    it('converts to a location', () => {
        const location = resultToLocation({
            repo: 'github.com/foo/bar',
            rev: '84bf4aea50d542be71e0e6339ff8e096b35c84e6',
            file: 'bonk/quux.ts',
            range: scip.Range.fromNumbers(10, 20, 15, 25),
        })

        assert.deepStrictEqual(location, {
            uri: new URL('git://github.com/foo/bar?84bf4aea50d542be71e0e6339ff8e096b35c84e6#bonk/quux.ts'),
            range: scip.Range.fromNumbers(10, 20, 15, 25),
        })
    })

    it('infers HEAD rev', () => {
        const location = resultToLocation({
            repo: 'github.com/foo/bar',
            rev: '',
            file: 'bonk/quux.ts',
            range: scip.Range.fromNumbers(10, 20, 15, 25),
        })

        assert.deepStrictEqual(location, {
            uri: new URL('git://github.com/foo/bar?HEAD#bonk/quux.ts'),
            range: scip.Range.fromNumbers(10, 20, 15, 25),
        })
    })
})

describe('searchResultToResults', () => {
    it('converts to a result list', () => {
        const results = searchResultToResults({
            file: {
                path: 'bonk/quux.ts',
                commit: { oid: '84bf4aea50d542be71e0e6339ff8e096b35c84e6' },
            },
            repository: { name: 'github.com/foo/bar' },
            symbols: [
                {
                    name: 'sym1',
                    fileLocal: true,
                    kind: 'class',
                    location: {
                        resource: { path: 'honk.ts' },
                        range: scip.Range.fromNumbers(1, 2, 3, 4),
                    },
                },
                {
                    name: 'sym2',
                    fileLocal: false,
                    kind: 'class',
                    location: {
                        resource: { path: 'ronk.ts' },
                        range: scip.Range.fromNumbers(4, 5, 6, 7),
                    },
                },
                {
                    name: 'sym3',
                    fileLocal: true,
                    kind: 'method',
                    location: {
                        resource: { path: 'zonk.ts' },
                        range: scip.Range.fromNumbers(6, 7, 8, 9),
                    },
                },
            ],
            lineMatches: [
                {
                    lineNumber: 20,
                    offsetAndLengths: [[3, 5]] as [number, number][],
                },
                {
                    lineNumber: 40,
                    offsetAndLengths: [
                        [1, 3],
                        [4, 6],
                    ] as [number, number][],
                },
            ],
        })

        const common = {
            repo: 'github.com/foo/bar',
            rev: '84bf4aea50d542be71e0e6339ff8e096b35c84e6',
            file: 'bonk/quux.ts',
        }

        assert.deepStrictEqual(results, [
            {
                ...common,
                symbolKind: 'class',
                file: 'honk.ts',
                fileLocal: true,
                range: scip.Range.fromNumbers(1, 2, 3, 4),
            },
            {
                ...common,
                symbolKind: 'class',
                file: 'ronk.ts',
                fileLocal: false,
                range: scip.Range.fromNumbers(4, 5, 6, 7),
            },
            {
                ...common,
                symbolKind: 'method',
                file: 'zonk.ts',
                fileLocal: true,
                range: scip.Range.fromNumbers(6, 7, 8, 9),
            },
            {
                ...common,
                range: scip.Range.fromNumbers(20, 3, 20, 8),
            },
            {
                ...common,
                range: scip.Range.fromNumbers(40, 1, 40, 4),
            },
            {
                ...common,
                range: scip.Range.fromNumbers(40, 4, 40, 10),
            },
        ])
    })
})
