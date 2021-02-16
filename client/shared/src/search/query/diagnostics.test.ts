import { getDiagnostics } from './diagnostics'
import { scanSearchQuery, ScanSuccess, ScanResult } from './scanner'
import { Token } from './token'
import { SearchPatternType } from '../../graphql-operations'

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

describe('getDiagnostics()', () => {
    test('do not raise invalid filter type', () => {
        expect(
            getDiagnostics(toSuccess(scanSearchQuery('repos:^github.com/sourcegraph')), SearchPatternType.literal)
        ).toStrictEqual([])
    })

    test('invalid filter value', () => {
        expect(getDiagnostics(toSuccess(scanSearchQuery('case:maybe')), SearchPatternType.literal)).toStrictEqual([
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
            getDiagnostics(toSuccess(scanSearchQuery('Configuration::doStuff(...)')), SearchPatternType.literal)
        ).toStrictEqual([])
    })

    test('search query containing quoted token, regexp pattern type', () => {
        expect(
            getDiagnostics(toSuccess(scanSearchQuery('"Configuration::doStuff(...)"')), SearchPatternType.regexp)
        ).toStrictEqual([])
    })

    test('search query containing parenthesized parameterss', () => {
        expect(
            getDiagnostics(toSuccess(scanSearchQuery('repo:a (file:b and c)')), SearchPatternType.regexp)
        ).toStrictEqual([])
    })

    test('search query containing quoted token, literal pattern type', () => {
        expect(
            getDiagnostics(toSuccess(scanSearchQuery('"Configuration::doStuff(...)"')), SearchPatternType.literal)
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
