import { parseGitHubHash } from './util'

describe('util', () => {
    describe('parseGitHubHash()', () => {
        it('parses nonexistent', () => expect(parseGitHubHash('')).toBe(undefined))
        it('parses empty', () => expect(parseGitHubHash('#')).toBe(undefined))
        it('parses single line', () => expect(parseGitHubHash('#L123')).toEqual({ startLine: 123, endLine: undefined }))
        it('parses range', () => expect(parseGitHubHash('#L123-L456')).toEqual({ startLine: 123, endLine: 456 }))
        it('handles invalid value', () => expect(parseGitHubHash('#Lfoo')).toBe(undefined))
        it('allows extra after', () =>
            expect(parseGitHubHash('#L123-L456-foo')).toEqual({ startLine: 123, endLine: 456 }))
    })
})
