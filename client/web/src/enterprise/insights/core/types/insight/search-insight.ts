import { Duration } from 'date-fns'

import { InsightExecutionType, InsightFilters, InsightType } from './common'

export type SearchBasedInsight = SearchRuntimeBasedInsight | SearchBackendBasedInsight

export interface SearchRuntimeBasedInsight {
    id: string
    title: string
    repositories: string[]
    series: SearchBasedInsightSeries[]
    step: Duration
    dashboardReferenceCount: number

    executionType: InsightExecutionType.Runtime
    type: InsightType.SearchBased
}

export interface SearchBackendBasedInsight {
    id: string
    title: string
    filters: InsightFilters
    series: SearchBasedInsightSeries[]
    step: Duration
    dashboardReferenceCount: number

    executionType: InsightExecutionType.Backend
    type: InsightType.SearchBased
}

export interface SearchBasedInsightSeries {
    id: string
    name: string
    query: string
    stroke?: string
}

export const isSearchBackendBasedInsight = (insight: SearchBasedInsight): insight is SearchBackendBasedInsight =>
    insight.executionType === InsightExecutionType.Backend
