import React from 'react'

import { throwError } from 'rxjs'

import { CodeInsightsBackend } from './code-insights-backend'
import { SeriesChartContent, CategoricalChartContent, BackendInsightDatum } from './code-insights-backend-types'

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
    public getBuiltInInsightData = errorMockMethod('getBuiltInInsightData')
    public createInsight = errorMockMethod('createInsight')
    public updateInsight = errorMockMethod('updateInsight')
    public deleteInsight = errorMockMethod('deleteInsight')
    public removeInsightFromDashboard = errorMockMethod('removeInsightFromDashboard')

    // Dashboards
    public getDashboardOwners = errorMockMethod('getDashboardSubjects')
    public createDashboard = errorMockMethod('createDashboard')
    public deleteDashboard = errorMockMethod('deleteDashboard')
    public updateDashboard = errorMockMethod('updateDashboard')
    public assignInsightsToDashboard = errorMockMethod('assignInsightsToDashboard')

    // Live preview fetchers
    public getSearchInsightContent = (): Promise<SeriesChartContent<unknown>> =>
        errorMockMethod('getSearchInsightContent')().toPromise()
    public getLangStatsInsightContent = (): Promise<CategoricalChartContent<unknown>> =>
        errorMockMethod('getLangStatsInsightContent')().toPromise()

    public getInsightPreviewContent = (): Promise<SeriesChartContent<BackendInsightDatum>> =>
        errorMockMethod('getInsightPreviewContent')().toPromise()

    public getFirstExampleRepository = errorMockMethod('getFirstExampleRepository')

    // License check
    public UIFeatures = { licensed: false, insightsLimit: null }
}

export const CodeInsightsBackendContext = React.createContext<CodeInsightsBackend>(new FakeDefaultCodeInsightsBackend())
