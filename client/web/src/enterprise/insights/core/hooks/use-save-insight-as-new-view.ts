import { type FetchResult, gql, useMutation } from '@apollo/client'

import type { SaveInsightAsNewViewResult, SaveInsightAsNewViewVariables } from '../../../../graphql-operations'
import { INSIGHT_VIEW_FRAGMENT } from '../backend/gql-backend/gql/GetInsights'
import { searchInsightCreationOptimisticUpdate } from '../backend/gql-backend/methods/create-insight/create-insight'
import { type BackendInsight, type InsightDashboard, type InsightFilters, isVirtualDashboard } from '../types'

export const SAVE_INSIGHT_AS_NEW_VIEW_GQL = gql`
    mutation SaveInsightAsNewView($input: SaveInsightAsNewViewInput!) {
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
                seriesDisplayOptions: filters.seriesDisplayOptions,
                filters: {
                    searchContexts: [filters.context],
                    excludeRepoRegex: filters.excludeRepoRegexp,
                    includeRepoRegex: filters.includeRepoRegexp,
                },
            },
            dashboard: dashboard ? (isVirtualDashboard(dashboard) ? null : dashboard.id) : null,
        },
    }
}

interface saveNewInsightViewVariables {
    insight: BackendInsight
    filters: InsightFilters
    title: string
    dashboard: InsightDashboard | null
}

interface useSaveInsightAsNewViewProps {
    /**
     * Dashboard in which we should include newly created insight optimistically after
     * it's created via save as new insight mutation.
     */
    dashboard: InsightDashboard | null
}

type UseSaveInsightAsNewViewResultTuple = [
    (variables: saveNewInsightViewVariables) => Promise<FetchResult<SaveInsightAsNewViewResult>>
]

export function useSaveInsightAsNewView(props: useSaveInsightAsNewViewProps): UseSaveInsightAsNewViewResultTuple {
    const { dashboard } = props

    const [saveInsightAsNewView] = useMutation<SaveInsightAsNewViewResult, SaveInsightAsNewViewVariables>(
        SAVE_INSIGHT_AS_NEW_VIEW_GQL,
        {
            update: (cache, result) => {
                const { data } = result

                if (!data) {
                    return
                }

                searchInsightCreationOptimisticUpdate(cache, data.saveInsightAsNewView.view, dashboard?.id ?? null)
            },
        }
    )

    return [
        variables => {
            const input = getSaveInsightAsNewViewGqlInput(variables)

            return saveInsightAsNewView({ variables: input })
        },
    ]
}
