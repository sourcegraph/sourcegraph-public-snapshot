import { getHoverResult } from './hover'
import { scanSearchQuery, ScanSuccess, ScanResult } from './scanner'
import { Token } from './token'
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
                  "value": "Matches the string \`abcd\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 2,
                "endColumn": 6
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

    test('smartQuery flag on regexp escape characters', () => {
        const scannedQuery = toSuccess(scanSearchQuery('\\q\\r\\n\\.\\\\', false, SearchPatternType.regexp))
        expect(getHoverResult(scannedQuery, { column: 1 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Escaped Character**. The character \`q\` is escaped."
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
                  "value": "**Escaped Character**. Match a carriage return."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 3,
                "endColumn": 5
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 5 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Escaped Character**. Match a new line."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 5,
                "endColumn": 7
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 7 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Escaped Character**. Match the character \`.\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 7,
                "endColumn": 9
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 9 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Escaped Character**. Match the character \`\\\\\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 9,
                "endColumn": 11
              }
            }
        `)
    })

    test('smartQuery flag on ordinary and negated character class', () => {
        const scannedQuery = toSuccess(scanSearchQuery('[^a-z][0-9]', false, SearchPatternType.regexp))
        expect(getHoverResult(scannedQuery, { column: 2 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Negated character class**. Match any character _not_ inside the square brackets."
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

        expect(getHoverResult(scannedQuery, { column: 7 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Character class**. Match any character inside the square brackets."
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

    test('smartQuery flag returns hover contents for revision syntax', () => {
        const scannedQuery = toSuccess(
            scanSearchQuery(
                'repo:^foo$@head:v1.3 rev:*refs/heads/*:*!refs/heads/release*',
                false,
                SearchPatternType.literal
            )
        )

        expect(getHoverResult(scannedQuery, { column: 11 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Search at revision**. Separates a repository pattern and the revisions to search, like commits or branches. The part before the \`@\` specifies the repositories to search, the part after the \`@\` specifies which revisions to search."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 11,
                "endColumn": 12
              }
            }
        `)

        expect(getHoverResult(scannedQuery, { column: 12 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Revision HEAD**. Search the repository at the latest HEAD commit of the default branch."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 12,
                "endColumn": 16
              }
            }
        `)

        expect(getHoverResult(scannedQuery, { column: 16 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 16,
                "endColumn": 17
              }
            }
        `)

        expect(getHoverResult(scannedQuery, { column: 17 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 17,
                "endColumn": 21
              }
            }
        `)

        expect(getHoverResult(scannedQuery, { column: 26 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Revision glob pattern to include**. A prefixing indicating that a glob pattern follows. Git references matching the glob pattern are included in the search. Typically used where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 26,
                "endColumn": 27
              }
            }
        `)

        expect(getHoverResult(scannedQuery, { column: 27 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 27,
                "endColumn": 38
              }
            }
        `)

        expect(getHoverResult(scannedQuery, { column: 41 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Revision glob pattern to exclude**. A prefix indicating that git references, like a commit or branch name, should be **excluded** from search based on the glob pattern that follows. Used in conjunction with a glob pattern that matches a set of commits or branches, followed by a a pattern to exclude from the set. For example, \`*refs/heads/*:*!refs/heads/release*\` searches all branches at the head commit, excluding branches matching \`release*\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 40,
                "endColumn": 42
              }
            }
        `)
    })

    test('smartQuery flag returns hover contents for structural syntax', () => {
        const scannedQuery = toSuccess(scanSearchQuery(':[var~\\w+] ...', false, SearchPatternType.structural))

        expect(getHoverResult(scannedQuery, { column: 1 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 1,
                "endColumn": 11
              }
            }
        `)

        expect(getHoverResult(scannedQuery, { column: 3 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Hole variable**. A descriptive name for the syntax matched by this hole."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 3,
                "endColumn": 6
              }
            }
        `)

        expect(getHoverResult(scannedQuery, { column: 6 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Regular expression separator**. Indicates the start of a regular expression that this hole should match."
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

        expect(getHoverResult(scannedQuery, { column: 9 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**One or more**. Match one or more of the previous expression."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 9,
                "endColumn": 10
              }
            }
        `)
        expect(getHoverResult(scannedQuery, { column: 10 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 1,
                "endColumn": 11
              }
            }
        `)

        expect(getHoverResult(scannedQuery, { column: 12 }, true)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Structural hole**. Matches code structures contextually. See the [syntax reference](https://docs.sourcegraph.com/code_search/reference/structural#syntax-reference) for a complete description."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 12,
                "endColumn": 15
              }
            }
        `)
    })
})
