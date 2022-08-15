import { Duration } from 'date-fns'

import { BaseInsight, InsightExecutionType, InsightFilters, InsightType } from '../common'

export interface SearchBasedInsight extends BaseInsight {
    repositories: string[]
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
