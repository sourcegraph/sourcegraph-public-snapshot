import { testDataRealMatches } from './PerFileResultRankingTestHelpers'
import { ZoektRanking } from './ZoektRanking'

describe('ZoektRanking', () => {
    const ranking = new ZoektRanking()
    test('collapsedResults', () => {
        expect(ranking.collapsedResults(testDataRealMatches, 1).grouped).toMatchInlineSnapshot(`
            Array [
              Object {
                "endLine": 5,
                "matches": Array [
                  Object {
                    "character": 51,
                    "highlightLength": 5,
                    "isInContext": false,
                    "line": 0,
                  },
                  Object {
                    "character": 48,
                    "highlightLength": 5,
                    "isInContext": false,
                    "line": 2,
                  },
                  Object {
                    "character": 15,
                    "highlightLength": 5,
                    "isInContext": false,
                    "line": 3,
                  },
                  Object {
                    "character": 39,
                    "highlightLength": 5,
                    "isInContext": false,
                    "line": 3,
                  },
                ],
                "position": Object {
                  "character": 52,
                  "line": 1,
                },
                "startLine": 0,
              },
            ]
        `)
    })
    test('expandedResults', () => {
        // reverse the data to demonstrate that zoekt-ranking does not sort the
        // results by line number, it preserves the original sort from the
        // server.
        const dataReversed = [...testDataRealMatches].reverse().slice(0, 3)
        expect(ranking.expandedResults(dataReversed, 1).grouped).toMatchInlineSnapshot(`
            Array [
              Object {
                "endLine": 177,
                "matches": Array [
                  Object {
                    "character": 16,
                    "highlightLength": 5,
                    "isInContext": false,
                    "line": 175,
                  },
                  Object {
                    "character": 32,
                    "highlightLength": 5,
                    "isInContext": false,
                    "line": 175,
                  },
                ],
                "position": Object {
                  "character": 17,
                  "line": 176,
                },
                "startLine": 174,
              },
              Object {
                "endLine": 174,
                "matches": Array [
                  Object {
                    "character": 10,
                    "highlightLength": 5,
                    "isInContext": false,
                    "line": 172,
                  },
                ],
                "position": Object {
                  "character": 11,
                  "line": 173,
                },
                "startLine": 171,
              },
            ]
        `)
    })
})
