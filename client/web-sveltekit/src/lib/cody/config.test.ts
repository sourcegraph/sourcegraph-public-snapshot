import { describe, test, expect } from 'vitest'

import { testCases } from '@sourcegraph/cody-context-filters-test-dataset/dataset.json'

import { CodyContextFilters, getFiltersFromCodyContextFilters } from './config'

describe('CodyContextFilters', () => {
    test('invalid re2 regex', async () => {
        const regexWithLookahead = '\\d(?=\\D)' // not supported in RE2
        const result = await CodyContextFilters.safeParseAsync({ exclude: [{ repoNamePattern: regexWithLookahead }] })
        expect(result.success).toBe(false)
    })
})

describe('getFiltersFromCodyContextFilters', () => {
    for (const testCase of testCases) {
        test(testCase.name, async () => {
            const filters = testCase['cody.contextFilters']
            if (!filters) {
                return
            }

            const filter = getFiltersFromCodyContextFilters(await CodyContextFilters.parseAsync(filters))

            const gotRepos = testCase.repos.filter(repo => filter(repo.name))
            expect(gotRepos).toEqual(testCase.includedRepos)
        })
    }
})
