import { useMemo } from 'react'

import { DEFAULT_ACTIVE_COLOR } from '../../components/form-color-input/FormColorInput'
import { createDefaultEditSeries } from '../../components/search-insight-creation-content/hooks/use-editable-series'
import { INITIAL_INSIGHT_VALUES } from '../../components/search-insight-creation-content/initial-insight-values'
import { CreateInsightFormFields } from '../../types'

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
    if (!query) {
        return null
    }

    const queryWithoutRepo = query.replaceAll(/repo:.+?($|\s)/gi, '')
    const repos = query.match(/repo:.+?($|\s)/gi)

    return {
        seriesQuery: queryWithoutRepo.replace(/\s+/g, ' ').trim(),
        repositories:
            repos !== null
                ? repos
                      .map(repo =>
                          repo
                              .replace(/(repo:)|(\^)|(\$)|(\\)/gi, '')
                              .replace(/\s+/g, ' ')
                              .trim()
                      )
                      .join(', ')
                : '',
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
