import { Duration } from 'date-fns'
import { LineChartContent } from 'sourcegraph'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

import { ExtensionInsight, Insight, InsightDashboard, SearchBasedInsightSeries } from '../types'

export interface DashboardCreateInput {
    name: string
    visibility: string
    insightIds?: string[]
    type?: string
}

export interface DashboardCreateResult {
    id: string
}

export interface DashboardUpdateInput {
    nextDashboardInput: DashboardCreateInput
    id: string
}

export interface AssignInsightsToDashboardInput {
    id: string
    prevInsightIds: string[]
    nextInsightIds: string[]
}

export interface DashboardUpdateResult {
    id: string
}

export interface DashboardDeleteInput {
    id: string
}

export interface FindInsightByNameInput {
    name: string
}

export interface InsightCreateInput {
    insight: Insight
    dashboard: InsightDashboard | null
}

export interface InsightUpdateInput {
    oldInsight: Insight
    newInsight: Insight
}

export interface SearchInsightSettings {
    series: SearchBasedInsightSeries[]
    step: Duration
    repositories: string[]
}

export interface LangStatsInsightsSettings {
    repository: string
    otherThreshold: number
}

export interface CaptureInsightSettings {
    repositories: string[]
    query: string
    step: Duration
}

export type ReachableInsight = Insight & {
    owner: {
        id: string
        name: string
    }
}

export interface BackendInsightData {
    id: string
    view: {
        title: string
        subtitle?: string
        content: LineChartContent<any, string>[]
        isFetchingHistoricalData: boolean
    }
}

export interface GetBuiltInsightInput<D extends keyof ViewContexts> {
    insight: ExtensionInsight
    options: { where: D; context: ViewContexts[D] }
}

export interface GetSearchInsightContentInput<D extends keyof ViewContexts> {
    insight: SearchInsightSettings
    options: {
        where: D
        context: ViewContexts[D]
    }
}

export interface GetLangStatsInsightContentInput<D extends keyof ViewContexts> {
    insight: LangStatsInsightsSettings
    options: {
        where: D
        context: ViewContexts[D]
    }
}

export interface RepositorySuggestionData {
    id: string
    name: string
}
