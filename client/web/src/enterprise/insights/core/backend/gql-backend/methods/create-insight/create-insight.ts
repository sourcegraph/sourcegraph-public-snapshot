import { type ApolloCache, type ApolloClient, gql } from '@apollo/client'
import { from, type Observable } from 'rxjs'

import type {
    CreateLangStatsInsightResult,
    CreateSearchBasedInsightResult,
    GetDashboardInsightsResult,
    GetDashboardInsightsVariables,
    InsightViewNode,
    PieChartSearchInsightInput,
} from '../../../../../../../graphql-operations'
import { InsightType } from '../../../../types'
import type {
    InsightCreateInput,
    MinimalCaptureGroupInsightData,
    MinimalComputeInsightData,
    MinimalSearchBasedInsightData,
} from '../../../code-insights-backend-types'
import { GET_DASHBOARD_INSIGHTS_GQL } from '../../gql/GetDashboardInsights'
import { INSIGHT_VIEW_FRAGMENT } from '../../gql/GetInsights'

import { getInsightCreateGqlInput, getLangStatsInsightCreateInput } from './serializators'

/**
 * Main handler to create insight with GQL api. It absorbs all implementation details around GQL api.
 */
export const createInsight = (apolloClient: ApolloClient<object>, input: InsightCreateInput): Observable<unknown> => {
    const { insight, dashboardId } = input

    switch (insight.type) {
        case InsightType.CaptureGroup:
        case InsightType.Compute:
        case InsightType.SearchBased: {
            return createSearchBasedInsight(apolloClient, insight, dashboardId)
        }

        case InsightType.LangStats: {
            return from(
                apolloClient.mutate<CreateLangStatsInsightResult, { input: PieChartSearchInsightInput }>({
                    mutation: gql`
                        mutation CreateLangStatsInsight($input: PieChartSearchInsightInput!) {
                            createPieChartSearchInsight(input: $input) {
                                view {
                                    id
                                }
                            }
                        }
                    `,
                    variables: { input: getLangStatsInsightCreateInput(insight, dashboardId) },
                })
            )
        }
    }
}

type CreationSeriesInsightData =
    | MinimalSearchBasedInsightData
    | MinimalCaptureGroupInsightData
    | MinimalComputeInsightData

function createSearchBasedInsight(
    apolloClient: ApolloClient<object>,
    insight: CreationSeriesInsightData,
    dashboardId: string | null
): Observable<unknown> {
    const input = getInsightCreateGqlInput(insight, dashboardId)

    return from(
        apolloClient.mutate<CreateSearchBasedInsightResult>({
            mutation: gql`
                mutation CreateSearchBasedInsight($input: LineChartSearchInsightInput!) {
                    createLineChartSearchInsight(input: $input) {
                        view {
                            ...InsightViewNode
                        }
                    }
                }
                ${INSIGHT_VIEW_FRAGMENT}
            `,
            variables: { input },
            update: (cache, result) => {
                const { data } = result

                if (!data) {
                    return
                }

                searchInsightCreationOptimisticUpdate(cache, data.createLineChartSearchInsight.view, dashboardId)
            },
        })
    )
}

/**
 * Updates Apollo caches after insight creation. Add insight to main insights gql query,
 * add newly created insight to the cache dashboard that insight was created from.
 */
export function searchInsightCreationOptimisticUpdate(
    cache: ApolloCache<unknown>,
    createdView: InsightViewNode,
    dashboardId: string | null
): void {
    if (dashboardId) {
        const cachedDashboardQuery = cache.readQuery<GetDashboardInsightsResult, GetDashboardInsightsVariables>({
            query: GET_DASHBOARD_INSIGHTS_GQL,
            variables: { id: dashboardId },
        })

        if (!cachedDashboardQuery) {
            return
        }

        const cachedDashboard = cachedDashboardQuery.insightsDashboards.nodes[0]
        const cachedDashboardInsights = [...(cachedDashboard.views?.nodes ?? [])]
        const updatedDashboard = {
            ...cachedDashboard,
            views: {
                ...cachedDashboard.views,
                nodes: [...cachedDashboardInsights, createdView],
            },
        }

        cache.writeQuery<GetDashboardInsightsResult>({
            query: GET_DASHBOARD_INSIGHTS_GQL,
            variables: { id: dashboardId },
            data: {
                insightsDashboards: {
                    ...cachedDashboardQuery.insightsDashboards,
                    nodes: [updatedDashboard],
                },
            },
        })
    }
}
