import { isInBlocklist } from './isInBlocklist'

describe('isInBlocklist', () => {
    const rawRepoName = 'github.com/sourcegraph/sourcegraph'

    const blocklistFactory = (enabled?: boolean) => (content?: string) => ({ enabled, content })

    describe('enabled=false', () => {
        const disabled = blocklistFactory(false)
        it('always returns "false"', () => {
            expect(isInBlocklist(rawRepoName, disabled(undefined))).toBeFalsy()
            expect(isInBlocklist(rawRepoName, disabled(''))).toBeFalsy()
            expect(isInBlocklist(rawRepoName, disabled('\n  \n  \n'))).toBeFalsy()
            expect(isInBlocklist(rawRepoName, disabled(rawRepoName))).toBeFalsy()
            expect(isInBlocklist(rawRepoName, disabled('*'))).toBeFalsy()
            expect(isInBlocklist(rawRepoName, disabled('github.com/*'))).toBeFalsy()
            expect(isInBlocklist(rawRepoName, disabled('github.com/sourcegraph/*'))).toBeFalsy()
            expect(isInBlocklist(rawRepoName, disabled('github.com/sourcegraph/source'))).toBeFalsy()
            expect(isInBlocklist(rawRepoName, disabled('github.com/sourcegraph/sourcegraph$'))).toBeFalsy()
        })
    })
    describe('enabled=true', () => {
        const enabled = blocklistFactory(true)
        it('correctly handles empty blocklist', () => {
            expect(isInBlocklist(rawRepoName, enabled())).toBeFalsy()
            expect(isInBlocklist(rawRepoName, enabled(''))).toBeFalsy()
            expect(isInBlocklist(rawRepoName, enabled('\n  \n  \n'))).toBeFalsy()
        })
        it('handles exact match', () => {
            expect(isInBlocklist(rawRepoName, enabled(rawRepoName))).toBeTruthy()
        })

        it('handles pattern', () => {
            expect(isInBlocklist(rawRepoName, enabled('*'))).toBeTruthy()
            expect(isInBlocklist(rawRepoName, enabled('github.com/*'))).toBeTruthy()
            expect(isInBlocklist(rawRepoName, enabled('github.com/sourcegraph/*'))).toBeTruthy()
            expect(isInBlocklist(rawRepoName, enabled('github.com/sourcegraph/source'))).toBeFalsy()
            expect(isInBlocklist(rawRepoName, enabled('github.com/sourcegraph/sourcegraph$'))).toBeTruthy()
        })

        it('handles with [https://] prefix', () => {
            expect(isInBlocklist(rawRepoName, enabled('https://github.com/sourcegraph/sourcegraph'))).toBeTruthy()
            expect(isInBlocklist(rawRepoName, enabled('https://github.com/sourcegraph/*'))).toBeTruthy()
        })

        it('handles single multi-line blocklist', () => {
            expect(
                isInBlocklist(rawRepoName, enabled('github.com/somerepo/*\n\ngithub.com/sourcegraph/*\n'))
            ).toBeTruthy()
        })
    })
})
