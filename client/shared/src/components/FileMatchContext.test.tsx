import { range } from 'lodash'

import { MatchItem } from './FileMatch'
import { calculateMatchGroups, mergeContext } from './FileMatchContext'

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
            const [, grouped] = calculateMatchGroups(testData6ConsecutiveMatches, maxMatches, context)
            expect(grouped).toMatchInlineSnapshot(`
                [
                  {
                    "matches": [
                      {
                        "line": 0,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 1,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 2,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 3,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": true
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
            const [, grouped] = calculateMatchGroups(testData6ConsecutiveMatches, maxMatches, context)
            expect(grouped).toMatchInlineSnapshot(`
                [
                  {
                    "matches": [
                      {
                        "line": 0,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 1,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 2,
                        "character": 0,
                        "highlightLength": 5,
                        "IsInContext": false
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
            const [, grouped] = calculateMatchGroups(testDataRealMatches, maxMatches, context)
            expect(grouped).toMatchInlineSnapshot(`
                [
                  {
                    "matches": [
                      {
                        "line": 0,
                        "character": 51,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 2,
                        "character": 48,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 3,
                        "character": 15,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 3,
                        "character": 39,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 8,
                        "character": 2,
                        "highlightLength": 5,
                        "IsInContext": false
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
                        "IsInContext": false
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
                        "IsInContext": false
                      },
                      {
                        "line": 24,
                        "character": 8,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 24,
                        "character": 19,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 27,
                        "character": 53,
                        "highlightLength": 5,
                        "IsInContext": false
                      },
                      {
                        "line": 28,
                        "character": 3,
                        "highlightLength": 5,
                        "IsInContext": true
                      },
                      {
                        "line": 29,
                        "character": 13,
                        "highlightLength": 5,
                        "IsInContext": true
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
    line: index,
}))

// Real match data from searching a file for `error`.
const testDataRealMatches: MatchItem[] = [
    {
        highlightRanges: [{ start: 51, highlightLength: 5 }],
        line: 0,
    },
    {
        highlightRanges: [{ start: 48, highlightLength: 5 }],
        line: 2,
    },
    {
        highlightRanges: [{ start: 15, highlightLength: 5 }],
        line: 3,
    },
    {
        highlightRanges: [{ start: 39, highlightLength: 5 }],
        line: 3,
    },
    {
        highlightRanges: [{ start: 2, highlightLength: 5 }],
        line: 8,
    },
    {
        highlightRanges: [{ start: 19, highlightLength: 5 }],
        line: 14,
    },
    {
        highlightRanges: [{ start: 11, highlightLength: 5 }],
        line: 20,
    },
    {
        highlightRanges: [{ start: 8, highlightLength: 5 }],
        line: 24,
    },
    {
        highlightRanges: [{ start: 19, highlightLength: 5 }],
        line: 24,
    },
    {
        highlightRanges: [{ start: 53, highlightLength: 5 }],
        line: 27,
    },
    {
        highlightRanges: [{ start: 3, highlightLength: 5 }],
        line: 28,
    },
    {
        highlightRanges: [{ start: 13, highlightLength: 5 }],
        line: 29,
    },
    {
        highlightRanges: [{ start: 2, highlightLength: 5 }],
        line: 30,
    },
    {
        highlightRanges: [{ start: 8, highlightLength: 5 }],
        line: 31,
    },
    {
        highlightRanges: [{ start: 31, highlightLength: 5 }],
        line: 31,
    },
    {
        highlightRanges: [{ start: 11, highlightLength: 5 }],
        line: 32,
    },
    {
        highlightRanges: [{ start: 23, highlightLength: 5 }],
        line: 33,
    },
    {
        highlightRanges: [{ start: 30, highlightLength: 5 }],
        line: 33,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        line: 34,
    },
    {
        highlightRanges: [{ start: 2, highlightLength: 5 }],
        line: 35,
    },
    {
        highlightRanges: [{ start: 18, highlightLength: 5 }],
        line: 40,
    },
    {
        highlightRanges: [{ start: 8, highlightLength: 5 }],
        line: 41,
    },
    {
        highlightRanges: [{ start: 27, highlightLength: 5 }],
        line: 41,
    },
    {
        highlightRanges: [{ start: 55, highlightLength: 5 }],
        line: 41,
    },
    {
        highlightRanges: [{ start: 3, highlightLength: 5 }],
        line: 42,
    },
    {
        highlightRanges: [{ start: 31, highlightLength: 5 }],
        line: 44,
    },
    {
        highlightRanges: [{ start: 33, highlightLength: 5 }],
        line: 45,
    },
    {
        highlightRanges: [{ start: 23, highlightLength: 5 }],
        line: 47,
    },
    {
        highlightRanges: [{ start: 30, highlightLength: 5 }],
        line: 48,
    },
    {
        highlightRanges: [{ start: 37, highlightLength: 5 }],
        line: 48,
    },
    {
        highlightRanges: [{ start: 17, highlightLength: 5 }],
        line: 51,
    },
    {
        highlightRanges: [{ start: 10, highlightLength: 5 }],
        line: 54,
    },
    {
        highlightRanges: [{ start: 32, highlightLength: 5 }],
        line: 60,
    },
    {
        highlightRanges: [{ start: 50, highlightLength: 5 }],
        line: 60,
    },
    {
        highlightRanges: [{ start: 40, highlightLength: 5 }],
        line: 61,
    },
    {
        highlightRanges: [{ start: 20, highlightLength: 5 }],
        line: 62,
    },
    {
        highlightRanges: [{ start: 18, highlightLength: 5 }],
        line: 63,
    },
    {
        highlightRanges: [{ start: 36, highlightLength: 5 }],
        line: 67,
    },
    {
        highlightRanges: [{ start: 54, highlightLength: 5 }],
        line: 67,
    },
    {
        highlightRanges: [{ start: 56, highlightLength: 5 }],
        line: 68,
    },
    {
        highlightRanges: [{ start: 22, highlightLength: 5 }],
        line: 70,
    },
    {
        highlightRanges: [{ start: 62, highlightLength: 5 }],
        line: 74,
    },
    {
        highlightRanges: [{ start: 13, highlightLength: 5 }],
        line: 75,
    },
    {
        highlightRanges: [{ start: 32, highlightLength: 5 }],
        line: 75,
    },
    {
        highlightRanges: [{ start: 70, highlightLength: 5 }],
        line: 84,
    },
    {
        highlightRanges: [{ start: 17, highlightLength: 5 }],
        line: 85,
    },
    {
        highlightRanges: [{ start: 39, highlightLength: 5 }],
        line: 85,
    },
    {
        highlightRanges: [{ start: 23, highlightLength: 5 }],
        line: 94,
    },
    {
        highlightRanges: [{ start: 35, highlightLength: 5 }],
        line: 95,
    },
    {
        highlightRanges: [{ start: 12, highlightLength: 5 }],
        line: 96,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        line: 97,
    },
    {
        highlightRanges: [{ start: 37, highlightLength: 5 }],
        line: 97,
    },
    {
        highlightRanges: [{ start: 14, highlightLength: 5 }],
        line: 98,
    },
    {
        highlightRanges: [{ start: 20, highlightLength: 5 }],
        line: 100,
    },
    {
        highlightRanges: [{ start: 9, highlightLength: 5 }],
        line: 101,
    },
    {
        highlightRanges: [{ start: 27, highlightLength: 5 }],
        line: 109,
    },
    {
        highlightRanges: [{ start: 20, highlightLength: 5 }],
        line: 112,
    },
    {
        highlightRanges: [{ start: 44, highlightLength: 5 }],
        line: 112,
    },
    {
        highlightRanges: [{ start: 14, highlightLength: 5 }],
        line: 113,
    },
    {
        highlightRanges: [{ start: 20, highlightLength: 5 }],
        line: 119,
    },
    {
        highlightRanges: [{ start: 30, highlightLength: 5 }],
        line: 133,
    },
    {
        highlightRanges: [{ start: 23, highlightLength: 5 }],
        line: 134,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        line: 135,
    },
    {
        highlightRanges: [{ start: 19, highlightLength: 5 }],
        line: 136,
    },
    {
        highlightRanges: [{ start: 48, highlightLength: 5 }],
        line: 136,
    },
    {
        highlightRanges: [{ start: 14, highlightLength: 5 }],
        line: 137,
    },
    {
        highlightRanges: [{ start: 14, highlightLength: 5 }],
        line: 143,
    },
    {
        highlightRanges: [{ start: 31, highlightLength: 5 }],
        line: 149,
    },
    {
        highlightRanges: [{ start: 26, highlightLength: 5 }],
        line: 152,
    },
    {
        highlightRanges: [{ start: 10, highlightLength: 5 }],
        line: 160,
    },
    {
        highlightRanges: [{ start: 40, highlightLength: 5 }],
        line: 160,
    },
    {
        highlightRanges: [{ start: 19, highlightLength: 5 }],
        line: 161,
    },
    {
        highlightRanges: [{ start: 12, highlightLength: 5 }],
        line: 162,
    },
    {
        highlightRanges: [{ start: 7, highlightLength: 5 }],
        line: 163,
    },
    {
        highlightRanges: [{ start: 7, highlightLength: 5 }],
        line: 164,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        line: 167,
    },
    {
        highlightRanges: [{ start: 23, highlightLength: 5 }],
        line: 167,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        line: 168,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        line: 171,
    },
    {
        highlightRanges: [{ start: 30, highlightLength: 5 }],
        line: 171,
    },
    {
        highlightRanges: [{ start: 41, highlightLength: 5 }],
        line: 171,
    },
    {
        highlightRanges: [{ start: 10, highlightLength: 5 }],
        line: 172,
    },
    {
        highlightRanges: [{ start: 16, highlightLength: 5 }],
        line: 175,
    },
    {
        highlightRanges: [{ start: 32, highlightLength: 5 }],
        line: 175,
    },
]
