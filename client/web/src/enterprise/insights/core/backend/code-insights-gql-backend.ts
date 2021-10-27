import { Observable, throwError, of } from 'rxjs'
import { map, mapTo } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import {
    CreateInsightsDashboardInput,
    InsightsDashboardsResult,
    UpdateInsightsDashboardInput,
} from '../../../../graphql-operations'
import { InsightDashboard } from '../types'
import { SupportedInsightSubject } from '../types/subjects'

import { CodeInsightsBackend } from './code-insights-backend'
import {
    DashboardCreateInput,
    DashboardDeleteInput,
    DashboardUpdateInput,
    RepositorySuggestionData,
} from './code-insights-backend-types'

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

export class CodeInsightsGqlBackend implements CodeInsightsBackend {
    // Insights
    public getInsights = errorMockMethod('getInsights')
    public getInsightById = errorMockMethod('getInsightById')
    public findInsightByName = errorMockMethod('findInsightByName')
    public getReachableInsights = errorMockMethod('getReachableInsights')
    public getBackendInsightData = errorMockMethod('getBackendInsightData')
    public getBuiltInInsightData = errorMockMethod('getBuiltInInsightData')
    public getInsightSubjects = (): Observable<SupportedInsightSubject[]> => {
        console.warn('TODO: Get insight subjects')
        return of([])
    }
    public getSubjectSettingsById = errorMockMethod('getSubjectSettingsById')
    public createInsight = errorMockMethod('createInsight')
    public createInsightWithNewFilters = errorMockMethod('createInsightWithNewFilters')
    public updateInsight = errorMockMethod('updateInsight')
    public deleteInsight = errorMockMethod('deleteInsight')

    // Dashboards
    public getDashboards = (): Observable<InsightDashboard[]> =>
        requestGraphQL<InsightsDashboardsResult>(
            gql`
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
            {}
        ).pipe(
            map(dataOrThrowErrors),
            map(data =>
                data.insightsDashboards.nodes.map(dashboard => ({
                    id: dashboard.id,
                    title: dashboard.title,
                    insightIds: dashboard.views?.nodes.map(view => view.id),
                }))
            )
        )
    public getDashboardById = errorMockMethod('getDashboardById')
    public findDashboardByName = errorMockMethod('findDashboardByName')

    public createDashboard = (input: DashboardCreateInput): Observable<void> => {
        if (!input.grants) {
            throw new Error('`grants` are required to create a new dashboard')
        }

        const mappedInput: CreateInsightsDashboardInput = {
            title: input.name,
            grants: input.grants,
        }

        return requestGraphQL(
            gql`
                mutation CreateDashboard($input: CreateInsightsDashboardInput!) {
                    createInsightsDashboard(input: $input) {
                        dashboard {
                            id
                        }
                    }
                }
            `,
            { input: mappedInput }
        ).pipe(mapTo(undefined))
    }

    public deleteDashboard = (input: DashboardDeleteInput): Observable<void> => {
        if (!input.id) {
            throw new Error('`id` is required to delete a dashboard')
        }

        return requestGraphQL(
            gql`
                mutation DeleteDashboard($id: ID!) {
                    deleteInsightsDashboard(id: $id) {
                        alwaysNil
                    }
                }
            `,
            input.id
        ).pipe(mapTo(undefined))
    }

    public updateDashboard = ({ id, nextDashboardInput }: DashboardUpdateInput): Observable<void> => {
        if (!id) {
            throw new Error('`id` is required to update a dashboard')
        }

        if (!nextDashboardInput.grants) {
            throw new Error('`grants` are required to update a dashboard')
        }

        const mappedInput: UpdateInsightsDashboardInput = {
            title: nextDashboardInput.name,
            grants: nextDashboardInput.grants,
        }

        return requestGraphQL(
            gql`
                mutation UpdateDashboard($id: ID!, $input: UpdateInsightsDashboardInput!) {
                    updateInsightsDashboard(id: $id, input: $input) {
                        dashboard {
                            id
                        }
                    }
                }
            `,
            {
                id,
                input: mappedInput,
            }
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
