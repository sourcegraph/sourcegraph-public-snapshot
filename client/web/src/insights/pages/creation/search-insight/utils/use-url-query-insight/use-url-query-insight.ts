import { useContext, useEffect, useState } from 'react'

import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { isRepoFilter } from '@sourcegraph/shared/src/search/query/validate'
import { ErrorLike, asError } from '@sourcegraph/shared/src/util/errors'
import { dedupeWhitespace } from '@sourcegraph/shared/src/util/strings'

import { InsightsApiContext } from '../../../../../core/backend/api-provider'
import { createDefaultEditSeries } from '../../components/search-insight-creation-content/hooks/use-editable-series'
import { INITIAL_INSIGHT_VALUES } from '../../components/search-insight-creation-content/initial-insight-values'
import { CreateInsightFormFields } from '../../types'

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
export function getInsightDataFromQuery(searchQuery: string): InsightData {
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
    const tokensWithoutRepoFilters = tokens.filter(token => !isRepoFilter(token))
    const humanReadableQueryString = stringHuman(tokensWithoutRepoFilters)

    return {
        seriesQuery: dedupeWhitespace(humanReadableQueryString),
        repositories,
    }
}

function createInsightFormFields(seriesQuery: string, repositories: string[] = []): CreateInsightFormFields {
    return {
        ...INITIAL_INSIGHT_VALUES,
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

export interface UseURLQueryInsightResult {
    /**
     * Insight data. undefined in case if we are in a loading state or
     * URL doesn't have query param.
     * */
    data: CreateInsightFormFields | ErrorLike | undefined

    /** Whether the search query  param is presented in URL. */
    hasQueryInsight: boolean
}

/**
 * Returns initial values for the search insight from query param.
 */
export function useURLQueryInsight(queryParameters: string): UseURLQueryInsightResult {
    const { getResolvedSearchRepositories } = useContext(InsightsApiContext)
    const [insightFormFields, setInsightFormFields] = useState<CreateInsightFormFields | ErrorLike | undefined>()

    const query = new URLSearchParams(queryParameters).get('query')

    useEffect(() => {
        if (query === null) {
            return
        }

        const { seriesQuery, repositories } = getInsightDataFromQuery(query)

        // If search query doesn't have repo we should run async repositories resolve
        // step to avoid case then run search with query without repo: filter we get
        // all indexed repositories.
        if (repositories.length > 0) {
            getResolvedSearchRepositories(query)
                .then(repositories => setInsightFormFields(createInsightFormFields(seriesQuery, repositories)))
                .catch(error => setInsightFormFields(asError(error)))
        } else {
            setInsightFormFields(createInsightFormFields(seriesQuery, repositories))
        }
    }, [getResolvedSearchRepositories, query])

    return {
        hasQueryInsight: query !== null,
        data: query !== null ? insightFormFields : undefined,
    }
}
