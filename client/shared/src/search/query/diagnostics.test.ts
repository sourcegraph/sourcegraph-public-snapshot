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
        })

        test('accepts author/before/after/message filters if type:diff is present', () => {
            expect(
                parseAndDiagnose(
                    'author:me before:yesterday after:"last week" message:test type:diff',
                    SearchPatternType.literal
                )
            ).toMatchInlineSnapshot(`[]`)
        })

        test('accepts author/before/after/message filters if type:commit is present', () => {
            expect(
                parseAndDiagnose(
                    'author:me before:yesterday after:"last week" message:test type:diff',
                    SearchPatternType.literal
                )
            ).toMatchInlineSnapshot(`[]`)
        })
    })
})
