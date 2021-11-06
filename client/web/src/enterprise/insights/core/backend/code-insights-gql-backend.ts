import { ApolloClient, ApolloQueryResult, gql } from '@apollo/client'
import { Duration } from 'date-fns'
import { Observable, throwError, of, from } from 'rxjs'
import { map, mapTo, switchMap } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { isDefined } from '@sourcegraph/shared/src/util/types'
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
    CreateInsightResult,
    LineChartSearchInsightDataSeriesInput,
    LineChartSearchInsightInput,
    TimeIntervalStepUnit,
} from '@sourcegraph/web/src/graphql-operations'

import {
    Insight,
    InsightDashboard,
    InsightsDashboardType,
    InsightType,
    isSearchBasedInsight,
    SearchBasedInsight,
} from '../types'
import { isSearchBackendBasedInsight, SearchBackendBasedInsight } from '../types/insight/search-insight'
import { SupportedInsightSubject } from '../types/subjects'

import { InsightStillProcessingError } from './api/get-backend-insight'
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
    ReachableInsight,
} from './code-insights-backend-types'
import { createViewContent } from './utils/create-view-content'

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

function getStepInterval(insight: SearchBasedInsight): [TimeIntervalStepUnit, number] {
    if (insight.type === InsightType.Backend) {
        return [TimeIntervalStepUnit.WEEK, 2]
    }

    const castUnits = (Object.keys(insight.step) as (keyof Duration)[])
        .map<[TimeIntervalStepUnit, number] | null>(key => {
            switch (key) {
                case 'hours':
                    return [TimeIntervalStepUnit.HOUR, insight.step[key] ?? 0]
                case 'days':
                    return [TimeIntervalStepUnit.DAY, insight.step[key] ?? 0]
                case 'weeks':
                    return [TimeIntervalStepUnit.WEEK, insight.step[key] ?? 0]
                case 'months':
                    return [TimeIntervalStepUnit.MONTH, insight.step[key] ?? 0]
                case 'years':
                    return [TimeIntervalStepUnit.YEAR, insight.step[key] ?? 0]
            }

            return null
        })
        .filter(isDefined)

    if (castUnits.length === 0) {
        throw new Error('Wrong time step format')
    }

    // Return first valid match
    return castUnits[0]
}

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

    public findInsightByName = (input: FindInsightByNameInput): Observable<Insight | null> =>
        this.getInsights().pipe(map(insights => insights.find(insight => insight.title === input.name) || null))
    public getReachableInsights = (subjectId: string): Observable<ReachableInsight[]> =>
        this.getInsights().pipe(
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

    public createInsight = (input: InsightCreateInput): Observable<unknown> => {
        const { insight, dashboard } = input

        if (isSearchBasedInsight(insight)) {
            // Prepare repository insight array
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

            return from(
                (async () => {
                    const { data } = await this.apolloClient.mutate<CreateInsightResult>({
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

                    // TODO [VK] add attach to dashboard API call with newly create id and dashboard id
                    const insightId = data?.createLineChartSearchInsight.view.id ?? ''
                    const dashboardId = dashboard?.id ?? ''

                    console.log(`Add insight with id ${insightId} to dashboard with id ${dashboardId}`)
                })()
            )
        }

        // TODO [VK] implement lang stats chart creation
        return of()
    }

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
    public getSearchInsightContent = <D extends keyof ViewContexts>(
        input: GetSearchInsightContentInput<D>
    ): Promise<LineChartContent<any, string>> => getSearchInsightContent(input.insight, input.options)

    public getLangStatsInsightContent = <D extends keyof ViewContexts>(
        input: GetLangStatsInsightContentInput<D>
    ): Promise<PieChartContent<any>> => getLangStatsInsightContent(input.insight, input.options)

    // Repositories API
    public getRepositorySuggestions = getRepositorySuggestions
    public getResolvedSearchRepositories = getResolvedSearchRepositories
}
