import { ApolloCache, ApolloClient, gql } from '@apollo/client'
import { from, Observable, of } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import {
    CreateLangStatsInsightResult,
    CreateSearchBasedInsightResult,
    FirstStepCreateSearchBasedInsightResult,
    GetDashboardInsightsResult,
    GetDashboardInsightsVariables,
    InsightViewNode,
    PieChartSearchInsightInput,
} from '../../../../../../../graphql-operations'
import { InsightDashboard, InsightExecutionType, InsightType, isVirtualDashboard } from '../../../../types'
import {
    InsightCreateInput,
    MinimalCaptureGroupInsightData,
    MinimalSearchBasedInsightData,
} from '../../../code-insights-backend-types'
import { createInsightView } from '../../deserialization/create-insight-view'
import { GET_DASHBOARD_INSIGHTS_GQL } from '../../gql/GetDashboardInsights'
import { INSIGHT_VIEW_FRAGMENT } from '../../gql/GetInsights'
import { updateInsight, UpdateResult } from '../update-insight/update-insight'

import { getInsightCreateGqlInput, getLangStatsInsightCreateInput } from './serializators'

/**
 * Main handler to create insight with GQL api. It absorbs all implementation details around GQL api.
 */
export const createInsight = (apolloClient: ApolloClient<object>, input: InsightCreateInput): Observable<unknown> => {
    const { insight, dashboard } = input

    switch (insight.type) {
        case InsightType.CaptureGroup:
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

type CreationSeriesInsightData = MinimalSearchBasedInsightData | MinimalCaptureGroupInsightData

function createSearchBasedInsight(
    apolloClient: ApolloClient<object>,
    insight: CreationSeriesInsightData,
    dashboard: InsightDashboard | null
): Observable<unknown> {
    const input = getInsightCreateGqlInput(insight, dashboard)

    // In case if we want to create an insight with some predefined filters we have to
    // create the insight first and only after update this newly created insight with filter values
    // This is due to lack of gql API flexibility and should be fixed as soon as BE gql API
    // supports filters in the create insight mutation.
    // TODO: Remove this imperative logic as soon as be supports filters
    if (insight.executionType === InsightExecutionType.Backend && insight.filters) {
        return from(
            apolloClient.mutate<FirstStepCreateSearchBasedInsightResult>({
                mutation: gql`
                    mutation FirstStepCreateSearchBasedInsight($input: LineChartSearchInsightInput!) {
                        createLineChartSearchInsight(input: $input) {
                            view {
                                ...InsightViewNode
                            }
                        }
                    }
                    ${INSIGHT_VIEW_FRAGMENT}
                `,
                variables: { input },
            })
        ).pipe(
            switchMap(result => {
                const { data } = result

                if (!data) {
                    return of()
                }

                const createdInsight = {
                    ...createInsightView(data.createLineChartSearchInsight.view),
                    filters: insight.filters,
                }

                return updateInsight(
                    apolloClient,
                    { insightId: createdInsight.id, nextInsightData: createdInsight },
                    (cache, result) => {
                        const { data } = result

                        if (!data) {
                            return
                        }

                        searchInsightCreationOptimisticUpdate(cache, data, dashboard)
                    }
                )
            })
        )
    }

    return from(
        apolloClient.mutate<CreateSearchBasedInsightResult>({
            mutation: gql`
                mutation CreateSearchBasedInsight($input: LineChartSearchInsightInput!) {
                    createLineChartSearchInsight(input: $input) {
                        view {
                            id
                        }
                    }
                }
            `,
            variables: { input },
        })
    )
}

/**
 * Updates Apollo cache after insight creation. Add insight to main insights gql query,
 * add newly created insight to the cache dashboard that insight was crated from.
 */
function searchInsightCreationOptimisticUpdate(
    cache: ApolloCache<unknown>,
    data: UpdateResult,
    dashboard: InsightDashboard | null
): void {
    const createInsightIdRaw =
        'updateLineChartSearchInsight' in data
            ? data.updateLineChartSearchInsight.view.id
            : data.updatePieChartSearchInsight.view.id

    const createdInsightId = cache.identify({
        __typename: 'InsightView',
        id: createInsightIdRaw,
    })

    const cachedInsight = cache.readFragment<InsightViewNode>({
        id: createdInsightId,
        fragment: INSIGHT_VIEW_FRAGMENT,
        fragmentName: 'InsightViewNode',
    })

    if (dashboard && !isVirtualDashboard(dashboard)) {
        const cachedDashboardQuery = cache.readQuery<GetDashboardInsightsResult, GetDashboardInsightsVariables>({
            query: GET_DASHBOARD_INSIGHTS_GQL,
            variables: { id: dashboard.id },
        })

        if (!cachedDashboardQuery || !cachedInsight) {
            return
        }

        const cachedDashboard = cachedDashboardQuery.insightsDashboards.nodes[0]
        const cachedDashboardInsights = [...(cachedDashboard.views?.nodes ?? [])]
        const updatedDashboard = {
            ...cachedDashboard,
            views: {
                ...cachedDashboard.views,
                nodes: [...cachedDashboardInsights, cachedInsight],
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
