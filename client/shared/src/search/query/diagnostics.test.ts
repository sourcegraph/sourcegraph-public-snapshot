import { SearchPatternType } from '../../graphql-operations'

import { getDiagnostics } from './diagnostics'
import { scanSearchQuery, ScanSuccess, ScanResult } from './scanner'
import { Token } from './token'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

function parseAndDiagnose(query: string, searchPattern: SearchPatternType): ReturnType<typeof getDiagnostics> {
    return getDiagnostics(toSuccess(scanSearchQuery(query)), searchPattern)
}

describe('getDiagnostics()', () => {
    describe('empty and invalid filter values', () => {
        test('do not raise invalid filter type', () => {
            expect(parseAndDiagnose('repos:^github.com/sourcegraph', SearchPatternType.literal)).toMatchInlineSnapshot(
                '[]'
            )
        })

        test('invalid filter value', () => {
            expect(parseAndDiagnose('case:maybe', SearchPatternType.literal)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: Invalid filter value, expected one of: yes, no.",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 5
                  }
                ]
            `)
        })

        test('search query containing colon, literal pattern type, do not raise error', () => {
            expect(parseAndDiagnose('Configuration::doStuff(...)', SearchPatternType.literal)).toMatchInlineSnapshot(
                '[]'
            )
        })

        test('search query containing quoted token, regexp pattern type', () => {
            expect(parseAndDiagnose('"Configuration::doStuff(...)"', SearchPatternType.regexp)).toMatchInlineSnapshot(
                '[]'
            )
        })

        test('search query containing parenthesized parameterss', () => {
            expect(parseAndDiagnose('repo:a (file:b and c)', SearchPatternType.regexp)).toMatchInlineSnapshot('[]')
        })

        test('search query with empty filter diagnostic', () => {
            expect(parseAndDiagnose('meatadata file: ', SearchPatternType.regexp)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 4,
                    "message": "Warning: This filter is empty. Remove the space between the filter and value or quote the value to include the space. E.g., file:\\" a term\\".",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 11,
                    "endColumn": 15
                  }
                ]
            `)
        })

        test('search query with both invalid filter and empty filter returns only one diagnostic for the first issue', () => {
            expect(parseAndDiagnose('meatadata type: ', SearchPatternType.regexp)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: Invalid filter value, expected one of: diff, commit, symbol, repo, path, file.",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 11,
                    "endColumn": 15
                  }
                ]
            `)
        })
    })

    describe('diff and commit only filters', () => {
        test('detects invalid author/before/after/message filters', () => {
            expect(parseAndDiagnose('author:me', SearchPatternType.literal)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: this filter requires 'type:commit' or 'type:diff' in the query",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 7
                  }
                ]
            `)
            expect(parseAndDiagnose('author:me test', SearchPatternType.literal)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: this filter requires 'type:commit' or 'type:diff' in the query",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 7
                  }
                ]
            `)
            expect(
                parseAndDiagnose('author:me before:yesterday after:"last week" message:test', SearchPatternType.literal)
            ).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: this filter requires 'type:commit' or 'type:diff' in the query",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 7
                  },
                  {
                    "severity": 8,
                    "message": "Error: this filter requires 'type:commit' or 'type:diff' in the query",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 11,
                    "endColumn": 17
                  },
                  {
                    "severity": 8,
                    "message": "Error: this filter requires 'type:commit' or 'type:diff' in the query",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 28,
                    "endColumn": 33
                  },
                  {
                    "severity": 8,
                    "message": "Error: this filter requires 'type:commit' or 'type:diff' in the query",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 46,
                    "endColumn": 53
                  }
                ]
            `)
            expect(parseAndDiagnose('until:yesterday since:"last week" m:test', SearchPatternType.literal))
                .toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: this filter requires 'type:commit' or 'type:diff' in the query",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 6
                  },
                  {
                    "severity": 8,
                    "message": "Error: this filter requires 'type:commit' or 'type:diff' in the query",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 17,
                    "endColumn": 22
                  },
                  {
                    "severity": 8,
                    "message": "Error: this filter requires 'type:commit' or 'type:diff' in the query",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 35,
                    "endColumn": 36
                  }
                ]
            `)
        })

        test('accepts author/before/after/message filters if type:diff is present', () => {
            expect(
                parseAndDiagnose(
                    'author:me before:yesterday after:"last week" message:test type:diff',
                    SearchPatternType.literal
                )
            ).toMatchInlineSnapshot('[]')
        })

        test('accepts author/before/after/message filters if type:commit is present', () => {
            expect(
                parseAndDiagnose(
                    'author:me before:yesterday after:"last week" message:test type:diff',
                    SearchPatternType.literal
                )
            ).toMatchInlineSnapshot('[]')
        })
    })

    describe('repo and rev filters', () => {
        test('detects rev without repo filter', () => {
            expect(parseAndDiagnose('rev:main test', SearchPatternType.literal)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: query contains 'rev:' without 'repo:'. Add a 'repo:' filter.",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 4
                  }
                ]
            `)
            expect(parseAndDiagnose('revision:main test', SearchPatternType.literal)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: query contains 'rev:' without 'repo:'. Add a 'repo:' filter.",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 9
                  }
                ]
            `)
        })

        test('detects rev with repo+rev tag filter', () => {
            expect(parseAndDiagnose('rev:main repo:test@dev test', SearchPatternType.literal)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: You have specified both '@' and 'rev:' for a repo filter and I don't know how to interpret this. Remove either '@' or 'rev:'",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 10,
                    "endColumn": 14
                  },
                  {
                    "severity": 8,
                    "message": "Error: You have specified both '@' and 'rev:' for a repo filter and I don't know how to interpret this. Remove either '@' or 'rev:'",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 4
                  }
                ]
            `)
            expect(parseAndDiagnose('rev:main r:test@dev test', SearchPatternType.literal)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: You have specified both '@' and 'rev:' for a repo filter and I don't know how to interpret this. Remove either '@' or 'rev:'",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 10,
                    "endColumn": 11
                  },
                  {
                    "severity": 8,
                    "message": "Error: You have specified both '@' and 'rev:' for a repo filter and I don't know how to interpret this. Remove either '@' or 'rev:'",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 4
                  }
                ]
            `)
        })

        test('detects rev with empty repo filter', () => {
            expect(parseAndDiagnose('rev:main repo: test', SearchPatternType.literal)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 4,
                    "message": "Warning: This filter is empty. Remove the space between the filter and value or quote the value to include the space. E.g., repo:\\" a term\\".",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 10,
                    "endColumn": 14
                  },
                  {
                    "severity": 8,
                    "message": "Error: query contains 'rev:' with an empty 'repo:' filter. Add a non-empty 'repo:' filter.",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 10,
                    "endColumn": 14
                  },
                  {
                    "severity": 8,
                    "message": "Error: query contains 'rev:' with an empty 'repo:' filter. Add a non-empty 'repo:' filter.",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 4
                  }
                ]
            `)
        })

        test('accepts rev filter if valid repo filter is present', () => {
            expect(parseAndDiagnose('repo:test rev:main repo: -repo:main@dev test', SearchPatternType.literal))
                .toMatchInlineSnapshot(`
                [
                  {
                    "severity": 4,
                    "message": "Warning: This filter is empty. Remove the space between the filter and value or quote the value to include the space. E.g., repo:\\" a term\\".",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 20,
                    "endColumn": 24
                  }
                ]
            `)
        })
    })

    describe('structural search and type: filter', () => {
        test('detects type: filter in structural search', () => {
            expect(parseAndDiagnose('type:symbol test', SearchPatternType.structural)).toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: Structural search syntax only applies to searching file contents and is not compatible with 'type:'. Remove this filter or switch to a different search type.",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 5
                  }
                ]
            `)
            expect(parseAndDiagnose('type:symbol test patterntype:structural', SearchPatternType.literal))
                .toMatchInlineSnapshot(`
                [
                  {
                    "severity": 8,
                    "message": "Error: Structural search syntax only applies to searching file contents and is not compatible with 'type:'. Remove this filter or switch to a different search type.",
                    "startLineNumber": 1,
                    "endLineNumber": 1,
                    "startColumn": 1,
                    "endColumn": 5
                  }
                ]
            `)
        })

        test('accepts type: filter in non-structure search', () => {
            expect(parseAndDiagnose('type:symbol test', SearchPatternType.regexp)).toMatchInlineSnapshot('[]')
            expect(parseAndDiagnose('type:symbol test', SearchPatternType.literal)).toMatchInlineSnapshot('[]')
            // patterntype: takes presedence
            expect(
                parseAndDiagnose('type:symbol test patterntype:literal', SearchPatternType.structural)
            ).toMatchInlineSnapshot('[]')
        })
    })
})
