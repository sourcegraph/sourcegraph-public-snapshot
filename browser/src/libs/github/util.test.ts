import { parseGitHubHash, parseURL } from './util'

describe('util', () => {
    describe('parseGitHubHash()', () => {
        test('parses nonexistent', () => expect(parseGitHubHash('')).toBe(undefined))
        test('parses empty', () => expect(parseGitHubHash('#')).toBe(undefined))
        test('parses single line', () =>
            expect(parseGitHubHash('#L123')).toEqual({ startLine: 123, endLine: undefined }))
        test('parses range', () => expect(parseGitHubHash('#L123-L456')).toEqual({ startLine: 123, endLine: 456 }))
        test('handles invalid value', () => expect(parseGitHubHash('#Lfoo')).toBe(undefined))
        test('allows extra after', () =>
            expect(parseGitHubHash('#L123-L456-foo')).toEqual({ startLine: 123, endLine: 456 }))
    })

    describe('parseURL()', () => {
        const testcases: {
            name: string
            url: string
        }[] = [
            {
                name: 'tree page',
                url: 'https://github.com/sourcegraph/sourcegraph/tree/master/client',
            },
            {
                name: 'blob page',
                url: 'https://github.com/sourcegraph/sourcegraph/blob/3.3/shared/src/hover/HoverOverlay.tsx',
            },
            {
                name: 'commit page',
                url: 'https://github.com/sourcegraph/sourcegraph/commit/fb054666c12b40f180c794db4829cbfd1a5fabae',
            },
            {
                name: 'pull request page',
                url: 'https://github.com/sourcegraph/sourcegraph/pull/3849',
            },
            {
                name: 'compare page',
                url:
                    'https://github.com/sourcegraph/sourcegraph-basic-code-intel/compare/new-extension-api-usage...fuzzy-locations',
            },
            {
                name: 'selections - single line',
                url: 'https://github.com/sourcegraph/sourcegraph/blob/master/jest.config.base.js#L5',
            },
            {
                name: 'selections - range',
                url: 'https://github.com/sourcegraph/sourcegraph/blob/master/jest.config.base.js#L5-L12',
            },
        ]
        for (const { name, url } of testcases) {
            test(name, () => {
                expect(parseURL(new URL(url))).toMatchSnapshot()
            })
        }
    })
})
