import { useMemo } from 'react'

import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter, Token } from '@sourcegraph/shared/src/search/query/token'
import { dedupeWhitespace } from '@sourcegraph/shared/src/util/strings';

import { DEFAULT_ACTIVE_COLOR } from '../../components/form-color-input/FormColorInput'
import { createDefaultEditSeries } from '../../components/search-insight-creation-content/hooks/use-editable-series'
import { INITIAL_INSIGHT_VALUES } from '../../components/search-insight-creation-content/initial-insight-values'
import { CreateInsightFormFields } from '../../types'

/**
 * Type guard for repo: filter token.
 *
 * @param token - query parsed lexical token
 */
const isRepoFilter = (token: Token): token is Filter => token.type === 'filter' && token.field.value === 'repo'

/**
 * Generate repositories string value without special reg exp
 * characters and extra whitespaces.
 */
const getSanitizedRepositoriesString = (repositories: string[]): string =>
    repositories
        .map(repo =>
            repo
                // Remove special regexp characters like ^\$
                .replace(/(\^)|(\$)|(\\)/gi, '')
                // Remove whitespaces at the start and end
                .trim()
        )
        .join(', ')

export interface InsightData {
    repositories: string
    seriesQuery: string
}

/**
 * Return serialized value of repositories and query from the URL query params.
 *
 * @param query -- query param with possible value for the creation UI
 *
 * Exported for testing only.
 */
export function getInsightDataFromQuery(query: string | null): InsightData | null {
    const sequence = scanSearchQuery(query ?? '')

    if (!query || sequence.type === 'error') {
        return null
    }

    const tokens = Array.isArray(sequence.term) ? sequence.term : [sequence.term]
    const repositories = []

    /** Find all repo: filters and get their values for insight repositories field */
    for (const token of tokens) {
        if (isRepoFilter(token)) {
            const repoValue = token.value?.value

            if (repoValue) {
                /**
                 * Split repo value in order to support case with multiple repo values
                 * in repo: filter.
                 *
                 * Example repo:^github\.com/org/repo-1$ | ^github\.com/org/repo-2$
                 */
                repositories.push(...repoValue.split('|'))
            }
        }
    }

    /**
     * Generate a string query from tokens without repo: filters for the insight
     * query field.
     */
    const tokensWithoutRepoFilters = tokens.filter(token => !isRepoFilter(token))
    const humanReadableQueryString = stringHuman(tokensWithoutRepoFilters)

    return {
        seriesQuery: dedupeWhitespace(humanReadableQueryString),
        repositories: getSanitizedRepositoriesString(repositories),
    }
}

/**
 * Returns initial values for the search insight from query param 'insight-query'.
 *
 * This logic is trying to find a value for the data query field in a query param
 * and extract the repo: filters for the repositories field in a creation UI.
 */
export function useUrlQueryInsight(queryParameters: string): CreateInsightFormFields | null {
    return useMemo(() => {
        const queryParametersString = new URLSearchParams(queryParameters).get('insight-query')

        const insightData = getInsightDataFromQuery(queryParametersString)

        if (insightData === null) {
            return null
        }

        return {
            ...INITIAL_INSIGHT_VALUES,
            series: [
                createDefaultEditSeries({
                    edit: true,
                    valid: true,
                    name: 'Search series #1',
                    query: insightData.seriesQuery,
                    stroke: DEFAULT_ACTIVE_COLOR,
                }),
            ],
            repositories: insightData.repositories,
        }
    }, [queryParameters])
}
