import React from 'react'

import { throwError } from 'rxjs'

import { CodeInsightsBackend } from './code-insights-backend'

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

/**
 * Default context api class. Provides mock methods only.
 */
export class FakeDefaultCodeInsightsBackend implements CodeInsightsBackend {
    // Insights
    public getInsights = errorMockMethod('getInsights')
    public getInsightById = errorMockMethod('getInsightById')
    public hasInsights = errorMockMethod('hasInsight')
    public getActiveInsightsCount = errorMockMethod('getNonFrozenInsightsCount')
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

    public getFirstExampleRepository = errorMockMethod('getFirstExampleRepository')

    // License check
    public UIFeatures = { licensed: false, insightsLimit: null }
}

export const CodeInsightsBackendContext = React.createContext<CodeInsightsBackend>(new FakeDefaultCodeInsightsBackend())
