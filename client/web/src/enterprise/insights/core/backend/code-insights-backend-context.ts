import React from 'react'

import { throwError } from 'rxjs'

import { CodeInsightsBackend } from './code-insights-backend'
import { SeriesChartContent, CategoricalChartContent, RepositorySuggestionData } from './code-insights-backend-types'

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

/**
 * Default context api class. Provides mock methods only.
 */
export class FakeDefaultCodeInsightsBackend implements CodeInsightsBackend {
    // Insights
    public getInsights = errorMockMethod('getInsights')
    public getInsightById = errorMockMethod('getInsightById')
    public findInsightByName = errorMockMethod('findInsightByName')
    public hasInsights = errorMockMethod('hasInsight')
    public getActiveInsightsCount = errorMockMethod('getNonFrozenInsightsCount')
    public getAccessibleInsightsList = errorMockMethod('getReachableInsights')
    public getBackendInsightData = errorMockMethod('getBackendInsightData')
    public getBuiltInInsightData = errorMockMethod('getBuiltInInsightData')
    public getInsightSubjects = errorMockMethod('getInsightSubjects')
    public getSubjectSettingsById = errorMockMethod('getSubjectSettingsById')
    public createInsight = errorMockMethod('createInsight')
    public createInsightWithNewFilters = errorMockMethod('createInsightWithNewFilters')
    public updateInsight = errorMockMethod('updateInsight')
    public deleteInsight = errorMockMethod('deleteInsight')
    public removeInsightFromDashboard = errorMockMethod('removeInsightFromDashboard')

    // Dashboards
    public getDashboards = errorMockMethod('getDashboards')
    public getDashboardById = errorMockMethod('getDashboardById')
    public getDashboardOwners = errorMockMethod('getDashboardSubjects')
    public findDashboardByName = errorMockMethod('findDashboardByName')
    public createDashboard = errorMockMethod('createDashboard')
    public deleteDashboard = errorMockMethod('deleteDashboard')
    public updateDashboard = errorMockMethod('updateDashboard')
    public assignInsightsToDashboard = errorMockMethod('assignInsightsToDashboard')

    // Live preview fetchers
    public getSearchInsightContent = (): Promise<SeriesChartContent<unknown>> =>
        errorMockMethod('getSearchInsightContent')().toPromise()
    public getLangStatsInsightContent = (): Promise<CategoricalChartContent<unknown>> =>
        errorMockMethod('getLangStatsInsightContent')().toPromise()

    public getCaptureInsightContent = (): Promise<SeriesChartContent<unknown>> =>
        errorMockMethod('getCaptureInsightContent')().toPromise()

    // Repositories API
    public getRepositorySuggestions = (): Promise<RepositorySuggestionData[]> =>
        errorMockMethod('getRepositorySuggestions')().toPromise()
    public getResolvedSearchRepositories = (): Promise<string[]> =>
        errorMockMethod('getResolvedSearchRepositories')().toPromise()
    public getFirstExampleRepository = errorMockMethod('getFirstExampleRepository')

    // License check
    public UIFeatures = { licensed: false, insightsLimit: null }
}

export const CodeInsightsBackendContext = React.createContext<CodeInsightsBackend>(new FakeDefaultCodeInsightsBackend())
