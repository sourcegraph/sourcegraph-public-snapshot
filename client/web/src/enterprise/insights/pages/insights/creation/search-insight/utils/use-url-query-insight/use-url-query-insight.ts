import { useEffect, useMemo, useState } from 'react'

import { gql, useLazyQuery } from '@apollo/client'

import { asError, ErrorLike, dedupeWhitespace } from '@sourcegraph/common'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { isFilterType, isRepoFilter } from '@sourcegraph/shared/src/search/query/validate'

import { SearchRepositoriesResult, SearchRepositoriesVariables } from '../../../../../../../../graphql-operations'
import { createDefaultEditSeries } from '../../../../../../components'
import { CreateInsightFormFields } from '../../types'

const GET_SEARCH_REPOSITORIES = gql`
    query SearchRepositories($query: String) {
        search(query: $query) {
            results {
                repositories {
                    name
                }
            }
        }
    }
`

export interface UseURLQueryInsightResult {
    /**
     * Insight data. undefined in case if we are in a loading state or
     * URL doesn't have query param.
     * */
    data: Partial<CreateInsightFormFields> | ErrorLike | undefined

    /** Whether the search query  param is presented in URL. */
    hasQueryInsight: boolean
}

/**
 * Returns initial values for the search insight from query param.
 */
export function useURLQueryInsight(queryParameters: string): UseURLQueryInsightResult {
    const [insightValues, setInsightValues] = useState<Partial<CreateInsightFormFields> | ErrorLike | undefined>()

    const [getResolvedSearchRepositories] = useLazyQuery<SearchRepositoriesResult, SearchRepositoriesVariables>(
        GET_SEARCH_REPOSITORIES
    )
    const query = useMemo(() => new URLSearchParams(queryParameters).get('query'), [queryParameters])

    useEffect(() => {
        if (query === null) {
            return
        }

        const { seriesQuery, repositories } = getInsightDataFromQuery(query)

        // If search query doesn't have repo we should run async repositories resolve
        // step to avoid case then run search with query without repo: filter we get
        // all indexed repositories.
        if (repositories.length > 0) {
            getResolvedSearchRepositories({ variables: { query } })
                .then(({ data }) => {
                    const repositories = data?.search?.results.repositories ?? []
                    setInsightValues(
                        createInsightFormFields(
                            seriesQuery,
                            repositories.map(repo => repo.name)
                        )
                    )
                })
                .catch(error => setInsightValues(asError(error)))
        } else {
            setInsightValues(createInsightFormFields(seriesQuery, repositories))
        }
    }, [getResolvedSearchRepositories, query])

    return {
        hasQueryInsight: query !== null,
        data: query !== null ? insightValues : undefined,
    }
}

export interface InsightData {
    repositories: string[]
    seriesQuery: string
}

/**
 * Return serialized value of repositories and query from the URL query params.
 *
 * @param searchQuery -- query param with possible value for the creation UI
 *
 * Exported for testing only.
 */
export function getInsightDataFromQuery(searchQuery: string | null): InsightData {
    const sequence = scanSearchQuery(searchQuery ?? '')

    if (!searchQuery || sequence.type === 'error') {
        return {
            seriesQuery: '',
            repositories: [],
        }
    }

    const tokens = Array.isArray(sequence.term) ? sequence.term : [sequence.term]
    const repositories = []

    // Find all repo: filters and get their values for insight repositories field
    for (const token of tokens) {
        if (isRepoFilter(token)) {
            const repoValue = token.value?.value

            if (repoValue) {
                repositories.push(repoValue)
            }
        }
    }

    // Generate a string query from tokens without repo: filters for the insight
    // query field.
    const tokensWithoutRepoFiltersAndContext = tokens.filter(
        token => !isRepoFilter(token) && !isFilterType(token, FilterType.context)
    )

    const humanReadableQueryString = stringHuman(tokensWithoutRepoFiltersAndContext)

    return {
        seriesQuery: dedupeWhitespace(humanReadableQueryString.trim()),
        repositories,
    }
}

function createInsightFormFields(seriesQuery: string, repositories: string[] = []): Partial<CreateInsightFormFields> {
    return {
        series: [
            createDefaultEditSeries({
                edit: true,
                valid: true,
                name: 'Search series #1',
                query: seriesQuery ?? '',
            }),
        ],
        repositories: repositories.join(', '),
    }
}
