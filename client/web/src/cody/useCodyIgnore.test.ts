import { describe, expect, it } from 'vitest'

import { testCases } from '@sourcegraph/cody-context-filters-test-dataset/dataset.json'

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
