import { ApolloClient, gql } from '@apollo/client'
import { Observable, throwError, of, from } from 'rxjs'
import { map, mapTo } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'
import {
    CreateDashboardResult,
    CreateInsightsDashboardInput,
    DeleteDashboardResult,
    InsightsDashboardsResult,
    UpdateDashboardResult,
    UpdateInsightsDashboardInput,
} from '@sourcegraph/web/src/graphql-operations'

import { InsightDashboard } from '../types'
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
} from './code-insights-backend-types'

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

export class CodeInsightsGqlBackend implements CodeInsightsBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    // Insights
    public getInsights = errorMockMethod('getInsights')
    public getInsightById = errorMockMethod('getInsightById')
    public findInsightByName = errorMockMethod('findInsightByName')
    public getReachableInsights = errorMockMethod('getReachableInsights')
    public getBackendInsightData = errorMockMethod('getBackendInsightData')
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
