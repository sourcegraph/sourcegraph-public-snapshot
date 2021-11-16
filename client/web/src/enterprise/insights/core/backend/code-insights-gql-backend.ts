import { ApolloCache, ApolloClient, gql } from '@apollo/client'
import { from, Observable, of, throwError } from 'rxjs'
import { map, mapTo, switchMap } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { UpdateLineChartSearchInsightInput } from '@sourcegraph/shared/src/graphql-operations'
import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'
import {
    AddInsightViewToDashboardResult,
    CreateDashboardResult,
    CreateInsightsDashboardInput,
    CreateLangStatsInsightResult,
    CreateSearchBasedInsightResult,
    DeleteDashboardResult,
    GetDashboardInsightsResult,
    GetInsightsResult,
    GetInsightViewResult,
    InsightsDashboardsResult,
    InsightSubjectsResult,
    LineChartSearchInsightInput,
    PieChartSearchInsightInput,
    RemoveInsightViewFromDashboardResult,
    UpdateDashboardResult,
    UpdateInsightsDashboardInput,
    UpdateLangStatsInsightResult,
    UpdateLangStatsInsightVariables,
    UpdateLineChartSearchInsightResult,
} from '@sourcegraph/web/src/graphql-operations'

import { Insight, InsightDashboard, InsightsDashboardScope, InsightsDashboardType, InsightType } from '../types'
import { ALL_INSIGHTS_DASHBOARD_ID } from '../types/dashboard/virtual-dashboard'
import { SearchBackendBasedInsight } from '../types/insight/search-insight'
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
    DashboardCreateResult,
    DashboardDeleteInput,
    DashboardUpdateInput,
    DashboardUpdateResult,
    FindInsightByNameInput,
    GetLangStatsInsightContentInput,
    GetSearchInsightContentInput,
    InsightCreateInput,
    InsightUpdateInput,
    ReachableInsight,
} from './code-insights-backend-types'
import { GET_DASHBOARD_INSIGHTS_GQL } from './gql/GetDashboardInsights'
import { GET_INSIGHTS_GQL, INSIGHT_VIEW_FRAGMENT } from './gql/GetInsights'
import { GET_INSIGHTS_DASHBOARDS_GQL } from './gql/GetInsightsDashboards'
import { GET_INSIGHTS_SUBJECTS_GQL } from './gql/GetInsightSubjects'
import { GET_INSIGHT_VIEW_GQL } from './gql/GetInsightView'
import { createLineChartContent } from './utils/create-line-chart-content'
import { createDashboardGrants } from './utils/get-dashboard-grants'
import { getInsightView } from './utils/insight-transformers'
import { parseDashboardScope } from './utils/parse-dashboard-scope'
import { prepareSearchInsightCreateInput, prepareSearchInsightUpdateInput } from './utils/search-insight-to-gql-input'

export class CodeInsightsGqlBackend implements CodeInsightsBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    // Insights
    public getInsights = (dashboardId: string): Observable<Insight[]> => {
        // Handle virtual dashboard that doesn't exist in BE gql API and cause of that
        // we need to use here insightViews query to fetch all available insights
        if (dashboardId === ALL_INSIGHTS_DASHBOARD_ID) {
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

    // TODO: This method is used only for insight title validation but since we don't have
    // limitations about title field in gql api remove this method and async validation for
    // title field as soon as setting-based api will be deprecated
    public findInsightByName = (input: FindInsightByNameInput): Observable<Insight | null> => of(null)

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

        switch (insight.viewType) {
            case InsightType.SearchBased: {
                const input: LineChartSearchInsightInput = prepareSearchInsightCreateInput(insight, dashboard)

                return from(
                    this.apolloClient.mutate<CreateSearchBasedInsightResult>({
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

            case InsightType.LangStats: {
                return from(
                    this.apolloClient.mutate<CreateLangStatsInsightResult, { input: PieChartSearchInsightInput }>({
                        mutation: gql`
                            mutation CreateLangStatsInsight($input: PieChartSearchInsightInput!) {
                                createPieChartSearchInsight(input: $input) {
                                    view {
                                        id
                                    }
                                }
                            }
                        `,
                        variables: {
                            input: {
                                query: '',
                                repositoryScope: { repositories: [insight.repository] },
                                presentationOptions: {
                                    title: insight.title,
                                    otherThreshold: insight.otherThreshold,
                                },
                                dashboards: [dashboard?.id ?? ''],
                            },
                        },
                    })
                )
            }
        }
    }

    public updateInsight = (input: InsightUpdateInput): Observable<void[]> => {
        const insight = input.newInsight
        const oldInsight = input.oldInsight

        switch (insight.viewType) {
            case InsightType.SearchBased: {
                const updateLineChartSearchInsightMutation = gql`
                    mutation UpdateLineChartSearchInsight($input: UpdateLineChartSearchInsightInput!, $id: ID!) {
                        updateLineChartSearchInsight(input: $input, id: $id) {
                            view {
                                ...InsightViewNode
                            }
                        }
                    }
                    ${INSIGHT_VIEW_FRAGMENT}
                `

                const input: UpdateLineChartSearchInsightInput = prepareSearchInsightUpdateInput(insight)

                return from(
                    this.apolloClient.mutate<UpdateLineChartSearchInsightResult>({
                        mutation: updateLineChartSearchInsightMutation,
                        variables: { input, id: oldInsight.id },
                    })
                ).pipe(mapTo([]))
            }
            case InsightType.LangStats: {
                return from(
                    this.apolloClient.mutate<UpdateLangStatsInsightResult, UpdateLangStatsInsightVariables>({
                        mutation: gql`
                            mutation UpdateLangStatsInsight($id: ID!, $input: UpdatePieChartSearchInsightInput!) {
                                updatePieChartSearchInsight(id: $id, input: $input) {
                                    view {
                                        ...InsightViewNode
                                    }
                                }
                            }
                            ${INSIGHT_VIEW_FRAGMENT}
                        `,
                        variables: {
                            id: oldInsight.id,
                            input: {
                                query: '',
                                repositoryScope: { repositories: [insight.repository] },
                                presentationOptions: {
                                    title: insight.title,
                                    otherThreshold: insight.otherThreshold,
                                },
                            },
                        },
                    })
                ).pipe(mapTo([]))
            }
        }
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
                update(cache: ApolloCache<DeleteDashboardResult>) {
                    const deletedInsightReference = cache.identify({ __typename: 'InsightView', id: insightId })

                    // Remove deleted insights from the apollo cache
                    cache.evict({ id: deletedInsightReference })
                },
            })
        )

    // Dashboards
    public getDashboards = (id?: string): Observable<InsightDashboard[]> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<InsightsDashboardsResult>({
                query: GET_INSIGHTS_DASHBOARDS_GQL,
                variables: { id },
            })
        ).pipe(
            map(result => {
                const { data } = result

                return [
                    {
                        id: 'all',
                        type: InsightsDashboardType.Virtual,
                        scope: InsightsDashboardScope.Personal,
                        title: 'All Insights',
                    },
                    ...data.insightsDashboards.nodes.map(
                        (dashboard): InsightDashboard => ({
                            id: dashboard.id,
                            type: InsightsDashboardType.Custom,
                            scope: parseDashboardScope(dashboard.grants),
                            title: dashboard.title,
                            insightIds: dashboard.views?.nodes.map(view => view.id),
                            grants: dashboard.grants,

                            // BE gql dashboards don't have setting key (it's setting cascade conception only)
                            settingsKey: null,
                        })
                    ),
                ]
            })
        )

    public getDashboardById = (dashboardId?: string): Observable<InsightDashboard | null> =>
        this.getDashboards(dashboardId).pipe(map(dashboards => dashboards.find(({ id }) => id === dashboardId) ?? null))

    // This is only used to check for duplicate dashboards. Thi is not required for the new GQL API.
    // So we just return null to get the form to always accept.
    public findDashboardByName = (name: string): Observable<InsightDashboard | null> => of(null)

    public getDashboardSubjects = (): Observable<SupportedInsightSubject[]> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<InsightSubjectsResult>({ query: GET_INSIGHTS_SUBJECTS_GQL })
        ).pipe(
            map(({ data }) => {
                const { currentUser, site } = data

                if (!currentUser) {
                    return []
                }

                return [{ ...currentUser }, ...currentUser.organizations.nodes, site]
            })
        )

    public createDashboard = (input: DashboardCreateInput): Observable<DashboardCreateResult> => {
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
                                title
                                views {
                                    nodes {
                                        id
                                    }
                                }
                                grants {
                                    users
                                    organizations
                                    global
                                }
                            }
                        }
                    }
                `,
                variables: { input: mappedInput },
                update(cache, result) {
                    const { data } = result

                    if (!data) {
                        return
                    }

                    cache.modify({
                        fields: {
                            insightsDashboards(dashboards) {
                                const newDashboardsReference = cache.writeFragment({
                                    data: data.createInsightsDashboard.dashboard,
                                    fragment: gql`
                                        fragment NewTodo on InsightsDashboard {
                                            id
                                            title
                                            views {
                                                nodes {
                                                    id
                                                }
                                            }
                                            grants {
                                                users
                                                organizations
                                                global
                                            }
                                        }
                                    `,
                                })

                                return { nodes: [...(dashboards.nodes ?? []), newDashboardsReference] }
                            },
                        },
                    })
                },
            })
        ).pipe(map(result => ({ id: result.data?.createInsightsDashboard.dashboard.id ?? 'unknown' })))
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

    public updateDashboard = ({
        previousDashboard,
        nextDashboardInput,
    }: DashboardUpdateInput): Observable<DashboardUpdateResult> => {
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
                                title
                                grants {
                                    users
                                    organizations
                                    global
                                }
                            }
                        }
                    }
                `,
                variables: {
                    id: previousDashboard.id,
                    input,
                },
            })
        ).pipe(
            map(result => {
                const { data } = result

                if (!data?.updateInsightsDashboard) {
                    throw new Error('The dashboard update was not successful')
                }

                return data.updateInsightsDashboard.dashboard
            })
        )
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

    public assignInsightsToDashboard = ({
        id,
        nextDashboardInput,
        previousDashboard,
    }: DashboardUpdateInput): Observable<unknown> => {
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

        const removeInsightViewFromDashboard = (insightViewId: string, dashboardId: string): Promise<any> =>
            this.apolloClient.mutate<RemoveInsightViewFromDashboardResult>({
                mutation: gql`
                    mutation RemoveInsightViewFromDashboard($insightViewId: ID!, $dashboardId: ID!) {
                        removeInsightViewFromDashboard(
                            input: { insightViewId: $insightViewId, dashboardId: $dashboardId }
                        ) {
                            dashboard {
                                id
                            }
                        }
                    }
                `,
                variables: { insightViewId, dashboardId },
            })

        const addedInsightIds =
            nextDashboardInput.insightIds?.filter(insightId => !previousDashboard.insightIds?.includes(insightId)) || []

        // Get array of removed insight view ids
        const removedInsightIds =
            previousDashboard.insightIds?.filter(insightId => !nextDashboardInput.insightIds?.includes(insightId)) || []

        return from(
            Promise.all([
                ...addedInsightIds.map(insightId => addInsightViewToDashboard(insightId, id || '')),
                ...removedInsightIds.map(insightId => removeInsightViewFromDashboard(insightId, id || '')),
            ])
        ).pipe(
            // Next query is needed to update local apollo cache and re-trigger getInsights query.
            // Usually Apollo does that under the hood by itself based on response from a mutation
            // but in this case since we don't have one single query to assign/unassign insights
            // from dashboard we have to call query manually.
            switchMap(() =>
                this.apolloClient.query<GetDashboardInsightsResult>({
                    query: GET_DASHBOARD_INSIGHTS_GQL,
                    variables: { id: id ?? '' },
                })
            )
        )
    }
}
