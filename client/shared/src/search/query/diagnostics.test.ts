import { SearchPatternType } from '../../graphql-operations'

import { getDiagnostics } from './diagnostics'
import { scanSearchQuery, ScanSuccess, ScanResult } from './scanner'
import { Token } from './token'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

describe('getDiagnostics()', () => {
    test('do not raise invalid filter type', () => {
        expect(
            getDiagnostics(toSuccess(scanSearchQuery('repos:^github.com/sourcegraph')), SearchPatternType.literal)
        ).toMatchInlineSnapshot('[]')
    })

    test('invalid filter value', () => {
        expect(getDiagnostics(toSuccess(scanSearchQuery('case:maybe')), SearchPatternType.literal))
            .toMatchInlineSnapshot(`
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
        expect(
            getDiagnostics(toSuccess(scanSearchQuery('Configuration::doStuff(...)')), SearchPatternType.literal)
        ).toMatchInlineSnapshot('[]')
    })

    test('search query containing quoted token, regexp pattern type', () => {
        expect(
            getDiagnostics(toSuccess(scanSearchQuery('"Configuration::doStuff(...)"')), SearchPatternType.regexp)
        ).toMatchInlineSnapshot('[]')
    })

    test('search query containing parenthesized parameterss', () => {
        expect(
            getDiagnostics(toSuccess(scanSearchQuery('repo:a (file:b and c)')), SearchPatternType.regexp)
        ).toMatchInlineSnapshot('[]')
    })

    test('search query with empty filter diagnostic', () => {
        expect(getDiagnostics(toSuccess(scanSearchQuery('meatadata file: ')), SearchPatternType.regexp))
            .toMatchInlineSnapshot(`
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
        expect(getDiagnostics(toSuccess(scanSearchQuery('meatadata type: ')), SearchPatternType.regexp))
            .toMatchInlineSnapshot(`
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
