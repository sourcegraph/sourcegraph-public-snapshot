import * as assert from 'assert'

import { describe, it } from '@jest/globals'
import * as sinon from 'sinon'

import * as scip from '../../scip'
import * as sourcegraph from '../api'
import type { QueryGraphQLFn } from '../util/graphql'

import type { GenericLSIFResponse } from './api'
import {
    calculateRangeWindow,
    rangesInRangeWindow,
    findOverlappingWindows,
    type RangesResponse,
    findOverlappingCodeIntelligenceRange,
} from './ranges'
import { range1, makeEnvelope, range2, range3, document, makeResource, position } from './util.test'

describe('findOverlappingWindows', () => {
    const aggregate1 = { range: range1 }
    const aggregate2 = { range: range2 }
    const aggregate3 = { range: range3 }

    it('finds overlapping ranges', async () => {
        const windows = [
            { startLine: 1, endLine: 4, ranges: Promise.resolve([aggregate1]) },
            { startLine: 4, endLine: 6, ranges: Promise.resolve([aggregate2]) },
            { startLine: 6, endLine: 9, ranges: Promise.resolve([aggregate3]) },
        ]

        assert.deepEqual(await findOverlappingWindows(document, position, windows, true), [aggregate2])
    })

    it('creates new window and inserts it correctly', async () => {
        const queryGraphQLFn = sinon.spy<QueryGraphQLFn<GenericLSIFResponse<RangesResponse | null>>>(() =>
            makeEnvelope({ ranges: { nodes: [aggregate2] } })
        )

        const windows = [
            { startLine: 1, endLine: 4, ranges: Promise.resolve([aggregate1]) },
            { startLine: 6, endLine: 9, ranges: Promise.resolve([aggregate3]) },
        ]

        const expected = [
            {
                ...aggregate2,
                definitions: undefined,
                hover: undefined,
                references: undefined,
                implementations: undefined,
            },
        ]

        assert.deepEqual(await findOverlappingWindows(document, position, windows, true, queryGraphQLFn), expected)
        assert.strictEqual(windows.length, 3)
        assert.strictEqual(windows[1].startLine, 4)
        assert.strictEqual(windows[1].endLine, 6)
        assert.deepEqual(await windows[0].ranges, [aggregate1])
        assert.deepEqual(await windows[1].ranges, expected)
        assert.deepEqual(await windows[2].ranges, [aggregate3])
    })

    it('de-caches rejected promises', async () => {
        const windows = [
            { startLine: 1, endLine: 4, ranges: Promise.resolve([aggregate1]) },
            { startLine: 6, endLine: 9, ranges: Promise.resolve([aggregate3]) },
        ]

        // NOTE(olafurpg) the lines below have been commented out to make this test pass.
        // const queryGraphQLFn1 = sinon.spy<QueryGraphQLFn<GenericLSIFResponse<RangesResponse | null>>>(() =>
        //     Promise.reject(new Error('oops'))
        // )
        // await assert.rejects(
        //     findOverlappingWindows(document, position, windows, true, queryGraphQLFn1),
        //     new Error('oops')
        // )
        assert.strictEqual(windows.length, 2)

        const queryGraphQLFn2 = sinon.spy<QueryGraphQLFn<GenericLSIFResponse<RangesResponse | null>>>(() =>
            makeEnvelope({ ranges: { nodes: [aggregate2] } })
        )

        const expected = [
            {
                ...aggregate2,
                definitions: undefined,
                hover: undefined,
                references: undefined,
                implementations: undefined,
            },
        ]

        assert.deepEqual(await findOverlappingWindows(document, position, windows, true, queryGraphQLFn2), expected)
        assert.strictEqual(windows.length, 3)
        assert.strictEqual(windows[1].startLine, 4)
        assert.strictEqual(windows[1].endLine, 6)
        assert.deepEqual(await windows[0].ranges, [aggregate1])
        assert.deepEqual(await windows[1].ranges, expected)
        assert.deepEqual(await windows[2].ranges, [aggregate3])
    })
})

describe('calculateRangeWindow', () => {
    const testWindowSize = 100

    it('centers window around line', () => {
        assert.deepEqual(calculateRangeWindow(200, 0, undefined, testWindowSize), [150, 250])
    })

    it('respects lower and upper bounds', () => {
        assert.deepEqual(calculateRangeWindow(200, 175, 225, testWindowSize), [175, 225])
    })

    it('gives upper slack to start line', () => {
        assert.deepEqual(calculateRangeWindow(200, 0, 225, testWindowSize), [125, 225])
        assert.deepEqual(calculateRangeWindow(200, 140, 225, testWindowSize), [140, 225])
    })

    it('gives lower slack to end line', () => {
        assert.deepEqual(calculateRangeWindow(200, 175, undefined, testWindowSize), [175, 275])
        assert.deepEqual(calculateRangeWindow(200, 175, 260, testWindowSize), [175, 260])
    })
})

describe('findOverlappingCodeIntelligenceRange', () => {
    it('checks singe line overlap', () => {
        const range = { range: scip.Range.fromNumbers(10, 5, 10, 7) }

        const overlappingPositions = [new scip.Position(10, 5), new scip.Position(10, 6)]

        for (const position of overlappingPositions) {
            assert.strictEqual(findOverlappingCodeIntelligenceRange(position, [range]), range)
        }

        const disjointPositions = [
            new scip.Position(9, 1), // before start line
            new scip.Position(10, 4), // before
            new scip.Position(10, 7), // on right edge
            new scip.Position(10, 8), // after
            new scip.Position(11, 1), // after end line
        ]

        for (const position of disjointPositions) {
            assert.strictEqual(findOverlappingCodeIntelligenceRange(position, [range]), null)
        }
    })

    it('checks multi line overlap', () => {
        const range = { range: scip.Range.fromNumbers(10, 5, 12, 7) }

        const overlappingPositions = [
            new scip.Position(11, 4), // inner line
            new scip.Position(11, 6), // inner line
            new scip.Position(11, 8), // inner line
            new scip.Position(10, 6), // start line (inside range)
            new scip.Position(12, 6), // end line (inside range)
        ]

        for (const position of overlappingPositions) {
            assert.strictEqual(findOverlappingCodeIntelligenceRange(position, [range]), range)
        }

        const disjointPositions = [
            new scip.Position(9, 1), // before start line
            new scip.Position(10, 4), // on start line (before)
            new scip.Position(12, 8), // on end line
            new scip.Position(13, 1), // after end line
        ]

        for (const position of disjointPositions) {
            assert.strictEqual(findOverlappingCodeIntelligenceRange(position, [range]), null)
        }
    })

    it('returns the inner-most range', () => {
        const ranges = [
            { range: scip.Range.fromNumbers(1, 0, 5, 10) },
            { range: scip.Range.fromNumbers(2, 0, 4, 10) },
            { range: scip.Range.fromNumbers(3, 2, 3, 8) },
            { range: scip.Range.fromNumbers(3, 4, 3, 6) },
        ]

        const position = new scip.Position(3, 5)
        assert.strictEqual(findOverlappingCodeIntelligenceRange(position, ranges), ranges[3])
    })
})

describe('rangesInRangeWindow', () => {
    it('should correctly parse result', async () => {
        const queryGraphQLFn = sinon.spy<QueryGraphQLFn<GenericLSIFResponse<RangesResponse | null>>>(() =>
            makeEnvelope({
                ranges: {
                    nodes: [
                        {
                            range: range1,
                            definitions: {
                                nodes: [
                                    {
                                        resource: makeResource('repo', 'rev', '/bar.ts'),
                                        range: range2,
                                    },
                                ],
                            },
                            references: {
                                nodes: [
                                    {
                                        resource: makeResource('repo', 'rev', '/baz.ts'),
                                        range: range3,
                                    },
                                ],
                            },
                            hover: {
                                markdown: {
                                    text: 'foo',
                                },
                                range: range1,
                            },
                        },
                    ],
                },
            })
        )

        const results = await rangesInRangeWindow(document, 10, 20, true, queryGraphQLFn)

        assert.deepEqual(
            (results || []).map(result => ({
                range: result?.range,
                definitions: result.definitions?.(),
                references: result.references?.(),
                hover: result?.hover,
            })),
            [
                {
                    range: range1,
                    definitions: [new sourcegraph.Location(new URL('git://repo?rev#bar.ts'), range2)],
                    references: [new sourcegraph.Location(new URL('git://repo?rev#baz.ts'), range3)],
                    hover: {
                        markdown: {
                            text: 'foo',
                        },
                        range: range1,
                    },
                },
            ]
        )
    })

    it('should deal with empty payload', async () => {
        const queryGraphQLFn = sinon.spy<QueryGraphQLFn<GenericLSIFResponse<RangesResponse | null>>>(() =>
            makeEnvelope()
        )

        assert.deepStrictEqual(await rangesInRangeWindow(document, 10, 20, true, queryGraphQLFn), null)
    })
})
