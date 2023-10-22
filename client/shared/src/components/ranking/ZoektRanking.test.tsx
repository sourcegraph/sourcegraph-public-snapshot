import { describe, expect, test } from '@jest/globals'

import {
    testDataRealMatchesByLineNumber,
    testDataRealMatchesByZoektRanking,
    testDataRealMultilineMatches,
} from './PerFileResultRankingTestHelpers'
import { ZoektRanking } from './ZoektRanking'

describe('ZoektRanking', () => {
    const ranking = new ZoektRanking(5)
    test('collapsedResults, single-line matches only', () => {
        expect(ranking.collapsedResults(testDataRealMatchesByLineNumber, 1).grouped).toMatchInlineSnapshot(`
            Array [
              Object {
                "endLine": 5,
                "matches": Array [
                  Object {
                    "endCharacter": 56,
                    "endLine": 0,
                    "startCharacter": 51,
                    "startLine": 0,
                  },
                  Object {
                    "endCharacter": 53,
                    "endLine": 2,
                    "startCharacter": 48,
                    "startLine": 2,
                  },
                  Object {
                    "endCharacter": 20,
                    "endLine": 3,
                    "startCharacter": 15,
                    "startLine": 3,
                  },
                  Object {
                    "endCharacter": 44,
                    "endLine": 3,
                    "startCharacter": 39,
                    "startLine": 3,
                  },
                ],
                "position": Object {
                  "character": 52,
                  "line": 1,
                },
                "startLine": 0,
              },
              Object {
                "endLine": 10,
                "matches": Array [
                  Object {
                    "endCharacter": 7,
                    "endLine": 8,
                    "startCharacter": 2,
                    "startLine": 8,
                  },
                ],
                "position": Object {
                  "character": 3,
                  "line": 9,
                },
                "startLine": 7,
              },
              Object {
                "endLine": 16,
                "matches": Array [
                  Object {
                    "endCharacter": 24,
                    "endLine": 14,
                    "startCharacter": 19,
                    "startLine": 14,
                  },
                ],
                "position": Object {
                  "character": 20,
                  "line": 15,
                },
                "startLine": 13,
              },
              Object {
                "endLine": 22,
                "matches": Array [
                  Object {
                    "endCharacter": 16,
                    "endLine": 20,
                    "startCharacter": 11,
                    "startLine": 20,
                  },
                ],
                "position": Object {
                  "character": 12,
                  "line": 21,
                },
                "startLine": 19,
              },
              Object {
                "endLine": 37,
                "matches": Array [
                  Object {
                    "endCharacter": 13,
                    "endLine": 24,
                    "startCharacter": 8,
                    "startLine": 24,
                  },
                  Object {
                    "endCharacter": 24,
                    "endLine": 24,
                    "startCharacter": 19,
                    "startLine": 24,
                  },
                  Object {
                    "endCharacter": 58,
                    "endLine": 27,
                    "startCharacter": 53,
                    "startLine": 27,
                  },
                  Object {
                    "endCharacter": 8,
                    "endLine": 28,
                    "startCharacter": 3,
                    "startLine": 28,
                  },
                  Object {
                    "endCharacter": 18,
                    "endLine": 29,
                    "startCharacter": 13,
                    "startLine": 29,
                  },
                  Object {
                    "endCharacter": 7,
                    "endLine": 30,
                    "startCharacter": 2,
                    "startLine": 30,
                  },
                  Object {
                    "endCharacter": 13,
                    "endLine": 31,
                    "startCharacter": 8,
                    "startLine": 31,
                  },
                  Object {
                    "endCharacter": 36,
                    "endLine": 31,
                    "startCharacter": 31,
                    "startLine": 31,
                  },
                  Object {
                    "endCharacter": 37,
                    "endLine": 32,
                    "startCharacter": 11,
                    "startLine": 32,
                  },
                  Object {
                    "endCharacter": 28,
                    "endLine": 33,
                    "startCharacter": 23,
                    "startLine": 33,
                  },
                  Object {
                    "endCharacter": 35,
                    "endLine": 33,
                    "startCharacter": 30,
                    "startLine": 33,
                  },
                  Object {
                    "endCharacter": 21,
                    "endLine": 34,
                    "startCharacter": 16,
                    "startLine": 34,
                  },
                  Object {
                    "endCharacter": 7,
                    "endLine": 35,
                    "startCharacter": 2,
                    "startLine": 35,
                  },
                ],
                "position": Object {
                  "character": 9,
                  "line": 25,
                },
                "startLine": 23,
              },
            ]
        `)
    })

    test('collapsed results, single-line matches only, matches in order of zoekt relevance ranking', () => {
        expect(ranking.collapsedResults(testDataRealMatchesByZoektRanking, 1).grouped).toMatchInlineSnapshot(`
            Array [
              Object {
                "endLine": 170,
                "matches": Array [
                  Object {
                    "endCharacter": 15,
                    "endLine": 160,
                    "startCharacter": 10,
                    "startLine": 160,
                  },
                  Object {
                    "endCharacter": 45,
                    "endLine": 160,
                    "startCharacter": 40,
                    "startLine": 160,
                  },
                  Object {
                    "endCharacter": 24,
                    "endLine": 161,
                    "startCharacter": 19,
                    "startLine": 161,
                  },
                  Object {
                    "endCharacter": 17,
                    "endLine": 162,
                    "startCharacter": 12,
                    "startLine": 162,
                  },
                  Object {
                    "endCharacter": 12,
                    "endLine": 163,
                    "startCharacter": 7,
                    "startLine": 163,
                  },
                  Object {
                    "endCharacter": 12,
                    "endLine": 164,
                    "startCharacter": 7,
                    "startLine": 164,
                  },
                  Object {
                    "endCharacter": 21,
                    "endLine": 167,
                    "startCharacter": 16,
                    "startLine": 167,
                  },
                  Object {
                    "endCharacter": 28,
                    "endLine": 167,
                    "startCharacter": 23,
                    "startLine": 167,
                  },
                  Object {
                    "endCharacter": 21,
                    "endLine": 168,
                    "startCharacter": 16,
                    "startLine": 168,
                  },
                ],
                "position": Object {
                  "character": 11,
                  "line": 161,
                },
                "startLine": 159,
              },
              Object {
                "endLine": 29,
                "matches": Array [
                  Object {
                    "endCharacter": 13,
                    "endLine": 24,
                    "startCharacter": 8,
                    "startLine": 24,
                  },
                  Object {
                    "endCharacter": 24,
                    "endLine": 24,
                    "startCharacter": 19,
                    "startLine": 24,
                  },
                  Object {
                    "endCharacter": 58,
                    "endLine": 27,
                    "startCharacter": 53,
                    "startLine": 27,
                  },
                ],
                "position": Object {
                  "character": 9,
                  "line": 25,
                },
                "startLine": 23,
              },
              Object {
                "endLine": 177,
                "matches": Array [
                  Object {
                    "endCharacter": 21,
                    "endLine": 171,
                    "startCharacter": 16,
                    "startLine": 171,
                  },
                  Object {
                    "endCharacter": 35,
                    "endLine": 171,
                    "startCharacter": 30,
                    "startLine": 171,
                  },
                  Object {
                    "endCharacter": 46,
                    "endLine": 171,
                    "startCharacter": 41,
                    "startLine": 171,
                  },
                  Object {
                    "endCharacter": 15,
                    "endLine": 172,
                    "startCharacter": 10,
                    "startLine": 172,
                  },
                  Object {
                    "endCharacter": 21,
                    "endLine": 175,
                    "startCharacter": 16,
                    "startLine": 175,
                  },
                  Object {
                    "endCharacter": 37,
                    "endLine": 175,
                    "startCharacter": 32,
                    "startLine": 175,
                  },
                ],
                "position": Object {
                  "character": 17,
                  "line": 172,
                },
                "startLine": 170,
              },
              Object {
                "endLine": 5,
                "matches": Array [
                  Object {
                    "endCharacter": 56,
                    "endLine": 0,
                    "startCharacter": 51,
                    "startLine": 0,
                  },
                  Object {
                    "endCharacter": 53,
                    "endLine": 2,
                    "startCharacter": 48,
                    "startLine": 2,
                  },
                  Object {
                    "endCharacter": 20,
                    "endLine": 3,
                    "startCharacter": 15,
                    "startLine": 3,
                  },
                  Object {
                    "endCharacter": 44,
                    "endLine": 3,
                    "startCharacter": 39,
                    "startLine": 3,
                  },
                ],
                "position": Object {
                  "character": 52,
                  "line": 1,
                },
                "startLine": 0,
              },
              Object {
                "endLine": 16,
                "matches": Array [
                  Object {
                    "endCharacter": 24,
                    "endLine": 14,
                    "startCharacter": 19,
                    "startLine": 14,
                  },
                ],
                "position": Object {
                  "character": 20,
                  "line": 15,
                },
                "startLine": 13,
              },
            ]
        `)
    })

    test('collapsed results, multiline matches only', () => {
        expect(ranking.collapsedResults(testDataRealMultilineMatches, 1).grouped).toMatchInlineSnapshot(`
            Array [
              Object {
                "endLine": 54,
                "matches": Array [
                  Object {
                    "endCharacter": 2,
                    "endLine": 52,
                    "startCharacter": 1,
                    "startLine": 50,
                  },
                ],
                "position": Object {
                  "character": 2,
                  "line": 51,
                },
                "startLine": 49,
              },
              Object {
                "endLine": 74,
                "matches": Array [
                  Object {
                    "endCharacter": 1,
                    "endLine": 65,
                    "startCharacter": 19,
                    "startLine": 60,
                  },
                  Object {
                    "endCharacter": 1,
                    "endLine": 72,
                    "startCharacter": 23,
                    "startLine": 67,
                  },
                ],
                "position": Object {
                  "character": 20,
                  "line": 61,
                },
                "startLine": 59,
              },
              Object {
                "endLine": 81,
                "matches": Array [
                  Object {
                    "endCharacter": 2,
                    "endLine": 79,
                    "startCharacter": 1,
                    "startLine": 77,
                  },
                ],
                "position": Object {
                  "character": 2,
                  "line": 78,
                },
                "startLine": 76,
              },
              Object {
                "endLine": 91,
                "matches": Array [
                  Object {
                    "endCharacter": 2,
                    "endLine": 89,
                    "startCharacter": 1,
                    "startLine": 87,
                  },
                ],
                "position": Object {
                  "character": 2,
                  "line": 88,
                },
                "startLine": 86,
              },
              Object {
                "endLine": 105,
                "matches": Array [
                  Object {
                    "endCharacter": 3,
                    "endLine": 103,
                    "startCharacter": 2,
                    "startLine": 101,
                  },
                ],
                "position": Object {
                  "character": 3,
                  "line": 102,
                },
                "startLine": 100,
              },
            ]
        `)
    })

    test('expandedResults, single-line matches only', () => {
        // reverse the data to demonstrate that zoekt-ranking does not sort the
        // results by line number, it preserves the original sort from the
        // server.
        const dataReversed = [...testDataRealMatchesByLineNumber].reverse().slice(0, 6)
        expect(ranking.expandedResults(dataReversed, 1).grouped).toMatchInlineSnapshot(`
            Array [
              Object {
                "endLine": 177,
                "matches": Array [
                  Object {
                    "endCharacter": 12,
                    "endLine": 164,
                    "startCharacter": 7,
                    "startLine": 164,
                  },
                  Object {
                    "endCharacter": 21,
                    "endLine": 167,
                    "startCharacter": 16,
                    "startLine": 167,
                  },
                  Object {
                    "endCharacter": 28,
                    "endLine": 167,
                    "startCharacter": 23,
                    "startLine": 167,
                  },
                  Object {
                    "endCharacter": 21,
                    "endLine": 168,
                    "startCharacter": 16,
                    "startLine": 168,
                  },
                  Object {
                    "endCharacter": 21,
                    "endLine": 171,
                    "startCharacter": 16,
                    "startLine": 171,
                  },
                  Object {
                    "endCharacter": 35,
                    "endLine": 171,
                    "startCharacter": 30,
                    "startLine": 171,
                  },
                  Object {
                    "endCharacter": 46,
                    "endLine": 171,
                    "startCharacter": 41,
                    "startLine": 171,
                  },
                  Object {
                    "endCharacter": 15,
                    "endLine": 172,
                    "startCharacter": 10,
                    "startLine": 172,
                  },
                  Object {
                    "endCharacter": 21,
                    "endLine": 175,
                    "startCharacter": 16,
                    "startLine": 175,
                  },
                  Object {
                    "endCharacter": 37,
                    "endLine": 175,
                    "startCharacter": 32,
                    "startLine": 175,
                  },
                ],
                "position": Object {
                  "character": 8,
                  "line": 165,
                },
                "startLine": 163,
              },
            ]
        `)
    })

    test('expanded matches, multiline matches only', () => {
        const multilineDataReversed = [...testDataRealMultilineMatches].reverse().slice(0, 6)
        expect(ranking.expandedResults(multilineDataReversed, 1).grouped).toMatchInlineSnapshot(`
            Array [
              Object {
                "endLine": 129,
                "matches": Array [
                  Object {
                    "endCharacter": 2,
                    "endLine": 118,
                    "startCharacter": 1,
                    "startLine": 116,
                  },
                  Object {
                    "endCharacter": 3,
                    "endLine": 123,
                    "startCharacter": 2,
                    "startLine": 121,
                  },
                  Object {
                    "endCharacter": 3,
                    "endLine": 127,
                    "startCharacter": 2,
                    "startLine": 125,
                  },
                ],
                "position": Object {
                  "character": 2,
                  "line": 117,
                },
                "startLine": 115,
              },
              Object {
                "endLine": 105,
                "matches": Array [
                  Object {
                    "endCharacter": 3,
                    "endLine": 103,
                    "startCharacter": 2,
                    "startLine": 101,
                  },
                ],
                "position": Object {
                  "character": 3,
                  "line": 102,
                },
                "startLine": 100,
              },
              Object {
                "endLine": 91,
                "matches": Array [
                  Object {
                    "endCharacter": 2,
                    "endLine": 89,
                    "startCharacter": 1,
                    "startLine": 87,
                  },
                ],
                "position": Object {
                  "character": 2,
                  "line": 88,
                },
                "startLine": 86,
              },
              Object {
                "endLine": 81,
                "matches": Array [
                  Object {
                    "endCharacter": 2,
                    "endLine": 79,
                    "startCharacter": 1,
                    "startLine": 77,
                  },
                ],
                "position": Object {
                  "character": 2,
                  "line": 78,
                },
                "startLine": 76,
              },
            ]
        `)
    })
})
