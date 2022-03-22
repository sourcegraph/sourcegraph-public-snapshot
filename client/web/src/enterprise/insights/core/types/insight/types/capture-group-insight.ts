import { Duration } from 'date-fns'

import { InsightExecutionType, InsightFilters, InsightType } from '../common'

export interface CaptureGroupInsight {
    id: string
    /**
     * We do not support capture group insight in setting based API or in
     * runtime mode. Capture group should always have data provided by BE.
     */
    executionType: InsightExecutionType.Backend
    type: InsightType.CaptureGroup
    title: string

    /** Capture group regexp query string */
    query: string

    /**
     * List of repositories that are used to collect data by query regexp field
     */
    repositories: string[]
    step: Duration
    dashboardReferenceCount: number
    filters: InsightFilters
}
