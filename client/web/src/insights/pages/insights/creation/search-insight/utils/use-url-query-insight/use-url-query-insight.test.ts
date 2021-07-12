import { getInsightDataFromQuery } from './use-url-query-insight'

describe('getInsightDataFromQuery', () => {
    it('should return null with empty query params', () => {
        const result = getInsightDataFromQuery('')

        expect(result).toStrictEqual({
            repositories: [],
            seriesQuery: '',
        })
    })

    describe('should return correct insight values ', () => {
        it('with repo: and "test" pattern query', () => {
            const queryString = 'context:global test repo:^github\\.com/sourcegraph/sourcegraph$  patterntype:literal'

            const result = getInsightDataFromQuery(queryString)

            expect(result).toStrictEqual({
                repositories: ['^github\\.com/sourcegraph/sourcegraph$'],
                seriesQuery: 'context:global test patterntype:literal',
            })
        })

        it('with multiple repo: filters and "test" pattern query', () => {
            const queryString =
                'context:global test repo:^github\\.com/sourcegraph/sourcegraph$ repo:^github\\.com/sourcegraph/about patterntype:literal'

            const result = getInsightDataFromQuery(queryString)

            expect(result).toStrictEqual({
                repositories: ['^github\\.com/sourcegraph/sourcegraph$', '^github\\.com/sourcegraph/about'],
                seriesQuery: 'context:global test patterntype:literal',
            })
        })

        it('with multiple repo: filters regexp literals and "test" pattern query', () => {
            const queryString =
                'context:global test repo:^github\\.com/sourcegraph/sourcegraph$|^github\\.com/sourcegraph/about patterntype:literal'

            const result = getInsightDataFromQuery(queryString)

            expect(result).toStrictEqual({
                repositories: ['^github\\.com/sourcegraph/sourcegraph$|^github\\.com/sourcegraph/about'],
                seriesQuery: 'context:global test patterntype:literal',
            })
        })

        it('with repo: filters and "repo: " pattern query', () => {
            const queryString =
                'context:global "repo: " repo:^github\\.com/sourcegraph/sourcegraph$ patterntype:literal'

            const result = getInsightDataFromQuery(queryString)

            expect(result).toStrictEqual({
                repositories: ['^github\\.com/sourcegraph/sourcegraph$'],
                seriesQuery: 'context:global "repo: " patterntype:literal',
            })
        })
    })
})
