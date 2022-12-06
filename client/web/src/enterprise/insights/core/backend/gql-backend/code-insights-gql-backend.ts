import { ApolloCache, ApolloClient, ApolloQueryResult, gql } from '@apollo/client'
import { from, Observable, of } from 'rxjs'
import { catchError, map, mapTo, switchMap } from 'rxjs/operators'
import {
    AddInsightViewToDashboardResult,
    DeleteDashboardResult,
    ExampleFirstRepositoryResult,
    ExampleTodoRepositoryResult,
    GetDashboardInsightsResult,
    GetFrozenInsightsCountResult,
    GetInsightsResult,
    HasAvailableCodeInsightResult,
    RemoveInsightViewFromDashboardResult,
    RemoveInsightViewFromDashboardVariables,
} from 'src/graphql-operations'

import { isDefined } from '@sourcegraph/common'
import { fromObservableQuery } from '@sourcegraph/http-client'

import { ALL_INSIGHTS_DASHBOARD } from '../../../constants'
import { Insight, InsightDashboard, InsightsDashboardOwner, isComputeInsight } from '../../types'
import { CodeInsightsBackend } from '../code-insights-backend'
import {
    AssignInsightsToDashboardInput,
    DashboardCreateInput,
    DashboardDeleteInput,
    DashboardUpdateInput,
    DashboardUpdateResult,
    InsightCreateInput,
    InsightUpdateInput,
    RemoveInsightFromDashboardInput,
    DashboardCreateResult,
} from '../code-insights-backend-types'

import { createInsightView } from './deserialization/create-insight-view'
import { GET_DASHBOARD_INSIGHTS_GQL } from './gql/GetDashboardInsights'
import { GET_EXAMPLE_FIRST_REPOSITORY_GQL, GET_EXAMPLE_TODO_REPOSITORY_GQL } from './gql/GetExampleRepository'
import { GET_INSIGHTS_GQL } from './gql/GetInsights'
import { REMOVE_INSIGHT_FROM_DASHBOARD_GQL } from './gql/RemoveInsightFromDashboard'
import { createDashboard } from './methods/create-dashboard/create-dashboard'
import { createInsight } from './methods/create-insight/create-insight'
import { getDashboardOwners } from './methods/get-dashboard-owners'
import { updateDashboard } from './methods/update-dashboard'
import { updateInsight } from './methods/update-insight/update-insight'

export class CodeInsightsGqlBackend implements CodeInsightsBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    // Insights
    public getInsights = (input: { dashboardId: string; withCompute: boolean }): Observable<Insight[]> => {
        const { dashboardId, withCompute } = input

        // Handle virtual dashboard that doesn't exist in BE gql API and cause of that
        // we need to use here insightViews query to fetch all available insights
        if (dashboardId === ALL_INSIGHTS_DASHBOARD.id) {
            return fromObservableQuery(
                this.apolloClient.watchQuery<GetInsightsResult>({
                    query: GET_INSIGHTS_GQL,
                    // Prevent unnecessary network request after mutation over dashboard or insights within
                    // current dashboard
                    nextFetchPolicy: 'cache-first',
                    errorPolicy: 'all',
                })
            ).pipe(
                map(({ data }) => data.insightViews.nodes.filter(isDefined).map(createInsightView)),
                map(insights => (withCompute ? insights : insights.filter(insight => !isComputeInsight(insight))))
            )
        }

        // Get all insights from the user-created dashboard
        return fromObservableQuery(
            this.apolloClient.watchQuery<GetDashboardInsightsResult>({
                query: GET_DASHBOARD_INSIGHTS_GQL,
                // Prevent unnecessary network request after mutation over dashboard or insights within
                // current dashboard
                nextFetchPolicy: 'cache-first',
                errorPolicy: 'all',
                variables: { id: dashboardId },
            })
        ).pipe(
            map(({ data }) => data.insightsDashboards.nodes[0]),
            map(dashboard => dashboard.views?.nodes.filter(isDefined).map(createInsightView) ?? []),
            map(insights => (withCompute ? insights : insights.filter(insight => !isComputeInsight(insight))))
        )
    }

    public getInsightById = (id: string): Observable<Insight | null> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<GetInsightsResult>({
                query: GET_INSIGHTS_GQL,
                variables: { id },
                errorPolicy: 'all',
            })
        ).pipe(
            map(({ data }) => {
                const insightData = data.insightViews.nodes[0]

                if (!insightData) {
                    return null
                }

                return createInsightView(insightData)
            }),
            catchError(() => of(null))
        )

    public hasInsights = (first: number): Observable<boolean> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<HasAvailableCodeInsightResult>({
                query: gql`
                    query HasAvailableCodeInsight($first: Int!) {
                        insightViews(first: $first) {
                            nodes {
                                id
                            }
                        }
                    }
                `,
                variables: { first },
                nextFetchPolicy: 'cache-only',
            })
        ).pipe(map(({ data }) => data.insightViews.nodes.length === first))

    public getActiveInsightsCount = (first: number): Observable<number> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<GetFrozenInsightsCountResult>({
                query: gql`
                    query GetFrozenInsightsCount($first: Int!) {
                        insightViews(first: $first, isFrozen: false) {
                            nodes {
                                id
                            }
                        }
                    }
                `,
                variables: { first },
            })
        ).pipe(map(({ data }) => data.insightViews.nodes.length))

    public createInsight = (input: InsightCreateInput): Observable<unknown> => createInsight(this.apolloClient, input)

    public updateInsight = (input: InsightUpdateInput): Observable<unknown> => updateInsight(this.apolloClient, input)

    public deleteInsight = (insightId: string): Observable<unknown> =>
        from(
            this.apolloClient.mutate({
                mutation: gql`
                    mutation DeleteInsightView($id: ID!) {
                        deleteInsightView(id: $id) {
                            alwaysNil
                        }
                    }
                `,
                variables: { id: insightId },
                update(cache: ApolloCache<DeleteDashboardResult>) {
                    const deletedInsightReference = cache.identify({ __typename: 'InsightView', id: insightId })

                    // Remove deleted insights from the apollo cache
                    cache.evict({ id: deletedInsightReference })
                },
            })
        )

    public removeInsightFromDashboard = (input: RemoveInsightFromDashboardInput): Observable<unknown> => {
        const { insightId, dashboardId } = input

        return from(
            this.apolloClient.mutate<RemoveInsightViewFromDashboardResult, RemoveInsightViewFromDashboardVariables>({
                mutation: REMOVE_INSIGHT_FROM_DASHBOARD_GQL,
                variables: { insightId, dashboardId },
                update(cache: ApolloCache<RemoveInsightViewFromDashboardResult>) {
                    const deletedInsightReference = cache.identify({ __typename: 'InsightView', id: insightId })

                    // Remove deleted insights from the apollo cache
                    cache.evict({ id: deletedInsightReference })
                },
            })
        )
    }

    // This is only used to check for duplicate dashboards. Thi is not required for the new GQL API.
    // So we just return null to get the form to always accept.
    public findDashboardByName = (name: string): Observable<InsightDashboard | null> => of(null)

    public getDashboardOwners = (): Observable<InsightsDashboardOwner[]> => getDashboardOwners(this.apolloClient)

    public createDashboard = (input: DashboardCreateInput): Observable<DashboardCreateResult> =>
        createDashboard(this.apolloClient, input)

    public deleteDashboard = ({ id }: DashboardDeleteInput): Observable<void> => {
        if (!id) {
            throw new Error('`id` is required to delete a dashboard')
        }

        return from(
            this.apolloClient.mutate<DeleteDashboardResult>({
                mutation: gql`
                    mutation DeleteDashboard($id: ID!) {
                        deleteInsightsDashboard(id: $id) {
                            alwaysNil
                        }
                    }
                `,
                variables: { id },
            })
        ).pipe(mapTo(undefined))
    }

    public updateDashboard = (input: DashboardUpdateInput): Observable<DashboardUpdateResult> =>
        updateDashboard(this.apolloClient, input)

    public assignInsightsToDashboard = ({
        id,
        prevInsightIds,
        nextInsightIds,
    }: AssignInsightsToDashboardInput): Observable<unknown> => {
        const addInsightViewToDashboard = (insightViewId: string, dashboardId: string): Promise<any> =>
            this.apolloClient.mutate<AddInsightViewToDashboardResult>({
                mutation: gql`
                    mutation AddInsightViewToDashboard($insightViewId: ID!, $dashboardId: ID!) {
                        addInsightViewToDashboard(input: { insightViewId: $insightViewId, dashboardId: $dashboardId }) {
                            dashboard {
                                id
                            }
                        }
                    }
                `,
                variables: { insightViewId, dashboardId },
            })

        const removeInsightViewFromDashboard = (insightId: string, dashboardId: string): Promise<any> =>
            this.apolloClient.mutate<RemoveInsightViewFromDashboardResult, RemoveInsightViewFromDashboardVariables>({
                mutation: REMOVE_INSIGHT_FROM_DASHBOARD_GQL,
                variables: { insightId, dashboardId },
            })

        const addedInsightIds = nextInsightIds.filter(insightId => !prevInsightIds.includes(insightId)) || []

        // Get array of removed insight view ids
        const removedInsightIds = prevInsightIds.filter(insightId => !nextInsightIds.includes(insightId)) || []

        return from(
            Promise.all([
                ...addedInsightIds.map(insightId => addInsightViewToDashboard(insightId, id || '')),
                ...removedInsightIds.map(insightId => removeInsightViewFromDashboard(insightId, id || '')),
            ])
        ).pipe(
            // Next query is needed to update local apollo cache and re-trigger getInsights query.
            // Usually Apollo does that under the hood by itself based on response from a mutation
            // but in this case since we don't have one single query to assign/unassigned insights
            // from dashboard we have to call query manually.
            switchMap(() =>
                this.apolloClient.query<GetDashboardInsightsResult>({
                    query: GET_DASHBOARD_INSIGHTS_GQL,
                    variables: { id: id ?? '' },
                })
            )
        )
    }

    public getFirstExampleRepository = (): Observable<string> => {
        const firstRepository = (): Observable<string> =>
            fromObservableQuery(
                this.apolloClient.watchQuery<ExampleFirstRepositoryResult>({
                    query: GET_EXAMPLE_FIRST_REPOSITORY_GQL,
                })
            ).pipe(map(getRepositoryName))

        const todoRepository = (): Observable<string> =>
            fromObservableQuery(
                this.apolloClient.watchQuery<ExampleTodoRepositoryResult>({
                    query: GET_EXAMPLE_TODO_REPOSITORY_GQL,
                })
            ).pipe(map(getRepositoryName))

        return todoRepository().pipe(
            switchMap(todoRepository => (todoRepository ? of(todoRepository) : firstRepository()))
        )
    }
}

const getRepositoryName = (
    result: ApolloQueryResult<ExampleTodoRepositoryResult | ExampleFirstRepositoryResult>
): string => result.data.search?.results.repositories[0]?.name || ''
