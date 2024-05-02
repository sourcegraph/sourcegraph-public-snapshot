import { describe, expect, it } from 'vitest'

import { alwaysTrue, getFilterFnsFromCodyContextFilters } from './useCodyIgnore'

describe('getFilterFnsFromCodyContextFilters', () => {
    it('ignores everything if failed to parse filters from site config', async () => {
        const regexWithLookahead = '\\d(?=\\D)' // not supported in RE2
        const { isRepoIgnored, isFileIgnored } = await getFilterFnsFromCodyContextFilters({
            exclude: [{ repoNamePattern: regexWithLookahead }],
        })
        expect(isRepoIgnored).toBe(alwaysTrue)
        expect(isFileIgnored).toBe(alwaysTrue)
    })

    // TODO: (taras-yemets) replace with shared Cody Ignore test dataset once it's published
    const testCases = [
        {
            name: 'Cody context filters are not defined',
            description: 'Any repo should be included.',
            includeByDefault: true,
            includeUnknown: false,
            'cody.contextFilters': null,
            repos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
            ],
            includedRepos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
            ],
            fileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    Path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    Path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    Path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    Path: '/file4.go',
                },
            ],
            includedFileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    Path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    Path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    Path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    Path: '/file4.go',
                },
            ],
        },
        {
            name: 'Include and exclude rules are not defined',
            description:
                'This scenario shouldn\'t happen. "cody.contextFilters" if defined in the site config, should have at least one property. Thus, either "include" or "exclude" should be defined. We rely on site config schema validation.',
            includeByDefault: true,
            includeUnknown: false,
            'cody.contextFilters': {},
            repos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
            ],
            includedRepos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
            ],
            fileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
            ],
            includedFileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
            ],
        },
        {
            name: 'Include and exclude rules are empty lists',
            description:
                'This scenario shouldn\'t happen. If either "include" or "exclude" field is defined, it should have at least one item. We rely on site config schema validation.',
            includeByDefault: true,
            includeUnknown: false,
            'cody.contextFilters': {
                include: [],
                exclude: [],
            },
            repos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
            ],
            includedRepos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
            ],
            fileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
            ],
            includedFileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
            ],
        },
        {
            name: 'Only include rules are defined',
            description: 'Only repos matching "include" repo name patterns should be included.',
            includeByDefault: true,
            includeUnknown: false,
            'cody.contextFilters': {
                include: [{ repoNamePattern: '^github\\.com\\/sourcegraph\\/.+' }],
            },
            repos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
            ],
            includedRepos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
            ],
            fileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
            ],
            includedFileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
            ],
        },
        {
            name: 'Only exclude rules are defined',
            description: 'Only repos not matching any of "include" repo name patterns should be included.',
            includeByDefault: true,
            includeUnknown: false,
            'cody.contextFilters': {
                exclude: [{ repoNamePattern: '^github\\.com\\/sourcegraph\\/.+' }],
            },
            repos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
            ],
            includedRepos: [{ name: 'github.com/docker/compose', id: 4 }],
            fileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
            ],
            includedFileChunks: [
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
            ],
        },
        {
            name: 'Include and exclude rules are defined',
            description:
                'Only repos matching any of "include" and not matching any of "exclude" repo name patterns should be included.',
            includeByDefault: true,
            includeUnknown: false,
            'cody.contextFilters': {
                include: [{ repoNamePattern: '^github\\.com\\/sourcegraph\\/.+' }],
                exclude: [{ repoNamePattern: '.*cody.*' }],
            },
            repos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
                { name: 'github.com/sourcegraph/cody', id: 5 },
            ],
            includedRepos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
            ],
            fileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/cody',
                        id: 5,
                    },
                    path: '/index.ts',
                },
            ],
            includedFileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
            ],
        },
        {
            name: 'Multiple include and exclude rules are defined',
            description:
                'Only repos matching any of "include" and not matching any of "exclude" repo name patterns should be included.',
            includeByDefault: true,
            includeUnknown: false,
            'cody.contextFilters': {
                include: [
                    { repoNamePattern: '^github\\.com\\/sourcegraph\\/.+' },
                    { repoNamePattern: '^github\\.com\\/docker\\/compose$' },
                    { repoNamePattern: '^github\\.com\\/.+\\/react' },
                ],
                exclude: [{ repoNamePattern: '.*cody.*' }, { repoNamePattern: '.+\\/docker\\/.+' }],
            },
            repos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
                { name: 'github.com/sourcegraph/cody', id: 5 },
                { name: 'github.com/facebook/react', id: 6 },
            ],
            includedRepos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/facebook/react', id: 6 },
            ],
            fileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/cody',
                        id: 5,
                    },
                    path: '/index.ts',
                },
                {
                    repo: {
                        name: 'github.com/facebook/react',
                        id: 6,
                    },
                    path: '/hooks.ts',
                },
            ],
            includedFileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/facebook/react',
                        id: 6,
                    },
                    path: '/hooks.ts',
                },
            ],
        },
        {
            name: 'Include rules contain repo name pattern matching any repo',
            description: 'Any repo should be included.',
            includeByDefault: true,
            includeUnknown: false,
            'cody.contextFilters': {
                include: [{ repoNamePattern: '.*' }],
            },
            repos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
                { name: 'github.com/sourcegraph/cody', id: 5 },
                { name: 'github.com/facebook/react', id: 6 },
            ],
            includedRepos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
                { name: 'github.com/sourcegraph/cody', id: 5 },
                { name: 'github.com/facebook/react', id: 6 },
            ],
            fileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/cody',
                        id: 5,
                    },
                    path: '/index.ts',
                },
                {
                    repo: {
                        name: 'github.com/facebook/react',
                        id: 6,
                    },
                    path: '/hooks.ts',
                },
            ],
            includedFileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/cody',
                        id: 5,
                    },
                    path: '/index.ts',
                },
                {
                    repo: {
                        name: 'github.com/facebook/react',
                        id: 6,
                    },
                    path: '/hooks.ts',
                },
            ],
        },
        {
            name: 'Exclude rules contain repo name pattern matching any repo',
            description: 'Neither repo should be included.',
            includeByDefault: true,
            includeUnknown: false,
            'cody.contextFilters': {
                include: [{ repoNamePattern: '^github\\.com\\/sourcegraph\\/.+' }],
                exclude: [{ repoNamePattern: '.*' }],
            },
            repos: [
                { name: 'github.com/sourcegraph/about', id: 1 },
                { name: 'github.com/sourcegraph/annotate', id: 2 },
                { name: 'github.com/sourcegraph/sourcegraph', id: 3 },
                { name: 'github.com/docker/compose', id: 4 },
            ],
            includedRepos: [],
            fileChunks: [
                {
                    repo: {
                        name: 'github.com/sourcegraph/about',
                        id: 1,
                    },
                    path: '/file1.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/annotate',
                        id: 2,
                    },
                    path: '/file2.go',
                },
                {
                    repo: {
                        name: 'github.com/sourcegraph/sourcegraph',
                        id: 3,
                    },
                    path: '/file3.go',
                },
                {
                    repo: {
                        name: 'github.com/docker/compose',
                        id: 4,
                    },
                    path: '/file4.go',
                },
            ],
            includedFileChunks: [],
        },
    ]

    for (const testCase of testCases) {
        it(testCase.name, async () => {
            const ccf = testCase['cody.contextFilters']
            if (!ccf) {
                return
            }
            const { isRepoIgnored, isFileIgnored } = await getFilterFnsFromCodyContextFilters(ccf)

            const gotRepos = testCase.repos.filter(r => !isRepoIgnored(r.name))
            expect(gotRepos).toEqual(testCase.includedRepos)

            const gotFileChunks = testCase.fileChunks.filter(fc => !isFileIgnored(fc.repo.name, fc.path))
            expect(gotFileChunks).toEqual(testCase.includedFileChunks)
        })
    }
})
