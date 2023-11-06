import { describe, expect, test } from '@jest/globals'
import { range } from 'lodash'

import { calculateMatchGroupsSorted, mergeContext } from './LineRanking'
import type { MatchItem } from './PerFileResultRanking'
import { testDataRealMatchesByLineNumber } from './PerFileResultRankingTestHelpers'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

describe('components/FileMatchContext', () => {
    describe('mergeContext', () => {
        test('handles empty input', () => {
            expect(mergeContext(1, [])).toEqual([])
        })
        test('does not merge context when there is only one line', () => {
            expect(mergeContext(1, [{ startLine: 5, startCharacter: 0, endLine: 5, endCharacter: 1 }])).toEqual([
                [{ startLine: 5, startCharacter: 0, endLine: 5, endCharacter: 1 }],
            ])
        })
        test('merges overlapping context', () => {
            expect(
                mergeContext(1, [
                    { startLine: 5, startCharacter: 0, endLine: 5, endCharacter: 1 },
                    { startLine: 6, startCharacter: 0, endLine: 6, endCharacter: 1 },
                ])
            ).toEqual([
                [
                    { startLine: 5, startCharacter: 0, endLine: 5, endCharacter: 1 },
                    { startLine: 6, startCharacter: 0, endLine: 6, endCharacter: 1 },
                ],
            ])
        })
        test('merges adjacent context', () => {
            expect(
                mergeContext(1, [
                    { startLine: 5, startCharacter: 0, endLine: 5, endCharacter: 1 },
                    { startLine: 8, startCharacter: 0, endLine: 8, endCharacter: 1 },
                ])
            ).toEqual([
                [
                    { startLine: 5, startCharacter: 0, endLine: 5, endCharacter: 1 },
                    { startLine: 8, startCharacter: 0, endLine: 8, endCharacter: 1 },
                ],
            ])
        })
        test('does not merge context when far enough apart', () => {
            expect(
                mergeContext(1, [
                    { startLine: 5, startCharacter: 0, endLine: 5, endCharacter: 1 },
                    { startLine: 9, startCharacter: 0, endLine: 9, endCharacter: 1 },
                ])
            ).toEqual([
                [{ startLine: 5, startCharacter: 0, endLine: 5, endCharacter: 1 }],
                [{ startLine: 9, startCharacter: 0, endLine: 9, endCharacter: 1 }],
            ])
        })
    })

    describe('calculateMatchGroups', () => {
        test('simple', () => {
            const maxMatches = 3
            const context = 1
            const { grouped } = calculateMatchGroupsSorted(testData6ConsecutiveMatches, maxMatches, context)
            expect(grouped).toMatchInlineSnapshot(`
                [
                  {
                    "matches": [
                      {
                        "startLine": 0,
                        "startCharacter": 0,
                        "endLine": 0,
                        "endCharacter": 5
                      },
                      {
                        "startLine": 1,
                        "startCharacter": 0,
                        "endLine": 1,
                        "endCharacter": 5
                      },
                      {
                        "startLine": 2,
                        "startCharacter": 0,
                        "endLine": 2,
                        "endCharacter": 5
                      },
                      {
                        "startLine": 3,
                        "startCharacter": 0,
                        "endLine": 3,
                        "endCharacter": 5
                      }
                    ],
                    "position": {
                      "line": 1,
                      "character": 1
                    },
                    "startLine": 0,
                    "endLine": 4
                  }
                ]
            `)
        })

        test('no context', () => {
            const maxMatches = 3
            const context = 0
            const { grouped } = calculateMatchGroupsSorted(testData6ConsecutiveMatches, maxMatches, context)
            expect(grouped).toMatchInlineSnapshot(`
                [
                  {
                    "matches": [
                      {
                        "startLine": 0,
                        "startCharacter": 0,
                        "endLine": 0,
                        "endCharacter": 5
                      },
                      {
                        "startLine": 1,
                        "startCharacter": 0,
                        "endLine": 1,
                        "endCharacter": 5
                      },
                      {
                        "startLine": 2,
                        "startCharacter": 0,
                        "endLine": 2,
                        "endCharacter": 5
                      }
                    ],
                    "position": {
                      "line": 1,
                      "character": 1
                    },
                    "startLine": 0,
                    "endLine": 3
                  }
                ]
            `)
        })

        test('complex grouping', () => {
            const maxMatches = 10
            const context = 2
            const { grouped } = calculateMatchGroupsSorted(testDataRealMatchesByLineNumber, maxMatches, context)
            expect(grouped).toMatchInlineSnapshot(`
                [
                  {
                    "matches": [
                      {
                        "startLine": 0,
                        "startCharacter": 51,
                        "endLine": 0,
                        "endCharacter": 56
                      },
                      {
                        "startLine": 2,
                        "startCharacter": 48,
                        "endLine": 2,
                        "endCharacter": 53
                      },
                      {
                        "startLine": 3,
                        "startCharacter": 15,
                        "endLine": 3,
                        "endCharacter": 20
                      },
                      {
                        "startLine": 3,
                        "startCharacter": 39,
                        "endLine": 3,
                        "endCharacter": 44
                      },
                      {
                        "startLine": 8,
                        "startCharacter": 2,
                        "endLine": 8,
                        "endCharacter": 7
                      }
                    ],
                    "position": {
                      "line": 1,
                      "character": 52
                    },
                    "startLine": 0,
                    "endLine": 11
                  },
                  {
                    "matches": [
                      {
                        "startLine": 14,
                        "startCharacter": 19,
                        "endLine": 14,
                        "endCharacter": 24
                      }
                    ],
                    "position": {
                      "line": 15,
                      "character": 20
                    },
                    "startLine": 12,
                    "endLine": 17
                  },
                  {
                    "matches": [
                      {
                        "startLine": 20,
                        "startCharacter": 11,
                        "endLine": 20,
                        "endCharacter": 16
                      },
                      {
                        "startLine": 24,
                        "startCharacter": 8,
                        "endLine": 24,
                        "endCharacter": 13
                      },
                      {
                        "startLine": 24,
                        "startCharacter": 19,
                        "endLine": 24,
                        "endCharacter": 24
                      },
                      {
                        "startLine": 27,
                        "startCharacter": 53,
                        "endLine": 27,
                        "endCharacter": 58
                      },
                      {
                        "startLine": 28,
                        "startCharacter": 3,
                        "endLine": 28,
                        "endCharacter": 8
                      },
                      {
                        "startLine": 29,
                        "startCharacter": 13,
                        "endLine": 29,
                        "endCharacter": 18
                      },
                      {
                        "startLine": 30,
                        "startCharacter": 2,
                        "endLine": 30,
                        "endCharacter": 7
                      },
                      {
                        "startLine": 31,
                        "startCharacter": 8,
                        "endLine": 31,
                        "endCharacter": 13
                      },
                      {
                        "startLine": 31,
                        "startCharacter": 31,
                        "endLine": 31,
                        "endCharacter": 36
                      }
                    ],
                    "position": {
                      "line": 21,
                      "character": 12
                    },
                    "startLine": 18,
                    "endLine": 32
                  }
                ]
            `)
        })

        test('no matches', () => {
            const maxMatches = 3
            const context = 1
            const { grouped } = calculateMatchGroupsSorted([], maxMatches, context)
            expect(grouped).toMatchInlineSnapshot('[]')
        })
    })
})

// "error" matched 5 times, once per line.
const testData6ConsecutiveMatches: MatchItem[] = range(0, 6).map(index => ({
    highlightRanges: [{ startLine: index, startCharacter: 0, endLine: index, endCharacter: 5 }],
    content: 'error',
    startLine: index,
    endLine: index,
}))
