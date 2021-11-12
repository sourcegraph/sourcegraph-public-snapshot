import { ApolloCache, ApolloClient, gql } from '@apollo/client'
import { from, Observable, of, throwError } from 'rxjs'
import { map, mapTo, switchMap } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { UpdateLineChartSearchInsightInput } from '@sourcegraph/shared/src/graphql-operations'
import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'
import {
    CreateDashboardResult,
    CreateInsightResult,
    CreateInsightsDashboardInput,
    DeleteDashboardResult,
    GetDashboardInsightsResult,
    GetInsightsResult,
    GetInsightViewResult,
    InsightsDashboardsResult,
    LineChartSearchInsightDataSeriesInput,
    LineChartSearchInsightInput,
    UpdateDashboardResult,
    UpdateInsightsDashboardInput,
    UpdateLineChartSearchInsightResult,
} from '@sourcegraph/web/src/graphql-operations'

import { Insight, InsightDashboard, InsightsDashboardType, isSearchBasedInsight, SearchBasedInsight } from '../types'
import {
    isSearchBackendBasedInsight,
    SearchBackendBasedInsight,
    SearchBasedBackendFilters,
} from '../types/insight/search-insight'
import { SupportedInsightSubject } from '../types/subjects'

import { InsightStillProcessingError } from './api/get-backend-insight'
import { getBuiltInInsight } from './api/get-built-in-insight'
import { getLangStatsInsightContent } from './api/get-lang-stats-insight-content'
import { getRepositorySuggestions } from './api/get-repository-suggestions'
import { getResolvedSearchRepositories } from './api/get-resolved-search-repositories'
import { getSearchInsightContent } from './api/get-search-insight-content/get-search-insight-content'
import { CodeInsightsBackend } from './code-insights-backend'
import {
    BackendInsightData,
    DashboardCreateInput,
    DashboardDeleteInput,
    DashboardUpdateInput,
    FindInsightByNameInput,
    GetLangStatsInsightContentInput,
    GetSearchInsightContentInput,
    InsightCreateInput,
    InsightUpdateInput,
    ReachableInsight,
} from './code-insights-backend-types'
import { GET_DASHBOARD_INSIGHTS_GQL } from './gql/GetDashboardInsights'
import { GET_INSIGHTS_GQL } from './gql/GetInsights'
import { GET_INSIGHTS_DASHBOARDS_GQL } from './gql/GetInsightsDashboards'
import { GET_INSIGHT_VIEW_GQL } from './gql/GetInsightView'
import { createLineChartContent } from './utils/create-line-chart-content'
import { createDashboardGrants } from './utils/get-dashboard-grants'
import { getInsightView, getStepInterval } from './utils/insight-transformers'
import { parseDashboardType } from './utils/parse-dashboard-type'

export class CodeInsightsGqlBackend implements CodeInsightsBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    // Insights
    public getInsights = (dashboardId: string): Observable<Insight[]> => {
        // Handle virtual dashboard that doesn't exist in BE gql API and cause of that
        // we need to use here insightViews query to fetch all available insights
        if (dashboardId === 'all') {
            return fromObservableQuery(
                this.apolloClient.watchQuery<GetInsightsResult>({
                    query: GET_INSIGHTS_GQL,
                })
            ).pipe(map(({ data }) => data.insightViews.nodes.map(getInsightView).filter(Boolean) as Insight[]))
        }

        // Get all insights from the user-created dashboard
        return fromObservableQuery(
            this.apolloClient.watchQuery<GetDashboardInsightsResult>({
                query: GET_DASHBOARD_INSIGHTS_GQL,
                variables: { id: dashboardId },
            })
        ).pipe(
            map(
                ({ data }) =>
                    (data.insightsDashboards.nodes[0].views?.nodes.map(getInsightView).filter(Boolean) as Insight[]) ??
                    []
            )
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

                return getInsightView(insightData) || null
            })
        )

    public findInsightByName = (input: FindInsightByNameInput): Observable<Insight | null> =>
        this.getInsights('all').pipe(map(insights => insights.find(insight => insight.title === input.name) || null))

    public getReachableInsights = (): Observable<ReachableInsight[]> =>
        this.getInsights('all').pipe(
            map(insights =>
                insights.map(insight => ({
                    ...insight,
                    owner: {
                        id: '',
                        name: '',
                    },
                }))
            )
        )

    // TODO: Rethink all of this method. Currently `createViewContent` expects a different format of
    // the `Insight` type than we use elsewhere. This is a temporary solution to make the code
    // fit with both of those shapes.
    public getBackendInsightData = (insight: SearchBackendBasedInsight): Observable<BackendInsightData> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<GetInsightViewResult>({
                query: GET_INSIGHT_VIEW_GQL,
                variables: { id: insight.id },
                // In order to avoid unnecessary requests and enable caching for BE insights
                fetchPolicy: 'cache-first',
            })
        ).pipe(
            // Note: this insight is guaranteed to exist since this function
            // is only called from within a loop of insight ids
            map(({ data }) => data.insightViews.nodes[0]),
            switchMap(data => {
                if (!data) {
                    return throwError(new InsightStillProcessingError())
                }

                return of(data)
            }),
            map(data => ({
                id: insight.id,
                view: {
                    title: insight.title ?? insight.title,
                    // TODO: is this still used anywhere?
                    subtitle: '',
                    content: [createLineChartContent({ series: data.dataSeries }, insight.series)],
                    isFetchingHistoricalData: data.dataSeries.some(
                        ({ status: { pendingJobs, backfillQueuedAt } }) => pendingJobs > 0 || backfillQueuedAt === null
                    ),
                },
            }))
        )

    public getBuiltInInsightData = getBuiltInInsight

    // We don't have insight visibility and subject levels in the new GQL API anymore.
    // it was part of setting-cascade based API.
    public getInsightSubjects = (): Observable<SupportedInsightSubject[]> => of([])

    public createInsight = (input: InsightCreateInput): Observable<unknown> => {
        const { insight, dashboard } = input

        if (isSearchBasedInsight(insight)) {
            const input: LineChartSearchInsightInput = this.prepareSearchInsightCreateInput(insight, dashboard)

            return from(
                this.apolloClient.mutate<CreateInsightResult>({
                    mutation: gql`
                        mutation CreateInsight($input: LineChartSearchInsightInput!) {
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

        // TODO [VK] implement lang stats chart creation
        return of()
    }

    public updateInsight = (input: InsightUpdateInput): Observable<void[]> => {
        // Extracting mutations here to make it easier to support different types of insights
        const updateLineChartSearchInsightMutation = gql`
            mutation UpdateLineChartSearchInsight($input: UpdateLineChartSearchInsightInput!, $id: ID!) {
                updateLineChartSearchInsight(input: $input, id: $id) {
                    view {
                        id
                    }
                }
            }
        `

        const insight = input.newInsight
        const oldInsight = input.oldInsight

        if (isSearchBasedInsight(insight)) {
            const input: UpdateLineChartSearchInsightInput = this.prepareSearchInsightUpdateInput(insight)

            return from(
                this.apolloClient.mutate<UpdateLineChartSearchInsightResult>({
                    mutation: updateLineChartSearchInsightMutation,
                    variables: { input, id: oldInsight.id },
                })
            ).pipe(mapTo([]))
        }

        return of()
    }

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
                update(cache: ApolloCache<DeleteDashboardResult>, result) {
                    const deletedInsightReference = cache.identify({ __typename: 'InsightView', id: insightId })

                    // Remove deleted insights from the apollo cache
                    cache.evict({ id: deletedInsightReference })
                },
            })
        )

    // Dashboards
    public getDashboards = (): Observable<InsightDashboard[]> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<InsightsDashboardsResult>({
                query: GET_INSIGHTS_DASHBOARDS_GQL,
            })
        ).pipe(
            map(({ data }) => [
                {
                    id: 'all',
                    type: InsightsDashboardType.All,
                },
                ...data.insightsDashboards.nodes.map(
                    (dashboard): InsightDashboard => ({
                        id: dashboard.id,
                        type: parseDashboardType(dashboard.grants),
                        title: dashboard.title,
                        insightIds: dashboard.views?.nodes.map(view => view.id),
                        grants: dashboard.grants,
                    })
                ),
            ])
        )

    public getDashboardById = (dashboardId?: string): Observable<InsightDashboard | undefined> =>
        this.getDashboards().pipe(map(dashboards => dashboards.find(({ id }) => id === dashboardId)))

    // This is only used to check for duplicate dashboards. Thi is not required for the new GQL API.
    // So we just return null to get the form to always accept.
    public findDashboardByName = (name: string): Observable<InsightDashboard | null> => of(null)

    public createDashboard = (input: DashboardCreateInput): Observable<void> => {
        if (!input.type) {
            throw new Error('`grants` are required to create a new dashboard')
        }

        const mappedInput: CreateInsightsDashboardInput = {
            title: input.name,
            grants: createDashboardGrants(input),
        }

        return from(
            this.apolloClient.mutate<CreateDashboardResult>({
                mutation: gql`
                    mutation CreateDashboard($input: CreateInsightsDashboardInput!) {
                        createInsightsDashboard(input: $input) {
                            dashboard {
                                id
                            }
                        }
                    }
                `,
                variables: { input: mappedInput },
            })
        ).pipe(mapTo(undefined))
    }

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

    public updateDashboard = ({ id, nextDashboardInput }: DashboardUpdateInput): Observable<void> => {
        if (!id) {
            throw new Error('`id` is required to update a dashboard')
        }

        if (!nextDashboardInput.type) {
            throw new Error('`grants` are required to update a dashboard')
        }

        const input: UpdateInsightsDashboardInput = {
            title: nextDashboardInput.name,
            grants: createDashboardGrants(nextDashboardInput),
        }

        return from(
            this.apolloClient.mutate<UpdateDashboardResult>({
                mutation: gql`
                    mutation UpdateDashboard($id: ID!, $input: UpdateInsightsDashboardInput!) {
                        updateInsightsDashboard(id: $id, input: $input) {
                            dashboard {
                                id
                            }
                        }
                    }
                `,
                variables: {
                    id,
                    input,
                },
            })
        ).pipe(mapTo(undefined))
    }

    // Live preview fetchers
    public getSearchInsightContent = <D extends keyof ViewContexts>(
        input: GetSearchInsightContentInput<D>
    ): Promise<LineChartContent<any, string>> => getSearchInsightContent(input.insight, input.options)

    public getLangStatsInsightContent = <D extends keyof ViewContexts>(
        input: GetLangStatsInsightContentInput<D>
    ): Promise<PieChartContent<any>> => getLangStatsInsightContent(input.insight, input.options)

    // Repositories API
    public getRepositorySuggestions = getRepositorySuggestions
    public getResolvedSearchRepositories = getResolvedSearchRepositories

    private prepareSearchInsightCreateInput(
        insight: SearchBasedInsight,
        dashboard: InsightDashboard | null
    ): LineChartSearchInsightInput {
        const repositories = !isSearchBackendBasedInsight(insight) ? insight.repositories : []

        const [unit, value] = getStepInterval(insight)
        const input: LineChartSearchInsightInput = {
            dataSeries: insight.series.map<LineChartSearchInsightDataSeriesInput>(series => ({
                query: series.query,
                options: {
                    label: series.name,
                    lineColor: series.stroke,
                },
                repositoryScope: { repositories },
                timeScope: { stepInterval: { unit, value } },
            })),
            options: { title: insight.title },
        }

        if (dashboard?.id) {
            input.dashboards = [dashboard.id]
        }
        return input
    }

    private prepareSearchInsightUpdateInput(
        insight: SearchBasedInsight & { filters?: SearchBasedBackendFilters }
    ): UpdateLineChartSearchInsightInput {
        const repositories = !isSearchBackendBasedInsight(insight) ? insight.repositories : []

        const [unit, value] = getStepInterval(insight)
        const input: UpdateLineChartSearchInsightInput = {
            dataSeries: insight.series.map<LineChartSearchInsightDataSeriesInput>(series => ({
                query: series.query,
                options: {
                    label: series.name,
                    lineColor: series.stroke,
                },
                repositoryScope: { repositories },
                timeScope: { stepInterval: { unit, value } },
            })),
            presentationOptions: {
                title: insight.title,
            },
            viewControls: {
                filters: {
                    includeRepoRegex: insight.filters?.includeRepoRegexp,
                    excludeRepoRegex: insight.filters?.excludeRepoRegexp,
                },
            },
        }
        return input
    }
}
