import { Duration } from 'date-fns'

import { BaseInsight, InsightExecutionType, InsightFilters, InsightType } from '../common'

export type SearchBasedInsight = SearchRuntimeBasedInsight | SearchBackendBasedInsight

export interface SearchRuntimeBasedInsight extends BaseInsight {
    repositories: string[]
    series: SearchBasedInsightSeries[]
    step: Duration

    executionType: InsightExecutionType.Runtime
    type: InsightType.SearchBased
}

export interface SearchBackendBasedInsight extends BaseInsight {
    filters: InsightFilters
    series: SearchBasedInsightSeries[]
    step: Duration

    executionType: InsightExecutionType.Backend
    type: InsightType.SearchBased
}

export interface SearchBasedInsightSeries {
    id: string
    name: string
    query: string
    stroke?: string
}
