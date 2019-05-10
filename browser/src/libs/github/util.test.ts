import { parseURL } from './util'

describe('util', () => {
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
            {
                name: 'snippet permalink',
                url:
                    'https://github.com/sourcegraph/sourcegraph/blob/6a91ccec97a46bfb511b7ff58d790554a7d075c8/client/browser/src/shared/repo/backend.tsx#L128-L151',
            },
            {
                name: 'pull request list',
                url: 'https://github.com/sourcegraph/sourcegraph/pulls',
            },
            {
                name: 'wiki page',
                url: 'https://github.com/sourcegraph/sourcegraph/pulls',
            },
            {
                name: 'branch name with forward slashes',
                url: 'http://ghe.sgdev.org/beyang/mux/blob/jr/branch/mux.go',
            },
        ]
        for (const { name, url } of testcases) {
            test(name, () => {
                expect(parseURL(new URL(url))).toMatchSnapshot()
            })
        }
    })
})
