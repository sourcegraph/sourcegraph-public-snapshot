import { SearchPatternType } from '../../graphql-operations'

import { getHoverResult } from './hover'
import { scanSearchQuery, ScanSuccess, ScanResult } from './scanner'
import { Token } from './token'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

describe('getHoverResult()', () => {
    test('returns hover contents for filters', () => {
        const scannedQuery = toSuccess(scanSearchQuery('repo:sourcegraph file:code_intelligence'))
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "Matches the string \`sourcegraph\`."
                },
                {
                  "value": "Include only results from files matching the given search pattern."
                },
                {
                  "value": "Matches the string \`code_intelligence\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 23,
                "endColumn": 40
              }
            }
        `)
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "Matches the string \`sourcegraph\`."
                },
                {
                  "value": "Include only results from files matching the given search pattern."
                },
                {
                  "value": "Matches the string \`code_intelligence\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 23,
                "endColumn": 40
              }
            }
        `)
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "Matches the string \`sourcegraph\`."
                },
                {
                  "value": "Include only results from files matching the given search pattern."
                },
                {
                  "value": "Matches the string \`code_intelligence\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 23,
                "endColumn": 40
              }
            }
        `)
    })

    test('returns hover contents for fields and regexp values', () => {
        const scannedQuery = toSuccess(scanSearchQuery('repo:^hey$'))
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in \`^prefix\`. Also often used with the end anchor \`$\` to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "Matches the string \`hey\`."
                },
                {
                  "value": "**End anchor**. Match the end of a string. Typically used to match a string suffix, as in \`suffix$\`. Also often used with the start anchor to match an exact string, as in \`^exact$\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 10,
                "endColumn": 11
              }
            }
        `)
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in \`^prefix\`. Also often used with the end anchor \`$\` to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "Matches the string \`hey\`."
                },
                {
                  "value": "**End anchor**. Match the end of a string. Typically used to match a string suffix, as in \`suffix$\`. Also often used with the start anchor to match an exact string, as in \`^exact$\`."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 10,
                "endColumn": 11
              }
            }
        `)
    })

    test('returns hover contents regexp patterns', () => {
        const scannedQuery = toSuccess(scanSearchQuery('\\b.*?', false, SearchPatternType.regexp))
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Word boundary**. Match a position where a word character comes after a non-word character, or vice versa. Typically used to match whole words, as in \`\\\\bword\\\\b\`."
                },
                {
                  "value": "**Dot**. Match any character except a line break."
                },
                {
                  "value": "**Zero or more**. Match zero or more of the previous expression."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Word boundary**. Match a position where a word character comes after a non-word character, or vice versa. Typically used to match whole words, as in \`\\\\bword\\\\b\`."
                },
                {
                  "value": "**Dot**. Match any character except a line break."
                },
                {
                  "value": "**Zero or more**. Match zero or more of the previous expression."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Word boundary**. Match a position where a word character comes after a non-word character, or vice versa. Typically used to match whole words, as in \`\\\\bword\\\\b\`."
                },
                {
                  "value": "**Dot**. Match any character except a line break."
                },
                {
                  "value": "**Zero or more**. Match zero or more of the previous expression."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Word boundary**. Match a position where a word character comes after a non-word character, or vice versa. Typically used to match whole words, as in \`\\\\bword\\\\b\`."
                },
                {
                  "value": "**Dot**. Match any character except a line break."
                },
                {
                  "value": "**Zero or more**. Match zero or more of the previous expression."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Word boundary**. Match a position where a word character comes after a non-word character, or vice versa. Typically used to match whole words, as in \`\\\\bword\\\\b\`."
                },
                {
                  "value": "**Dot**. Match any character except a line break."
                },
                {
                  "value": "**Zero or more**. Match zero or more of the previous expression."
                },
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

    test('regexp group range encloses pattern', () => {
        const scannedQuery = toSuccess(scanSearchQuery('(abcd){1,3}', false, SearchPatternType.regexp))
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Group**. Groups together multiple expressions to match."
                },
                {
                  "value": "Matches the string \`abcd\`."
                },
                {
                  "value": "**Group**. Groups together multiple expressions to match."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Group**. Groups together multiple expressions to match."
                },
                {
                  "value": "Matches the string \`abcd\`."
                },
                {
                  "value": "**Group**. Groups together multiple expressions to match."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Group**. Groups together multiple expressions to match."
                },
                {
                  "value": "Matches the string \`abcd\`."
                },
                {
                  "value": "**Group**. Groups together multiple expressions to match."
                },
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

    test('regexp escape characters', () => {
        const scannedQuery = toSuccess(scanSearchQuery('\\q\\r\\n\\.\\\\', false, SearchPatternType.regexp))
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Escaped Character**. The character \`q\` is escaped."
                },
                {
                  "value": "**Escaped Character**. Match a carriage return."
                },
                {
                  "value": "**Escaped Character**. Match a new line."
                },
                {
                  "value": "**Escaped Character**. Match the character \`.\`."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Escaped Character**. The character \`q\` is escaped."
                },
                {
                  "value": "**Escaped Character**. Match a carriage return."
                },
                {
                  "value": "**Escaped Character**. Match a new line."
                },
                {
                  "value": "**Escaped Character**. Match the character \`.\`."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Escaped Character**. The character \`q\` is escaped."
                },
                {
                  "value": "**Escaped Character**. Match a carriage return."
                },
                {
                  "value": "**Escaped Character**. Match a new line."
                },
                {
                  "value": "**Escaped Character**. Match the character \`.\`."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Escaped Character**. The character \`q\` is escaped."
                },
                {
                  "value": "**Escaped Character**. Match a carriage return."
                },
                {
                  "value": "**Escaped Character**. Match a new line."
                },
                {
                  "value": "**Escaped Character**. Match the character \`.\`."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Escaped Character**. The character \`q\` is escaped."
                },
                {
                  "value": "**Escaped Character**. Match a carriage return."
                },
                {
                  "value": "**Escaped Character**. Match a new line."
                },
                {
                  "value": "**Escaped Character**. Match the character \`.\`."
                },
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

    test('ordinary and negated character class', () => {
        const scannedQuery = toSuccess(scanSearchQuery('[^a-z][0-9]', false, SearchPatternType.regexp))
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Negated character class**. Match any character _not_ inside the square brackets."
                },
                {
                  "value": "**Character range**. Match a character in the range \`a-z\`."
                },
                {
                  "value": "**Character range**. Match a character in the range \`a-z\`."
                },
                {
                  "value": "**Character range**. Match a character in the range \`a-z\`."
                },
                {
                  "value": "**Character class**. Match any character inside the square brackets."
                },
                {
                  "value": "**Character class**. Match any character inside the square brackets."
                },
                {
                  "value": "**Character range**. Match a character in the range \`0-9\`."
                },
                {
                  "value": "**Character range**. Match a character in the range \`0-9\`."
                },
                {
                  "value": "**Character range**. Match a character in the range \`0-9\`."
                },
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

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "**Negated character class**. Match any character _not_ inside the square brackets."
                },
                {
                  "value": "**Character range**. Match a character in the range \`a-z\`."
                },
                {
                  "value": "**Character range**. Match a character in the range \`a-z\`."
                },
                {
                  "value": "**Character range**. Match a character in the range \`a-z\`."
                },
                {
                  "value": "**Character class**. Match any character inside the square brackets."
                },
                {
                  "value": "**Character class**. Match any character inside the square brackets."
                },
                {
                  "value": "**Character range**. Match a character in the range \`0-9\`."
                },
                {
                  "value": "**Character range**. Match a character in the range \`0-9\`."
                },
                {
                  "value": "**Character range**. Match a character in the range \`0-9\`."
                },
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

    test('literal search interprets parentheses as patterns', () => {
        const scannedQuery = toSuccess(scanSearchQuery('(abcd)', false, SearchPatternType.literal))
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
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

    test('returns hover contents for revision syntax', () => {
        const scannedQuery = toSuccess(
            scanSearchQuery(
                'repo:^foo$@head:v1.3 rev:*refs/heads/*:*!refs/heads/release*',
                false,
                SearchPatternType.literal
            )
        )

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in \`^prefix\`. Also often used with the end anchor \`$\` to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "Matches the string \`foo\`."
                },
                {
                  "value": "**End anchor**. Match the end of a string. Typically used to match a string suffix, as in \`suffix$\`. Also often used with the start anchor to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "**Search at revision**. Separates a repository pattern and the revisions to search, like commits or branches. The part before the \`@\` specifies the repositories to search, the part after the \`@\` specifies which revisions to search."
                },
                {
                  "value": "**Revision HEAD**. Search the repository at the latest HEAD commit of the default branch."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "Search a revision (branch, commit hash, or tag) instead of the default branch."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to include**. A prefixing indicating that a glob pattern follows. Git references matching the glob pattern are included in the search. Typically used where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to exclude**. A prefix indicating that git references, like a commit or branch name, should be **excluded** from search based on the glob pattern that follows. Used in conjunction with a glob pattern that matches a set of commits or branches, followed by a a pattern to exclude from the set. For example, \`*refs/heads/*:*!refs/heads/release*\` searches all branches at the head commit, excluding branches matching \`release*\`."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 61,
                "endColumn": 61
              }
            }
        `)

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in \`^prefix\`. Also often used with the end anchor \`$\` to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "Matches the string \`foo\`."
                },
                {
                  "value": "**End anchor**. Match the end of a string. Typically used to match a string suffix, as in \`suffix$\`. Also often used with the start anchor to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "**Search at revision**. Separates a repository pattern and the revisions to search, like commits or branches. The part before the \`@\` specifies the repositories to search, the part after the \`@\` specifies which revisions to search."
                },
                {
                  "value": "**Revision HEAD**. Search the repository at the latest HEAD commit of the default branch."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "Search a revision (branch, commit hash, or tag) instead of the default branch."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to include**. A prefixing indicating that a glob pattern follows. Git references matching the glob pattern are included in the search. Typically used where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to exclude**. A prefix indicating that git references, like a commit or branch name, should be **excluded** from search based on the glob pattern that follows. Used in conjunction with a glob pattern that matches a set of commits or branches, followed by a a pattern to exclude from the set. For example, \`*refs/heads/*:*!refs/heads/release*\` searches all branches at the head commit, excluding branches matching \`release*\`."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 61,
                "endColumn": 61
              }
            }
        `)

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in \`^prefix\`. Also often used with the end anchor \`$\` to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "Matches the string \`foo\`."
                },
                {
                  "value": "**End anchor**. Match the end of a string. Typically used to match a string suffix, as in \`suffix$\`. Also often used with the start anchor to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "**Search at revision**. Separates a repository pattern and the revisions to search, like commits or branches. The part before the \`@\` specifies the repositories to search, the part after the \`@\` specifies which revisions to search."
                },
                {
                  "value": "**Revision HEAD**. Search the repository at the latest HEAD commit of the default branch."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "Search a revision (branch, commit hash, or tag) instead of the default branch."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to include**. A prefixing indicating that a glob pattern follows. Git references matching the glob pattern are included in the search. Typically used where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to exclude**. A prefix indicating that git references, like a commit or branch name, should be **excluded** from search based on the glob pattern that follows. Used in conjunction with a glob pattern that matches a set of commits or branches, followed by a a pattern to exclude from the set. For example, \`*refs/heads/*:*!refs/heads/release*\` searches all branches at the head commit, excluding branches matching \`release*\`."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 61,
                "endColumn": 61
              }
            }
        `)

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in \`^prefix\`. Also often used with the end anchor \`$\` to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "Matches the string \`foo\`."
                },
                {
                  "value": "**End anchor**. Match the end of a string. Typically used to match a string suffix, as in \`suffix$\`. Also often used with the start anchor to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "**Search at revision**. Separates a repository pattern and the revisions to search, like commits or branches. The part before the \`@\` specifies the repositories to search, the part after the \`@\` specifies which revisions to search."
                },
                {
                  "value": "**Revision HEAD**. Search the repository at the latest HEAD commit of the default branch."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "Search a revision (branch, commit hash, or tag) instead of the default branch."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to include**. A prefixing indicating that a glob pattern follows. Git references matching the glob pattern are included in the search. Typically used where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to exclude**. A prefix indicating that git references, like a commit or branch name, should be **excluded** from search based on the glob pattern that follows. Used in conjunction with a glob pattern that matches a set of commits or branches, followed by a a pattern to exclude from the set. For example, \`*refs/heads/*:*!refs/heads/release*\` searches all branches at the head commit, excluding branches matching \`release*\`."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 61,
                "endColumn": 61
              }
            }
        `)

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in \`^prefix\`. Also often used with the end anchor \`$\` to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "Matches the string \`foo\`."
                },
                {
                  "value": "**End anchor**. Match the end of a string. Typically used to match a string suffix, as in \`suffix$\`. Also often used with the start anchor to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "**Search at revision**. Separates a repository pattern and the revisions to search, like commits or branches. The part before the \`@\` specifies the repositories to search, the part after the \`@\` specifies which revisions to search."
                },
                {
                  "value": "**Revision HEAD**. Search the repository at the latest HEAD commit of the default branch."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "Search a revision (branch, commit hash, or tag) instead of the default branch."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to include**. A prefixing indicating that a glob pattern follows. Git references matching the glob pattern are included in the search. Typically used where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to exclude**. A prefix indicating that git references, like a commit or branch name, should be **excluded** from search based on the glob pattern that follows. Used in conjunction with a glob pattern that matches a set of commits or branches, followed by a a pattern to exclude from the set. For example, \`*refs/heads/*:*!refs/heads/release*\` searches all branches at the head commit, excluding branches matching \`release*\`."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 61,
                "endColumn": 61
              }
            }
        `)

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in \`^prefix\`. Also often used with the end anchor \`$\` to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "Matches the string \`foo\`."
                },
                {
                  "value": "**End anchor**. Match the end of a string. Typically used to match a string suffix, as in \`suffix$\`. Also often used with the start anchor to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "**Search at revision**. Separates a repository pattern and the revisions to search, like commits or branches. The part before the \`@\` specifies the repositories to search, the part after the \`@\` specifies which revisions to search."
                },
                {
                  "value": "**Revision HEAD**. Search the repository at the latest HEAD commit of the default branch."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "Search a revision (branch, commit hash, or tag) instead of the default branch."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to include**. A prefixing indicating that a glob pattern follows. Git references matching the glob pattern are included in the search. Typically used where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to exclude**. A prefix indicating that git references, like a commit or branch name, should be **excluded** from search based on the glob pattern that follows. Used in conjunction with a glob pattern that matches a set of commits or branches, followed by a a pattern to exclude from the set. For example, \`*refs/heads/*:*!refs/heads/release*\` searches all branches at the head commit, excluding branches matching \`release*\`."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 61,
                "endColumn": 61
              }
            }
        `)

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Include only results from repositories matching the given search pattern."
                },
                {
                  "value": "**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in \`^prefix\`. Also often used with the end anchor \`$\` to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "Matches the string \`foo\`."
                },
                {
                  "value": "**End anchor**. Match the end of a string. Typically used to match a string suffix, as in \`suffix$\`. Also often used with the start anchor to match an exact string, as in \`^exact$\`."
                },
                {
                  "value": "**Search at revision**. Separates a repository pattern and the revisions to search, like commits or branches. The part before the \`@\` specifies the repositories to search, the part after the \`@\` specifies which revisions to search."
                },
                {
                  "value": "**Revision HEAD**. Search the repository at the latest HEAD commit of the default branch."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "Search a revision (branch, commit hash, or tag) instead of the default branch."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to include**. A prefixing indicating that a glob pattern follows. Git references matching the glob pattern are included in the search. Typically used where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision separator**. Separates multiple revisions to search across. For example, \`1a35d48:feature:3.15\` searches the repository for matches at commit \`1a35d48\`, or a branch named \`feature\`, or a tag \`3.15\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                },
                {
                  "value": "**Revision glob pattern to exclude**. A prefix indicating that git references, like a commit or branch name, should be **excluded** from search based on the glob pattern that follows. Used in conjunction with a glob pattern that matches a set of commits or branches, followed by a a pattern to exclude from the set. For example, \`*refs/heads/*:*!refs/heads/release*\` searches all branches at the head commit, excluding branches matching \`release*\`."
                },
                {
                  "value": "**Revision using git reference path**. Search the branch name or tag at the head commit. Search across git objects, like commits or branches, that match this git reference path. Typically used in conjunction with glob patterns, where a pattern like \`*refs/heads/*\` searches across all repository branches at the head commit."
                },
                {
                  "value": "**Revision wildcard**. Glob syntax to match zero or more characters in a revision. Typically used to match multiple branches or tags based on a git reference path. For example, \`refs/tags/v3.*\` matches all tags that start with \`v3.\`."
                },
                {
                  "value": "**Revision branch name or tag**. Search the branch name or tag at the head commit."
                }
              ],
              "range": {
                "startLineNumber": 1,
                "endLineNumber": 1,
                "startColumn": 61,
                "endColumn": 61
              }
            }
        `)
    })

    test('returns hover contents for structural syntax', () => {
        const scannedQuery = toSuccess(scanSearchQuery(':[var~\\w+] ...', false, SearchPatternType.structural))

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Matches the character \`\`."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
                {
                  "value": "**Hole variable**. A descriptive name for the syntax matched by this hole."
                },
                {
                  "value": "**Regular expression separator**. Indicates the start of a regular expression that this hole should match."
                },
                {
                  "value": "**Word**. Match any word character. "
                },
                {
                  "value": "**One or more**. Match one or more of the previous expression."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
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

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Matches the character \`\`."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
                {
                  "value": "**Hole variable**. A descriptive name for the syntax matched by this hole."
                },
                {
                  "value": "**Regular expression separator**. Indicates the start of a regular expression that this hole should match."
                },
                {
                  "value": "**Word**. Match any word character. "
                },
                {
                  "value": "**One or more**. Match one or more of the previous expression."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
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

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Matches the character \`\`."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
                {
                  "value": "**Hole variable**. A descriptive name for the syntax matched by this hole."
                },
                {
                  "value": "**Regular expression separator**. Indicates the start of a regular expression that this hole should match."
                },
                {
                  "value": "**Word**. Match any word character. "
                },
                {
                  "value": "**One or more**. Match one or more of the previous expression."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
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

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Matches the character \`\`."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
                {
                  "value": "**Hole variable**. A descriptive name for the syntax matched by this hole."
                },
                {
                  "value": "**Regular expression separator**. Indicates the start of a regular expression that this hole should match."
                },
                {
                  "value": "**Word**. Match any word character. "
                },
                {
                  "value": "**One or more**. Match one or more of the previous expression."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
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
        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Matches the character \`\`."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
                {
                  "value": "**Hole variable**. A descriptive name for the syntax matched by this hole."
                },
                {
                  "value": "**Regular expression separator**. Indicates the start of a regular expression that this hole should match."
                },
                {
                  "value": "**Word**. Match any word character. "
                },
                {
                  "value": "**One or more**. Match one or more of the previous expression."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
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

        expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
            {
              "contents": [
                {
                  "value": "Matches the character \`\`."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
                {
                  "value": "**Hole variable**. A descriptive name for the syntax matched by this hole."
                },
                {
                  "value": "**Regular expression separator**. Indicates the start of a regular expression that this hole should match."
                },
                {
                  "value": "**Word**. Match any word character. "
                },
                {
                  "value": "**One or more**. Match one or more of the previous expression."
                },
                {
                  "value": "**Regular expression hole**. Match the regular expression defined inside this hole."
                },
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

test('returns hover contents for select', () => {
    const scannedQuery = toSuccess(scanSearchQuery('select:repo repo:foo', false, SearchPatternType.literal))

    expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
        {
          "contents": [
            {
              "value": "Selects the kind of result to display."
            },
            {
              "value": "Select and display distinct repository paths from search results."
            },
            {
              "value": "Include only results from repositories matching the given search pattern."
            },
            {
              "value": "Matches the string \`foo\`."
            }
          ],
          "range": {
            "startLineNumber": 1,
            "endLineNumber": 1,
            "startColumn": 18,
            "endColumn": 21
          }
        }
    `)
})

test('returns repo:contains hovers', () => {
    const scannedQuery = toSuccess(scanSearchQuery('repo:contains.file(foo)', false, SearchPatternType.literal))

    expect(getHoverResult(scannedQuery)).toMatchInlineSnapshot(`
        {
          "contents": [
            {
              "value": "Include only results from repositories matching the given search pattern."
            },
            {
              "value": "**Built-in predicate**. Search only inside repositories that contain a **file path** matching the regular expression \`foo\`."
            },
            {
              "value": "**Built-in predicate**. Search only inside repositories that contain a **file path** matching the regular expression \`foo\`."
            },
            {
              "value": "**Built-in predicate**. Search only inside repositories that contain a **file path** matching the regular expression \`foo\`."
            },
            {
              "value": "**Built-in predicate**. Search only inside repositories that contain a **file path** matching the regular expression \`foo\`."
            },
            {
              "value": "Matches the string \`foo\`."
            },
            {
              "value": "**Built-in predicate**. Search only inside repositories that contain a **file path** matching the regular expression \`foo\`."
            }
          ],
          "range": {
            "startLineNumber": 1,
            "endLineNumber": 1,
            "startColumn": 6,
            "endColumn": 24
          }
        }
    `)
})
