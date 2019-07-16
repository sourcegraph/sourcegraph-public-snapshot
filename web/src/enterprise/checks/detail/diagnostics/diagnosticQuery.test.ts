import {
    appendToDiagnosticQuery,
    DiagnosticResolutionStatus,
    parseDiagnosticQuery,
    replaceInDiagnosticQuery,
    isInDiagnosticQuery,
} from './diagnosticQuery'

describe('parseDiagnosticQuery', () => {
    test('empty', () => expect(parseDiagnosticQuery('')).toEqual({ input: '' }))
    test('whitespace', () => expect(parseDiagnosticQuery('  ')).toEqual({ input: '  ' }))
    test('text', () => expect(parseDiagnosticQuery('a')).toEqual({ input: 'a', message: 'a' }))
    test('type', () => expect(parseDiagnosticQuery('type:a')).toEqual({ input: 'type:a', type: 'a' }))
    test('tag', () => expect(parseDiagnosticQuery('tag:a')).toEqual({ input: 'tag:a', tag: ['a'] }))
    test('unresolved', () =>
        expect(parseDiagnosticQuery('is:unresolved')).toEqual({
            input: 'is:unresolved',
            status: DiagnosticResolutionStatus.Unresolved,
        }))
    test('pending', () =>
        expect(parseDiagnosticQuery('is:pending')).toEqual({
            input: 'is:pending',
            status: DiagnosticResolutionStatus.PendingResolution,
        }))
    test('both statuses', () =>
        expect(parseDiagnosticQuery('is:pending is:unresolved')).toEqual({
            input: 'is:pending is:unresolved',
            status: undefined,
        }))
    test('repo', () =>
        expect(parseDiagnosticQuery('repo:a')).toEqual({ input: 'repo:a', document: [{ pattern: 'git://a/**' }] }))
    test('combined', () =>
        expect(parseDiagnosticQuery('a repo:b tag:c d type:e is:pending repo:f tag:g')).toEqual({
            input: 'a repo:b tag:c d type:e is:pending repo:f tag:g',
            message: 'a d',
            type: 'e',
            document: [{ pattern: 'git://b/**' }, { pattern: 'git://f/**' }],
            tag: ['c', 'g'],
            status: DiagnosticResolutionStatus.PendingResolution,
        }))
})

describe('appendToDiagnosticQuery', () => {
    test('repo', () => expect(appendToDiagnosticQuery('a repo:b', 'repo:', 'c')).toBe('a repo:b repo:c'))
})

describe('replaceInDiagnosticQuery', () => {
    test('is:', () => expect(replaceInDiagnosticQuery('a is:unresolved', 'is:', 'resolved')).toBe('a is:resolved'))
})

describe('isInDiagnosticQuery', () => {
    test('true', () => expect(isInDiagnosticQuery('a is:unresolved', 'is:', 'unresolved')).toBe(true))
    test('false', () => expect(isInDiagnosticQuery('a repo:b', 'repo:', 'c')).toBe(false))
})
