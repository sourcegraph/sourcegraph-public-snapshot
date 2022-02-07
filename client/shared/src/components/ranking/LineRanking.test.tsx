import { range } from 'lodash'

import { calculateMatchGroupsSorted, mergeContext } from './LineRanking'
import { MatchItem } from './PerFileResultRanking'
import { testDataRealMatches } from './PerFileResultRankingTestHelpers'

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
            expect(mergeContext(1, [{ line: 5 }])).toEqual([[{ line: 5 }]])
        })
        test('merges overlapping context', () => {
            expect(mergeContext(1, [{ line: 5 }, { line: 6 }])).toEqual([[{ line: 5 }, { line: 6 }]])
        })
        test('merges adjacent context', () => {
            expect(mergeContext(1, [{ line: 5 }, { line: 8 }])).toEqual([[{ line: 5 }, { line: 8 }]])
        })
        test('does not merge context when far enough apart', () => {
            expect(mergeContext(1, [{ line: 5 }, { line: 9 }])).toEqual([[{ line: 5 }], [{ line: 9 }]])
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
                        "line": 0,
                        "character": 0,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 1,
                        "character": 0,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 2,
                        "character": 0,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 3,
                        "character": 0,
                        "highlightLength": 5,
                        "isInContext": true
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
                        "line": 0,
                        "character": 0,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 1,
                        "character": 0,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 2,
                        "character": 0,
                        "highlightLength": 5,
                        "isInContext": false
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
            const { grouped } = calculateMatchGroupsSorted(testDataRealMatches, maxMatches, context)
            expect(grouped).toMatchInlineSnapshot(`
                [
                  {
                    "matches": [
                      {
                        "line": 0,
                        "character": 51,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 2,
                        "character": 48,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 3,
                        "character": 15,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 3,
                        "character": 39,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 8,
                        "character": 2,
                        "highlightLength": 5,
                        "isInContext": false
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
                        "line": 14,
                        "character": 19,
                        "highlightLength": 5,
                        "isInContext": false
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
                        "line": 20,
                        "character": 11,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 24,
                        "character": 8,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 24,
                        "character": 19,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 27,
                        "character": 53,
                        "highlightLength": 5,
                        "isInContext": false
                      },
                      {
                        "line": 28,
                        "character": 3,
                        "highlightLength": 5,
                        "isInContext": true
                      },
                      {
                        "line": 29,
                        "character": 13,
                        "highlightLength": 5,
                        "isInContext": true
                      }
                    ],
                    "position": {
                      "line": 21,
                      "character": 12
                    },
                    "startLine": 18,
                    "endLine": 30
                  }
                ]
            `)
        })
    })
})

// "error" matched 5 times, once per line.
const testData6ConsecutiveMatches: MatchItem[] = range(0, 6).map(index => ({
    highlightRanges: [{ start: 0, highlightLength: 5 }],
    preview: 'error',
    line: index,
}))
