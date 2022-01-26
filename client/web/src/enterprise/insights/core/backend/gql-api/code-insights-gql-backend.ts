import { ApolloCache, ApolloClient, gql } from '@apollo/client'
import { from, Observable, of } from 'rxjs'
import { map, mapTo, switchMap } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'
import {
    AddInsightViewToDashboardResult,
    CreateDashboardResult,
    CreateInsightsDashboardInput,
    DeleteDashboardResult,
    GetDashboardInsightsResult,
    GetInsightsResult,
    InsightsDashboardsResult,
    InsightSubjectsResult,
    RemoveInsightViewFromDashboardResult,
    UpdateDashboardResult,
    UpdateInsightsDashboardInput,
} from 'src/graphql-operations'

import { fromObservableQuery } from '@sourcegraph/http-client'
import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

import { BackendInsight, Insight, InsightDashboard, InsightsDashboardScope, InsightsDashboardType } from '../../types'
import { ALL_INSIGHTS_DASHBOARD_ID } from '../../types/dashboard/virtual-dashboard'
import { SupportedInsightSubject } from '../../types/subjects'
import { CodeInsightsBackend } from '../code-insights-backend'
import {
    BackendInsightData,
    CaptureInsightSettings,
    DashboardCreateInput,
    DashboardCreateResult,
    DashboardDeleteInput,
    DashboardUpdateInput,
    DashboardUpdateResult,
    GetLangStatsInsightContentInput,
    GetSearchInsightContentInput,
    InsightCreateInput,
    InsightUpdateInput,
    ReachableInsight,
} from '../code-insights-backend-types'
import { getBuiltInInsight } from '../core/api/get-built-in-insight'
import { getLangStatsInsightContent } from '../core/api/get-lang-stats-insight-content'
import { getRepositorySuggestions } from '../core/api/get-repository-suggestions'
import { getResolvedSearchRepositories } from '../core/api/get-resolved-search-repositories'
import { getSearchInsightContent } from '../core/api/get-search-insight-content/get-search-insight-content'
import { parseDashboardScope } from '../utils/parse-dashboard-scope'

import { createInsightView } from './deserialization/create-insight-view'
import { GET_DASHBOARD_INSIGHTS_GQL } from './gql/GetDashboardInsights'
import { GET_INSIGHTS_GQL } from './gql/GetInsights'
import { GET_INSIGHTS_DASHBOARDS_GQL } from './gql/GetInsightsDashboards'
import { GET_INSIGHTS_SUBJECTS_GQL } from './gql/GetInsightSubjects'
import { createInsight } from './methods/create-insight/create-insight'
import { getBackendInsightData } from './methods/get-backend-insight-data/get-backend-insight-data'
import { getCaptureGroupInsightsPreview } from './methods/get-capture-group-insight-preivew'
import { updateInsight } from './methods/update-insight/update-insight'
import { createDashboardGrants } from './utils/get-dashboard-grants'

export class CodeInsightsGqlBackend implements CodeInsightsBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    // Insights
    public getInsights = (input: { dashboardId: string }): Observable<Insight[]> => {
        const { dashboardId } = input

        // Handle virtual dashboard that doesn't exist in BE gql API and cause of that
        // we need to use here insightViews query to fetch all available insights
        if (dashboardId === ALL_INSIGHTS_DASHBOARD_ID) {
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

                return createInsightView(insightData) || null
            })
        )

    // TODO: This method is used only for insight title validation but since we don't have
    // limitations about title field in gql api remove this method and async validation for
    // title field as soon as setting-based api will be deprecated
    public findInsightByName = (): Observable<Insight | null> => of(null)

    public getReachableInsights = (): Observable<ReachableInsight[]> =>
        this.getInsights({ dashboardId: ALL_INSIGHTS_DASHBOARD_ID }).pipe(
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

    public getBackendInsightData = (insight: BackendInsight): Observable<BackendInsightData> =>
        getBackendInsightData(this.apolloClient, insight)

    public getBuiltInInsightData = getBuiltInInsight

    // We don't have insight visibility and subject levels in the new GQL API anymore.
    // it was part of setting-cascade based API.
    public getInsightSubjects = (): Observable<SupportedInsightSubject[]> => of([])

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

    // Dashboards
    public getDashboards = (id?: string): Observable<InsightDashboard[]> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<InsightsDashboardsResult>({
                query: GET_INSIGHTS_DASHBOARDS_GQL,
                variables: { id },
                fetchPolicy: 'cache-first',
            })
        ).pipe(
            map(result => {
                const { data } = result

                return [
                    {
                        id: ALL_INSIGHTS_DASHBOARD_ID,
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

    public getDashboardById = (input: { dashboardId: string | undefined }): Observable<InsightDashboard | null> => {
        const { dashboardId } = input

        // the 'all' dashboardId is not a real dashboard so return nothing
        if (dashboardId === ALL_INSIGHTS_DASHBOARD_ID) {
            return of(null)
        }

        return this.getDashboards(dashboardId).pipe(
            map(dashboards => dashboards.find(({ id }) => id === dashboardId) ?? null)
        )
    }

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
                                        fragment NewDashboard on InsightsDashboard {
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

    public getCaptureInsightContent = (input: CaptureInsightSettings): Promise<LineChartContent<any, string>> =>
        getCaptureGroupInsightsPreview(this.apolloClient, input)

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
