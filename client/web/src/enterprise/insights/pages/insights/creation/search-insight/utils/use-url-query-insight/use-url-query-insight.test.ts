import { describe, expect, it } from '@jest/globals'

import { getInsightDataFromQuery } from './use-url-query-insight'

describe('getInsightDataFromQuery', () => {
    it('should return null with empty query params', () => {
        const result = getInsightDataFromQuery('')

        expect(result).toStrictEqual({
            repoQuery: '',
            seriesQuery: '',
        })
    })

    describe('should return correct insight values ', () => {
        it('removes context from the query', () => {
            const queryString = 'context:global test repo:^github\\.com/sourcegraph/sourcegraph$  patterntype:literal'

            const result = getInsightDataFromQuery(queryString)

            expect(result).toStrictEqual({
                repoQuery: 'repo:^github\\.com/sourcegraph/sourcegraph$',
                seriesQuery: 'test patterntype:literal',
            })
        })

        it('with repo: and "test" pattern query', () => {
            const queryString = 'test repo:^github\\.com/sourcegraph/sourcegraph$  patterntype:literal'

            const result = getInsightDataFromQuery(queryString)

            expect(result).toStrictEqual({
                repoQuery: 'repo:^github\\.com/sourcegraph/sourcegraph$',
                seriesQuery: 'test patterntype:literal',
            })
        })

        it('with multiple repo: filters and "test" pattern query', () => {
            const queryString =
                'test repo:^github\\.com/sourcegraph/sourcegraph$ repo:^github\\.com/sourcegraph/about patterntype:literal'

            const result = getInsightDataFromQuery(queryString)

            expect(result).toStrictEqual({
                repoQuery: 'repo:^github\\.com/sourcegraph/sourcegraph$ repo:^github\\.com/sourcegraph/about',
                seriesQuery: 'test patterntype:literal',
            })
        })

        it('with multiple repo: filters regexp literals and "test" pattern query', () => {
            const queryString =
                'context:global test repo:^github\\.com/sourcegraph/sourcegraph$|^github\\.com/sourcegraph/about patterntype:literal'

            const result = getInsightDataFromQuery(queryString)

            expect(result).toStrictEqual({
                repoQuery: 'repo:^github\\.com/sourcegraph/sourcegraph$|^github\\.com/sourcegraph/about',
                seriesQuery: 'test patterntype:literal',
            })
        })

        it('with repo: filters and "repo: " pattern query', () => {
            const queryString =
                'context:global "repo: " repo:^github\\.com/sourcegraph/sourcegraph$ patterntype:literal'

            const result = getInsightDataFromQuery(queryString)

            expect(result).toStrictEqual({
                repoQuery: 'repo:^github\\.com/sourcegraph/sourcegraph$',
                seriesQuery: '"repo: " patterntype:literal',
            })
        })
    })
})
