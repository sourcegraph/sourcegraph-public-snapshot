import { Duration } from 'date-fns'
import { LineChartContent } from 'sourcegraph'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

import { ExtensionInsight, Insight, InsightDashboard, InsightTypePrefix, SettingsBasedInsightDashboard } from '../types'
import {
    SearchBackendBasedInsight,
    SearchBasedBackendFilters,
    SearchBasedInsightSeries,
} from '../types/insight/search-insight'

export interface DashboardCreateInput {
    name: string
    visibility: string
}

export interface DashboardUpdateInput {
    previousDashboard: SettingsBasedInsightDashboard
    nextDashboard: SettingsBasedInsightDashboard
}

export interface DashboardDeleteInput {
    dashboardSettingKey: string
    dashboardOwnerId: string
}

export interface FindInsightByNameInput {
    name: string
    type: InsightTypePrefix
}

export interface InsightCreateInput {
    insight: Insight
    dashboard: InsightDashboard | null
}

export interface InsightUpdateInput {
    oldInsight: Insight
    newInsight: Insight
}

export interface InsightCreateWithFiltersInput {
    insightName: string
    originalInsight: SearchBackendBasedInsight
    dashboard: InsightDashboard
    filters: SearchBasedBackendFilters
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

export type ReachableInsight = Insight & {
    owner: {
        id: string
        name: string
    }
}

export interface SubjectSettingsData {
    id: number | null
    contents: string
}

export interface BackendInsightData {
    id: string
    view: {
        title: string
        subtitle: string
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
