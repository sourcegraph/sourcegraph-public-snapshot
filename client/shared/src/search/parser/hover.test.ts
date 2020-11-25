import { getHoverResult } from './hover'
import { scanSearchQuery, ScanSuccess, Token, ScanResult } from './scanner'
import { SearchPatternType } from '../../graphql-operations'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

describe('getHoverResult()', () => {
    test('returns hover contents for filters', () => {
        const scannedQuery = toSuccess(scanSearchQuery('repo:sourcegraph file:code_intelligence'))
        expect(getHoverResult(scannedQuery, { column: 4 })).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 1,
                "endColumn": 17
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 18 })).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from files matching the given search pattern."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 18,
                "endColumn": 40
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 30 })).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from files matching the given search pattern."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 18,
                "endColumn": 40
              }
            }
        `)
    })

    test('smartQuery flag returns hover contents for fields and regexp values', () => {
        const scannedQuery = toSuccess(scanSearchQuery('repo:^hey$'))
        expect(getHoverResult(scannedQuery, { column: 3 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 1,
                "endColumn": 6
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 6 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in \`^prefix\`. Also often used with the end anchor \`$\` to match an exact string, as in \`^exact$\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 6,
                "endColumn": 7
              }
            }
        `)
    })

    test('smartQuery flag returns hover contents regexp patterns', () => {
        const scannedQuery = toSuccess(scanSearchQuery('\\b.*?', false, SearchPatternType.regexp))
        expect(getHoverResult(scannedQuery, { column: 1 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Word boundary**. Match a position where a word character comes after a non-word character, or vice versa. Typically used to match whole words, as in \`\\\\bword\\\\b\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 1,
                "endColumn": 3
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 2 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Word boundary**. Match a position where a word character comes after a non-word character, or vice versa. Typically used to match whole words, as in \`\\\\bword\\\\b\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 1,
                "endColumn": 3
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 3 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Dot**. Match any character except a line break."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 3,
                "endColumn": 4
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 4 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Zero or more**. Match zero or more of the previous expression."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 4,
                "endColumn": 5
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 5 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Lazy**. Match as few as characters as possible that match the previous expression."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 5,
                "endColumn": 6
              }
            }
        `)
    })

    test('smartQuery flag regexp group range encloses pattern', () => {
        const scannedQuery = toSuccess(scanSearchQuery('(abcd){1,3}', false, SearchPatternType.regexp))
        expect(getHoverResult(scannedQuery, { column: 1 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Group**. Groups together multiple expressions to match."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 1,
                "endColumn": 7
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 2 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Matches the character \`a\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 2,
                "endColumn": 3
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 8 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Range**. Match between 1 and 3 of the previous expression."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 7,
                "endColumn": 12
              }
            }
        `)
    })

    test('smartQuery flag as literal search interprets parentheses as patterns', () => {
        const scannedQuery = toSuccess(scanSearchQuery('(abcd)', false, SearchPatternType.literal))
        expect(getHoverResult(scannedQuery, { column: 1 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Matches the string \`(abcd)\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 1,
                "endColumn": 7
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 2 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Matches the string \`(abcd)\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 1,
                "endColumn": 7
              }
            }
        `)
    })
})
