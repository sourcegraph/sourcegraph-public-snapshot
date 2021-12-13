import { ApolloCache, ApolloClient, gql } from '@apollo/client'
import { from, Observable, of } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import {
    CreateLangStatsInsightResult,
    CreateSearchBasedInsightResult,
    FirstStepCreateSearchBasedInsightResult,
    PieChartSearchInsightInput,
    UpdateLineChartSearchInsightResult,
    UpdateLineChartSearchInsightVariables,
} from '../../../../../../../graphql-operations'
import {
    CaptureGroupInsight,
    InsightDashboard,
    InsightExecutionType,
    InsightType,
    isVirtualDashboard,
    SearchBasedInsight,
} from '../../../../types'
import { SearchBackendBasedInsight } from '../../../../types/insight/search-insight'
import { InsightCreateInput } from '../../../code-insights-backend-types'
import { createInsightView } from '../../deserialization/create-insight-view'
import { INSIGHT_VIEW_FRAGMENT } from '../../gql/GetInsights'
import { getSearchInsightUpdateInput } from '../update-insight/serializators'

import { getInsightCreateGqlInput, getLangStatsInsightCreateInput } from './serializators'

/**
 * Main handler to create insight with GQL api. It absorbs all implementation details around GQL api.
 */
export const createInsight = (apolloClient: ApolloClient<object>, input: InsightCreateInput): Observable<unknown> => {
    const { insight, dashboard } = input

    switch (insight.viewType) {
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

function createSearchBasedInsight(
    apolloClient: ApolloClient<object>,
    insight: SearchBasedInsight | CaptureGroupInsight,
    dashboard: InsightDashboard | null
): Observable<unknown> {
    const input = getInsightCreateGqlInput(insight, dashboard)

    // In case if we want to create an insight with some predefined filters we have to
    // create the insight first and only after update this newly created insight with filter values
    // This is due to lack of gql API flexibility and should be fixed as soon as BE gql API
    // supports filters in the create insight mutation.
    // TODO: Remove this imperative logic as soon as be supports filters
    if (insight.type === InsightExecutionType.Backend && insight.filters) {
        const filters = insight.filters
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

                const createdInsight = createInsightView(
                    data.createLineChartSearchInsight.view
                ) as SearchBackendBasedInsight

                const input = getSearchInsightUpdateInput({ ...createdInsight, filters })

                return apolloClient.mutate<UpdateLineChartSearchInsightResult, UpdateLineChartSearchInsightVariables>({
                    mutation: gql`
                        mutation UpdateLineChartSearchInsight($input: UpdateLineChartSearchInsightInput!, $id: ID!) {
                            updateLineChartSearchInsight(input: $input, id: $id) {
                                view {
                                    ...InsightViewNode
                                }
                            }
                        }
                        ${INSIGHT_VIEW_FRAGMENT}
                    `,
                    variables: { input, id: createdInsight.id },
                    update(cache, result) {
                        const { data } = result

                        if (!data) {
                            return
                        }

                        searchInsightCreationOptimisticUpdate(cache, data, dashboard)
                    },
                })
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
    cache: ApolloCache<object>,
    data: UpdateLineChartSearchInsightResult,
    dashboard: InsightDashboard | null
): void {
    const createdInsightId = cache.identify({
        __typename: 'InsightView',
        id: data.updateLineChartSearchInsight.view.id,
    })

    if (dashboard && !isVirtualDashboard(dashboard)) {
        const Insight = cache.readFragment<any>({
            id: createdInsightId,
            fragment: INSIGHT_VIEW_FRAGMENT,
        })

        const dashboardReference = cache.identify({
            __typename: 'InsightsDashboard',
            id: dashboard?.id ?? '',
        })

        const cachedDashboard = cache.readFragment<any>({
            id: dashboardReference,
            fragmentName: 'DashboardFragment',
            fragment: gql`
                fragment DashboardFragment on InsightsDashboard {
                    id
                    title
                    views {
                        nodes {
                            ...InsightViewNode
                        }
                    }
                    grants {
                        users
                        organizations
                        global
                    }
                }
                ${INSIGHT_VIEW_FRAGMENT}
            `,
        })

        cache.writeFragment({
            id: dashboardReference,
            fragmentName: 'DashboardFragment',
            fragment: gql`
                fragment DashboardFragment on InsightsDashboard {
                    id
                    title
                    views {
                        nodes {
                            ...InsightViewNode
                        }
                    }
                    grants {
                        users
                        organizations
                        global
                    }
                }
                ${INSIGHT_VIEW_FRAGMENT}
            `,
            data: {
                ...cachedDashboard,
                views: { nodes: [...cachedDashboard.views.nodes, Insight] },
            },
        })
    }
}
