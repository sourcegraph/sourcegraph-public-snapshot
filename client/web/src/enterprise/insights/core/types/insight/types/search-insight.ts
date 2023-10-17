import type { Duration } from 'date-fns'

import type { BaseInsight, InsightFilters, InsightType } from '../common'

export interface SearchBasedInsight extends BaseInsight {
    type: InsightType.SearchBased
    repositories: string[]
    repoQuery: string
    filters: InsightFilters
    series: SearchBasedInsightSeries[]
    step: Duration
}

export interface SearchBasedInsightSeries {
    id: string
    name: string
    query: string
    stroke?: string
}
