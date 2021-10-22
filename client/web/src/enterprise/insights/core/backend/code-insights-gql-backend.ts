import { Observable, throwError, of } from 'rxjs'
import { map } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import { InsightsDashboardsResult } from '../../../../graphql-operations'
import { InsightDashboard } from '../types'
import { SupportedInsightSubject } from '../types/subjects'

import { CodeInsightsBackend } from './code-insights-backend'
import { RepositorySuggestionData } from './code-insights-backend-types'

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
    public getDashboardById = errorMockMethod('getDashboardById')
    public findDashboardByName = errorMockMethod('findDashboardByName')
    public createDashboard = errorMockMethod('createDashboard')
    public deleteDashboard = errorMockMethod('deleteDashboard')
    public updateDashboard = errorMockMethod('updateDashboard')

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
