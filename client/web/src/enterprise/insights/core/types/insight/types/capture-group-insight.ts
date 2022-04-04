import { Duration } from 'date-fns'

import { BaseInsight, InsightExecutionType, InsightFilters, InsightType } from '../common'

export interface CaptureGroupInsight extends BaseInsight {
    /**
     * We do not support capture group insight in runtime mode.
     * Capture group should always have data provided by BE.
     */
    executionType: InsightExecutionType.Backend
    type: InsightType.CaptureGroup

    /** Capture group regexp query string */
    query: string

    /**
     * List of repositories that are used to collect data by query regexp field
     */
    repositories: string[]
    step: Duration
    filters: InsightFilters
}
