import { isInBlocklist } from './isInBlocklist'

describe('isInBlocklist', () => {
    const rawRepoName = 'github.com/sourcegraph/sourcegraph'

    it('correctly handles empty blocklist', () => {
        expect(isInBlocklist('', rawRepoName)).toBeFalsy()
        expect(isInBlocklist('\n  \n  \n', rawRepoName)).toBeFalsy()
    })
    it('handles exact match', () => {
        expect(isInBlocklist('github.com/sourcegraph/sourcegraph', 'github.com/sourcegraph/sourcegraph')).toBeTruthy()
    })

    it('handles pattern', () => {
        expect(isInBlocklist('*', 'github.com/sourcegraph/sourcegraph')).toBeTruthy()
        expect(isInBlocklist('github.com/*', 'github.com/sourcegraph/sourcegraph')).toBeTruthy()
        expect(isInBlocklist('github.com/sourcegraph/*', 'github.com/sourcegraph/sourcegraph')).toBeTruthy()
        expect(isInBlocklist('github.com/sourcegraph/source', 'github.com/sourcegraph/sourcegraph')).toBeFalsy()
        expect(isInBlocklist('github.com/sourcegraph/sourcegraph$', 'github.com/sourcegraph/sourcegraph')).toBeTruthy()
    })

    it('handles with [https://] prefix', () => {
        expect(
            isInBlocklist('https://github.com/sourcegraph/sourcegraph', 'github.com/sourcegraph/sourcegraph')
        ).toBeTruthy()
        expect(isInBlocklist('https://github.com/sourcegraph/*', 'github.com/sourcegraph/sourcegraph')).toBeTruthy()
    })

    it('handles single multi-line blocklist', () => {
        expect(
            isInBlocklist('github.com/somerepo/*\n\ngithub.com/sourcegraph/*\n', 'github.com/sourcegraph/sourcegraph')
        ).toBeTruthy()
    })
})
