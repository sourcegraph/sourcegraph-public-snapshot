import { ApolloCache, ApolloClient, ApolloQueryResult, gql } from '@apollo/client'
import { from, Observable, of } from 'rxjs'
import { catchError, map, mapTo, switchMap } from 'rxjs/operators'
import {
    AddInsightViewToDashboardResult,
    DeleteDashboardResult,
    ExampleFirstRepositoryResult,
    ExampleTodoRepositoryResult,
    GetAccessibleInsightsListResult,
    GetDashboardInsightsResult,
    GetFrozenInsightsCountResult,
    GetInsightsResult,
    HasAvailableCodeInsightResult,
    RemoveInsightViewFromDashboardResult,
    RemoveInsightViewFromDashboardVariables,
} from 'src/graphql-operations'

import { fromObservableQuery } from '@sourcegraph/http-client'

import { ALL_INSIGHTS_DASHBOARD } from '../../constants'
import { BackendInsight, Insight, InsightDashboard, InsightsDashboardOwner } from '../../types'
import { CodeInsightsBackend } from '../code-insights-backend'
import {
    AccessibleInsightInfo,
    AssignInsightsToDashboardInput,
    BackendInsightData,
    DashboardCreateInput,
    DashboardDeleteInput,
    DashboardUpdateInput,
    DashboardUpdateResult,
    GetLangStatsInsightContentInput,
    GetSearchInsightContentInput,
    InsightCreateInput,
    InsightUpdateInput,
    RemoveInsightFromDashboardInput,
    CategoricalChartContent,
    SeriesChartContent,
    UiFeaturesConfig,
    DashboardCreateResult,
    InsightPreviewSettings,
} from '../code-insights-backend-types'
import { getRepositorySuggestions } from '../core/api/get-repository-suggestions'
import { getResolvedSearchRepositories } from '../core/api/get-resolved-search-repositories'

import { createInsightView } from './deserialization/create-insight-view'
import { GET_ACCESSIBLE_INSIGHTS_LIST } from './gql/GetAccessibleInsightsList'
import { GET_DASHBOARD_INSIGHTS_GQL } from './gql/GetDashboardInsights'
import { GET_EXAMPLE_FIRST_REPOSITORY_GQL, GET_EXAMPLE_TODO_REPOSITORY_GQL } from './gql/GetExampleRepository'
import { GET_INSIGHTS_GQL } from './gql/GetInsights'
import { REMOVE_INSIGHT_FROM_DASHBOARD_GQL } from './gql/RemoveInsightFromDashboard'
import { createDashboard } from './methods/create-dashboard/create-dashboard'
import { createInsight } from './methods/create-insight/create-insight'
import { getBackendInsightData } from './methods/get-backend-insight-data/get-backend-insight-data'
import { getBuiltInInsight } from './methods/get-built-in-insight-data'
import { getLangStatsInsightContent } from './methods/get-built-in-insight-data/get-lang-stats-insight-content'
import { getSearchInsightContent } from './methods/get-built-in-insight-data/get-search-insight-content'
import { getDashboardOwners } from './methods/get-dashboard-owners'
import { getDashboardById } from './methods/get-dashboards/get-dashboard-by-id'
import { getDashboards } from './methods/get-dashboards/get-dashboards'
import { getInsightsPreview } from './methods/get-insight-preview'
import { updateDashboard } from './methods/update-dashboard'
import { updateInsight } from './methods/update-insight/update-insight'

export class CodeInsightsGqlBackend implements CodeInsightsBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    // Insights
    public getInsights = (input: { dashboardId: string }): Observable<Insight[]> => {
        const { dashboardId } = input

        // Handle virtual dashboard that doesn't exist in BE gql API and cause of that
        // we need to use here insightViews query to fetch all available insights
        if (dashboardId === ALL_INSIGHTS_DASHBOARD.id) {
            return fromObservableQuery(
                this.apolloClient.watchQuery<GetInsightsResult>({ query: GET_INSIGHTS_GQL })
            ).pipe(map(({ data }) => data.insightViews.nodes.map(createInsightView)))
        }

        // Get all insights from the user-created dashboard
        return fromObservableQuery(
            this.apolloClient.watchQuery<GetDashboardInsightsResult>({
                query: GET_DASHBOARD_INSIGHTS_GQL,
                // Prevent unnecessary network request after mutation over dashboard or insights within
                // current dashboard
                nextFetchPolicy: 'cache-first',
                variables: { id: dashboardId },
            })
        ).pipe(
            map(({ data }) => data.insightsDashboards.nodes[0]),
            map(dashboard => dashboard.views?.nodes.map(createInsightView) ?? [])
        )
    }

    public getInsightById = (id: string): Observable<Insight | null> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<GetInsightsResult>({
                query: GET_INSIGHTS_GQL,
                variables: { id },
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

    // TODO: This method is used only for insight title validation but since we don't have
    // limitations about title field in gql api remove this method and async validation for
    // title field as soon as setting-based api will be deprecated
    public findInsightByName = (): Observable<Insight | null> => of(null)

    public getAccessibleInsightsList = (): Observable<AccessibleInsightInfo[]> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<GetAccessibleInsightsListResult>({
                query: GET_ACCESSIBLE_INSIGHTS_LIST,
            })
        ).pipe(
            map(response =>
                response.data.insightViews.nodes.map(view => ({
                    id: view.id,
                    title: view.presentation.title,
                }))
            )
        )

    public getBackendInsightData = (insight: BackendInsight): Observable<BackendInsightData> =>
        getBackendInsightData(this.apolloClient, insight)

    public getBuiltInInsightData = getBuiltInInsight

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

    // Dashboards
    public getDashboards = (id?: string): Observable<InsightDashboard[]> => getDashboards(this.apolloClient, id)

    public getDashboardById = (input: { dashboardId: string | undefined }): Observable<InsightDashboard | null> =>
        getDashboardById(this.apolloClient, input)

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

    // Live preview fetchers
    public getSearchInsightContent = (input: GetSearchInsightContentInput): Promise<SeriesChartContent<any>> =>
        getSearchInsightContent(input).then(data => data.content)

    public getLangStatsInsightContent = (
        input: GetLangStatsInsightContentInput
    ): Promise<CategoricalChartContent<any>> => getLangStatsInsightContent(input).then(data => data.content)

    public getInsightPreviewContent = (input: InsightPreviewSettings): Promise<SeriesChartContent<any>> =>
        getInsightsPreview(this.apolloClient, input)

    // Repositories API
    public getRepositorySuggestions = getRepositorySuggestions
    public getResolvedSearchRepositories = getResolvedSearchRepositories

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

    public readonly UIFeatures: UiFeaturesConfig = {
        licensed: true,
        insightsLimit: null,
    }
}

const getRepositoryName = (
    result: ApolloQueryResult<ExampleTodoRepositoryResult | ExampleFirstRepositoryResult>
): string => result.data.search?.results.repositories[0]?.name || ''
