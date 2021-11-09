import { ApolloClient, ApolloQueryResult, gql } from '@apollo/client'
import { Observable, throwError, of, from } from 'rxjs'
import { map, mapTo, switchMap } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'
import {
    CreateDashboardResult,
    CreateInsightsDashboardInput,
    DeleteDashboardResult,
    GetInsightResult,
    GetInsightsResult,
    InsightsDashboardsResult,
    InsightFields,
    InsightsPermissionGrantsInput,
    UpdateDashboardResult,
    UpdateInsightsDashboardInput,
} from '@sourcegraph/web/src/graphql-operations'

import { Insight, InsightDashboard, InsightsDashboardType, InsightType, SearchBasedInsight } from '../types'
import { SearchBackendBasedInsight } from '../types/insight/search-insight'
import { SupportedInsightSubject } from '../types/subjects'

import { InsightStillProcessingError } from './api/get-backend-insight'
import { CodeInsightsBackend } from './code-insights-backend'
import {
    BackendInsightData,
    DashboardCreateInput,
    DashboardDeleteInput,
    DashboardUpdateInput,
    RepositorySuggestionData,
} from './code-insights-backend-types'
import { createViewContent } from './utils/create-view-content'

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

/**
 * Helper function to parse the dashboard type from the grants object.
 * TODO: Remove this function when settings api is deprecated
 *
 * @param grants {object} - A grants object from an insight dashboard
 * @param grants.global {boolean}
 * @param grants.users {string[]}
 * @param grants.organizations {string[]}
 * @returns - The type of the dashboard
 */
export const parseType = (grants?: {
    global?: boolean
    users?: string[]
    organizations?: string[]
}): InsightsDashboardType.Personal | InsightsDashboardType.Organization | InsightsDashboardType.Global => {
    if (grants?.global) {
        return InsightsDashboardType.Global
    }
    if (grants?.organizations?.length) {
        return InsightsDashboardType.Organization
    }
    return InsightsDashboardType.Personal
}

/**
 * Helper function to parse a grants object from a given type and visibility.
 * TODO: Remove this function when settings api is deprecated
 *
 * @param type {('personal'|'organization'|'global')} - The type of the dashboard
 * @param visibility {string} - Usually the user or organization id
 * @returns - A properly formatted grants object
 */
export const parseGrants = (type: string, visibility: string): InsightsPermissionGrantsInput => {
    const grants: InsightsPermissionGrantsInput = {}
    if (type === 'personal') {
        grants.users = [visibility]
    }
    if (type === 'organization') {
        grants.organizations = [visibility]
    }
    if (type === 'global') {
        grants.global = true
    }

    return grants
}

const mapInsightView = (insight: GetInsightsResult['insightViews']['nodes'][0]): Insight => ({
    __typename: insight.__typename,
    presentationType: insight.presentation.__typename,
    type: InsightType.Backend,
    id: insight.id,
    visibility: '',
    title: insight.presentation.title,
    series: insight.dataSeries.map(series => ({
        name: series.label,
        query:
            insight.dataSeriesDefinitions.find(definition => definition.seriesId === series.seriesId)?.query ||
            'QUERY NOT FOUND',
        stroke: insight.presentation.seriesPresentation.find(presentation => presentation.seriesId === series.seriesId)
            ?.color,
    })),
})

const mapInsightFields = (insight: GetInsightsResult['insightViews']['nodes'][0]): InsightFields => ({
    id: insight.id,
    title: insight.presentation.title,
    description: '',
    series: insight.dataSeries.map(series => ({
        label: series.label,
        points: series.points,
        status: series.status,
    })),
})

const insightViewsFieldsFragment = gql`
    fragment InsightViewsFields on InsightView {
        id
        presentation {
            __typename
            ... on LineChartInsightViewPresentation {
                title
                seriesPresentation {
                    seriesId
                    label
                    color
                }
            }
        }
        dataSeries {
            seriesId
            label
            points {
                dateTime
                value
            }
            status {
                totalPoints
                pendingJobs
                completedJobs
                failedJobs
                backfillQueuedAt
            }
        }
        dataSeriesDefinitions {
            ... on SearchInsightDataSeriesDefinition {
                seriesId
                query
                repositoryScope {
                    repositories
                }
                timeScope {
                    ... on InsightIntervalTimeScope {
                        unit
                        value
                    }
                }
            }
        }
    }
`

export class CodeInsightsGqlBackend implements CodeInsightsBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    // Insights
    public getInsights = (ids?: string[]): Observable<Insight[]> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<GetInsightsResult>({
                query: gql`
                    query GetInsights {
                        insightViews {
                            nodes {
                                ...InsightViewsFields
                            }
                        }
                    }
                    ${insightViewsFieldsFragment}
                `,
            })
        ).pipe(
            map(({ data }) => {
                const insightViews = data.insightViews.nodes.map(mapInsightView)
                if (ids) {
                    return insightViews.filter(insight => ids.includes(insight.id))
                }

                return insightViews
            })
        )

    private getInsightView = (id: string): Observable<ApolloQueryResult<GetInsightResult>> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<GetInsightResult>({
                query: gql`
                    query GetInsight($id: ID!) {
                        insightViews(id: $id) {
                            nodes {
                                ...InsightViewsFields
                            }
                        }
                    }
                    ${insightViewsFieldsFragment}
                `,
                variables: { id },
            })
        )

    public getInsightById = (id: string): Observable<Insight | null> =>
        this.getInsightView(id).pipe(
            map(({ data }) => {
                const insightData = data.insightViews.nodes[0]

                if (!insightData) {
                    return null
                }

                // TODO [VK] Support lang stats insight
                // TODO [VK] Support different type of insight backend based and FE insight
                return mapInsightView(insightData)
            })
        )

    public findInsightByName = errorMockMethod('findInsightByName')
    public getReachableInsights = errorMockMethod('getReachableInsights')

    // TODO: Rethink all of this method. Currently `createViewContent` expects a different format of
    // the `Insight` type than we use elsewhere. This is a temporary solution to make the code
    // fit with both of those shapes.
    public getBackendInsightData = (insight: SearchBackendBasedInsight): Observable<BackendInsightData> =>
        this.getInsightView(insight.id).pipe(
            // Note: this insight is guaranteed to exist since this function
            // is only called from within a loop of insight ids
            map(({ data }) => ({
                insight: mapInsightView(data.insightViews.nodes[0]) as SearchBasedInsight,
                insightFields: mapInsightFields(data.insightViews.nodes[0]),
            })),
            switchMap(({ insight, insightFields }) => {
                if (!insight) {
                    return throwError(new InsightStillProcessingError())
                }

                return of({ insight, insightFields })
            }),
            map(({ insight, insightFields }) => ({
                id: insight.id,
                view: {
                    title: insight.title ?? insight.title,
                    subtitle: '', // TODO: is this still used anywhere?
                    content: [
                        createViewContent(
                            {
                                id: insight.id,
                                description: '',
                                title: insight.title,
                                series: insightFields.series,
                            },
                            insight.series
                        ),
                    ],
                    isFetchingHistoricalData: insightFields.series.some(
                        ({ status: { pendingJobs, backfillQueuedAt } }) => pendingJobs > 0 || backfillQueuedAt === null
                    ),
                },
            }))
        )
    public getBuiltInInsightData = errorMockMethod('getBuiltInInsightData')

    // We don't have insight visibility and subject levels in the new GQL API anymore.
    // it was part of setting-cascade based API.
    public getInsightSubjects = (): Observable<SupportedInsightSubject[]> => of([])

    public getSubjectSettingsById = errorMockMethod('getSubjectSettingsById')
    public createInsight = errorMockMethod('createInsight')
    public createInsightWithNewFilters = errorMockMethod('createInsightWithNewFilters')
    public updateInsight = errorMockMethod('updateInsight')
    public deleteInsight = errorMockMethod('deleteInsight')

    // Dashboards
    public getDashboards = (): Observable<InsightDashboard[]> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<InsightsDashboardsResult>({
                query: gql`
                    query InsightsDashboards {
                        insightsDashboards {
                            nodes {
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
            })
        ).pipe(
            map(({ data }) =>
                data.insightsDashboards.nodes.map(
                    (dashboard): InsightDashboard => ({
                        id: dashboard.id,
                        title: dashboard.title,
                        insightIds: dashboard.views?.nodes.map(view => view.id),
                        grants: dashboard.grants,
                        type: parseType(dashboard.grants),
                    })
                )
            )
        )
    public getDashboardById = (dashboardId?: string): Observable<InsightDashboard | undefined> =>
        this.getDashboards().pipe(map(dashboards => dashboards.find(({ id }) => id === dashboardId)))

    public findDashboardByName = errorMockMethod('findDashboardByName')

    public createDashboard = (input: DashboardCreateInput): Observable<void> => {
        if (!input.type) {
            throw new Error('`grants` are required to create a new dashboard')
        }

        const mappedInput: CreateInsightsDashboardInput = {
            title: input.name,
            grants: parseGrants(input.type, input.visibility),
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
            grants: parseGrants(nextDashboardInput.type, nextDashboardInput.visibility),
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
    public getSearchInsightContent = (): Promise<LineChartContent<any, string>> =>
        errorMockMethod('getSearchInsightContent')().toPromise()
    public getLangStatsInsightContent = (): Promise<PieChartContent<any>> =>
        errorMockMethod('getLangStatsInsightContent')().toPromise()

    // Repositories API
    public getRepositorySuggestions = (): Promise<RepositorySuggestionData[]> =>
        errorMockMethod('getRepositorySuggestions')().toPromise()
    public getResolvedSearchRepositories = (): Promise<string[]> =>
        errorMockMethod('getResolvedSearchRepositories')().toPromise()
}
