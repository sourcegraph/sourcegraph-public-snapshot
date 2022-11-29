import { FetchResult, gql, useMutation } from '@apollo/client'

import { SaveInsightAsNewViewResult, SaveInsightAsNewViewVariables } from '../../../../graphql-operations'
import { parseSeriesDisplayOptions } from '../../components/insights-view-grid/components/backend-insight/components/drill-down-filters-panel/drill-down-filters/utils'
import { INSIGHT_VIEW_FRAGMENT } from '../backend/gql-backend/gql/GetInsights'
import { BackendInsight, InsightFilters } from '../types'

export const SAVE_INSIGHT_AS_NEW_VIEW_GQL = gql`
    mutation SaveInsightAsNewView($input: InsightNewViewInput!) {
        saveInsightAsNewView(input: $input) {
            view {
                ...InsightViewNode
            }
        }
    }
    ${INSIGHT_VIEW_FRAGMENT}
`

export function getSaveInsightAsNewViewGqlInput(input: saveNewInsightViewVariables): SaveInsightAsNewViewVariables {
    const { insight, filters, title, dashboard } = input
    return {
        input: {
            insightViewId: insight.id,
            options: { title },
            viewControls: {
                seriesDisplayOptions:
                    insight.seriesDisplayOptions || parseSeriesDisplayOptions(insight.appliedSeriesDisplayOptions),
                filters: {
                    searchContexts: [filters.context],
                    excludeRepoRegex: filters.excludeRepoRegexp,
                    includeRepoRegex: filters.includeRepoRegexp,
                },
            },
            dashboard,
        },
    }
}

interface saveNewInsightViewVariables {
    insight: BackendInsight
    filters: InsightFilters
    title: string
    dashboard: string | null
}

type useSaveInsightAsNewViewTuple = [
    (variables: saveNewInsightViewVariables) => Promise<FetchResult<SaveInsightAsNewViewResult>>
]

export function useSaveInsightAsNewView(): useSaveInsightAsNewViewTuple {
    const [saveInsightAsNewView] = useMutation<SaveInsightAsNewViewResult, SaveInsightAsNewViewVariables>(
        SAVE_INSIGHT_AS_NEW_VIEW_GQL
    )

    return [
        variables => {
            const input = getSaveInsightAsNewViewGqlInput(variables)

            return saveInsightAsNewView({ variables: input })
        },
    ]
}
