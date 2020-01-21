import { getDiagnostics } from './diagnostics'
import { parseSearchQuery, ParseSuccess, Sequence } from './parser'

describe('getDiagnostics()', () => {
    test('invalid filter type', () => {
        expect(
            getDiagnostics((parseSearchQuery('repos:^github.com/sourcegraph') as ParseSuccess<Sequence>).token)
        ).toStrictEqual([
            {
                endColumn: 6,
                endLineNumber: 1,
                message: 'Invalid filter type.',
                severity: 8,
                startColumn: 1,
                startLineNumber: 1,
            },
        ])
    })

    test('invalid filter value', () => {
        expect(getDiagnostics((parseSearchQuery('case:maybe') as ParseSuccess<Sequence>).token)).toStrictEqual([
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

    test('search query containing colon', () => {
        expect(
            getDiagnostics((parseSearchQuery('Configuration::doStuff(...)') as ParseSuccess<Sequence>).token)
        ).toStrictEqual([
            {
                endColumn: 28,
                endLineNumber: 1,
                message: 'Quoting the query may help if you want a literal match.',
                severity: 4,
                startColumn: 1,
                startLineNumber: 1,
            },
        ])
    })
})
