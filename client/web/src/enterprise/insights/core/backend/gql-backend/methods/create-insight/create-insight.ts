import { ApolloCache, ApolloClient, gql } from '@apollo/client'
import { from, Observable } from 'rxjs'

import {
    CreateLangStatsInsightResult,
    CreateSearchBasedInsightResult,
    GetDashboardInsightsResult,
    GetDashboardInsightsVariables,
    InsightViewNode,
    PieChartSearchInsightInput,
} from '../../../../../../../graphql-operations'
import { InsightDashboard, InsightType, isVirtualDashboard } from '../../../../types'
import {
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
    const { insight, dashboard } = input

    switch (insight.type) {
        case InsightType.CaptureGroup:
        case InsightType.Compute:
        case InsightType.SearchBased: {
            return createSearchBasedInsight(apolloClient, insight, dashboard)
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
                    variables: { input: getLangStatsInsightCreateInput(insight, dashboard) },
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
    dashboard: InsightDashboard | null
): Observable<unknown> {
    const input = getInsightCreateGqlInput(insight, dashboard)

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

                searchInsightCreationOptimisticUpdate(cache, data.createLineChartSearchInsight.view, dashboard)
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
    dashboard: InsightDashboard | null
): void {
    if (dashboard && !isVirtualDashboard(dashboard)) {
        const cachedDashboardQuery = cache.readQuery<GetDashboardInsightsResult, GetDashboardInsightsVariables>({
            query: GET_DASHBOARD_INSIGHTS_GQL,
            variables: { id: dashboard.id },
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
            variables: { id: dashboard.id },
            data: {
                insightsDashboards: {
                    ...cachedDashboardQuery.insightsDashboards,
                    nodes: [updatedDashboard],
                },
            },
        })
    }
}
