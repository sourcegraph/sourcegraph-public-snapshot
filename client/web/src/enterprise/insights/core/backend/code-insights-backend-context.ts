import React from 'react'
import { throwError } from 'rxjs'

import { CodeInsightsBackend } from './code-insights-backend'

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

export class FakeDefaultCodeInsightsBackend implements CodeInsightsBackend {
    // Insights
    public getInsights = errorMockMethod('getInsights')
    public getInsightById = errorMockMethod('getInsightById')
    public findInsightByName = errorMockMethod('findInsightByName')
    public getReachableInsights = errorMockMethod('getReachableInsights')
    public getBackendInsightData = errorMockMethod('getBackendInsightData')
    public getBuiltInInsightData = errorMockMethod('getBuiltInInsightData')
    public getInsightSubjects = errorMockMethod('getInsightSubjects')
    public getSubjectSettingsById = errorMockMethod('getSubjectSettingsById')
    public createInsight = errorMockMethod('createInsight')
    public createInsightWithNewFilters = errorMockMethod('createInsightWithNewFilters')
    public updateInsight = errorMockMethod('updateInsight')
    public deleteInsight = errorMockMethod('deleteInsight')

    // Dashboards
    public getDashboards = errorMockMethod('getDashboards')
    public getDashboardById = errorMockMethod('getDashboardById')
    public findDashboardByName = errorMockMethod('findDashboardByName')
    public createDashboard = errorMockMethod('createDashboard')
    public deleteDashboard = errorMockMethod('deleteDashboard')
    public updateDashboard = errorMockMethod('updateDashboard')

    // Live preview fetchers
    public getSearchInsightContent = () => errorMockMethod('getSearchInsightContent')().toPromise()
    public getLangStatsInsightContent = () => errorMockMethod('getLangStatsInsightContent')().toPromise()

    // Repositories API
    public getRepositorySuggestions = () => errorMockMethod('getRepositorySuggestions')().toPromise()
    public getResolvedSearchRepositories = () => errorMockMethod('getResolvedSearchRepositories')().toPromise()
}

export const CodeInsightsBackendContext = React.createContext<CodeInsightsBackend>(new FakeDefaultCodeInsightsBackend())
