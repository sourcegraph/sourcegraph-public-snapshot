import { Observable, throwError, of } from 'rxjs'
import { map, mapTo, mergeMap } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import {
    CreateInsightsDashboardInput,
    InsightsDashboardsResult,
    UpdateInsightsDashboardInput,
} from '../../../../graphql-operations'
import { InsightDashboard, InsightsDashboardType } from '../types'
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
                                    dataSeries {
                                        label
                                        points {
                                            dateTime
                                            value
                                        }
                                    }
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
    public getDashboardById = (dashboardId?: string): Observable<InsightDashboard | null> =>
        this.getDashboards().pipe(mergeMap(dashboards => dashboards.filter(dashboard => dashboard.id === dashboardId)))

    public findDashboardByName = (name: string): Observable<InsightDashboard | null> =>
        this.getDashboards().pipe(
            mergeMap(dashboards => dashboards.filter(dashboard => 'title' in dashboard && dashboard.title === name))
        )

    // TODO: Update input to use CreateInsightsDashboardInput directly
    public createDashboard = (input: DashboardCreateInput): Observable<void> => {
        const mappedInput: CreateInsightsDashboardInput = {
            title: input.name,
            grants: {
                global: input.visibility === InsightsDashboardType.Global,
                users: input.visibility === InsightsDashboardType.Personal ? ['TODO: Get userID'] : [],
                organizations: input.visibility === InsightsDashboardType.Organization ? ['TODO: Get orgID'] : [],
            },
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

    // TODO: Update input to use ID directly
    public deleteDashboard = (input: DashboardDeleteInput): Observable<void> => {
        const mappedInput: { id: string } = { id: input.dashboardSettingKey }
        return requestGraphQL(
            gql`
                mutation DeleteDashboard($id: ID!) {
                    deleteInsightsDashboard(id: $id) {
                        alwaysNil
                    }
                }
            `,
            mappedInput
        ).pipe(mapTo(undefined))
    }

    // TODO: Update input to use UpdateInsightsDashboardInput directly
    public updateDashboard = (input: DashboardUpdateInput): Observable<void> => {
        const mappedInput: UpdateInsightsDashboardInput = {
            title: input.nextDashboardInput.name,
            grants: {
                global: input.nextDashboardInput.visibility === InsightsDashboardType.Global,
                users:
                    input.nextDashboardInput.visibility === InsightsDashboardType.Personal ? ['TODO: Get userID'] : [],
                organizations:
                    input.nextDashboardInput.visibility === InsightsDashboardType.Organization
                        ? ['TODO: Get orgID']
                        : [],
            },
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
            mappedInput
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
