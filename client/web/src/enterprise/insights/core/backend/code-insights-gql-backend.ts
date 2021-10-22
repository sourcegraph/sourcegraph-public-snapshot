import { ApolloClient, gql } from '@apollo/client'
import { Duration } from 'date-fns'
import { Observable, throwError, of, from } from 'rxjs'
import { map, mapTo } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { isDefined } from '@sourcegraph/shared/src/util/types'
import {
    CreateDashboardResult,
    CreateInsightsDashboardInput,
    DeleteDashboardResult,
    InsightsDashboardsResult,
    UpdateDashboardResult,
    UpdateInsightsDashboardInput,
    CreateInsightResult,
    GetInsightResult,
    LineChartSearchInsightDataSeriesInput,
    LineChartSearchInsightInput,
    TimeIntervalStepUnit,
} from '@sourcegraph/web/src/graphql-operations'

import {
    InsightDashboard,
    Insight,
    InsightType,
    InsightTypePrefix,
    isSearchBasedInsight,
    SearchBasedInsight,
} from '../types'
import { isSearchBackendBasedInsight } from '../types/insight/search-insight'
import { SupportedInsightSubject } from '../types/subjects'

import { getLangStatsInsightContent } from './api/get-lang-stats-insight-content'
import { getRepositorySuggestions } from './api/get-repository-suggestions'
import { getResolvedSearchRepositories } from './api/get-resolved-search-repositories'
import { getSearchInsightContent } from './api/get-search-insight-content/get-search-insight-content'
import { CodeInsightsBackend } from './code-insights-backend'
import {
    DashboardCreateInput,
    DashboardDeleteInput,
    DashboardUpdateInput,
    GetLangStatsInsightContentInput,
    GetSearchInsightContentInput,
    InsightCreateInput,
} from './code-insights-backend-types'

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

export class CodeInsightsGqlBackend implements CodeInsightsBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    // Insights
    public getInsights = errorMockMethod('getInsights')

    public getInsightById = (id: string): Observable<Insight | null> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<GetInsightResult>({
                query: gql`
                    query GetInsight($ids: [ID!]) {
                        insights(ids: $ids) {
                            nodes {
                                id
                                title
                                series {
                                    label
                                }
                            }
                        }
                    }
                `,
                variables: { ids: [id] },
            })
        ).pipe(
            map(result => {
                const { data } = result
                const insightData = data.insights?.nodes[0]

                if (!insightData) {
                    return null
                }

                // TODO [VK] Support lang stats insight
                // TODO [VK] Support different type of insight backend based and FE insight
                return {
                    type: InsightType.Backend,
                    // This is our convenstion around insight id, by this id prefix
                    // we make a difference between search and lang stats insight
                    id: `${InsightTypePrefix.search}.${insightData.id}`,
                    visibility: '',
                    title: insightData.title,
                    series: insightData.series.map(series => ({
                        name: series.label,
                        // TODO [VK] we don't have series query and color in gql API but we should add this
                        query: '',
                        stroke: '',
                    })),
                }
            })
        )

    public findInsightByName = errorMockMethod('findInsightByName')
    public getReachableInsights = errorMockMethod('getReachableInsights')
    public getBackendInsightData = errorMockMethod('getBackendInsightData')
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
                            }
                        }
                    }
                `,
            })
        ).pipe(
            map(({ data }) =>
                data.insightsDashboards.nodes.map(dashboard => ({
                    id: dashboard.id,
                    title: dashboard.title,
                    insightIds: dashboard.views?.nodes.map(view => view.id),
                }))
            )
        )
    // TODO [VK]: Omit for now usage of dashboard info in the creation UI.
    // We have to implement and merge that after the dashboard page migration will be ready.
    public getDashboardById = (): Observable<null> => of(null)
    public findDashboardByName = errorMockMethod('findDashboardByName')

    public createDashboard = (input: DashboardCreateInput): Observable<void> => {
        if (!input.grants) {
            throw new Error('`grants` are required to create a new dashboard')
        }

        const mappedInput: CreateInsightsDashboardInput = {
            title: input.name,
            grants: input.grants,
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

        if (!nextDashboardInput.grants) {
            throw new Error('`grants` are required to update a dashboard')
        }

        const input: UpdateInsightsDashboardInput = {
            title: nextDashboardInput.name,
            grants: nextDashboardInput.grants,
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
