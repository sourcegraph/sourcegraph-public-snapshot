import { getDiagnostics } from './diagnostics'
import { scanSearchQuery, ScanSuccess, Sequence } from './scanner'
import { SearchPatternType } from '../../graphql-operations'

describe('getDiagnostics()', () => {
    test('do not raise invalid filter type', () => {
        expect(
            getDiagnostics(
                (scanSearchQuery('repos:^github.com/sourcegraph') as ScanSuccess<Sequence>).token,
                SearchPatternType.literal
            )
        ).toStrictEqual([])
    })

    test('invalid filter value', () => {
        expect(
            getDiagnostics((scanSearchQuery('case:maybe') as ScanSuccess<Sequence>).token, SearchPatternType.literal)
        ).toStrictEqual([
            {
                endColumn: 5,
                endLineNumber: 1,
                message: 'Invalid filter value, expected one of: yes, no.',
                severity: 8,
                startColumn: 1,
                startLineNumber: 1,
            },
        ])
    })

    test('search query containing colon, literal pattern type, do not raise error', () => {
        expect(
            getDiagnostics(
                (scanSearchQuery('Configuration::doStuff(...)') as ScanSuccess<Sequence>).token,
                SearchPatternType.literal
            )
        ).toStrictEqual([])
    })

    test('search query containing quoted token, regexp pattern type', () => {
        expect(
            getDiagnostics(
                (scanSearchQuery('"Configuration::doStuff(...)"') as ScanSuccess<Sequence>).token,
                SearchPatternType.regexp
            )
        ).toStrictEqual([])
    })

    test('search query containing parenthesized parameterss', () => {
        expect(
            getDiagnostics(
                (scanSearchQuery('repo:a (file:b and c)') as ScanSuccess<Sequence>).token,
                SearchPatternType.regexp
            )
        ).toStrictEqual([])
    })

    test('search query containing quoted token, literal pattern type', () => {
        expect(
            getDiagnostics(
                (scanSearchQuery('"Configuration::doStuff(...)"') as ScanSuccess<Sequence>).token,
                SearchPatternType.literal
            )
        ).toStrictEqual([
            {
                endColumn: 30,
                endLineNumber: 1,
                message: 'Your search is interpreted literally and contains quotes. Did you mean to search for quotes?',
                severity: 4,
                startColumn: 1,
                startLineNumber: 1,
            },
        ])
    })
})
