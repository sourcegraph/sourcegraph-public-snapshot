import type { Duration } from 'date-fns'

import type { BaseInsight, InsightFilters, InsightType } from '../common'

export interface CaptureGroupInsight extends BaseInsight {
    type: InsightType.CaptureGroup

    /** Capture group regexp query string */
    query: string

    repoQuery: string

    /**
     * List of repositories that are used to collect data by query regexp field
     */
    repositories: string[]
    step: Duration
    filters: InsightFilters
}
