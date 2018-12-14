import assert from 'assert'
import { parseGitHubHash } from './util'

describe('parseGitHubHash', () => {
    it('parses nonexistent', () => assert.deepStrictEqual(parseGitHubHash(''), undefined))
    it('parses empty', () => assert.deepStrictEqual(parseGitHubHash('#'), undefined))
    it('parses single line', () =>
        assert.deepStrictEqual(parseGitHubHash('#L123'), { startLine: 123, endLine: undefined }))
    it('parses range', () => assert.deepStrictEqual(parseGitHubHash('#L123-L456'), { startLine: 123, endLine: 456 }))
    it('handles invalid value', () => assert.deepStrictEqual(parseGitHubHash('#Lfoo'), undefined))
    it('allows extra after', () =>
        assert.deepStrictEqual(parseGitHubHash('#L123-L456-foo'), { startLine: 123, endLine: 456 }))
})
