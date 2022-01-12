import { Duration } from 'date-fns'

import { InsightFilters } from '../../../../../schema/settings.schema'

import { InsightExecutionType, InsightType, SyntheticInsightFields } from './common'

export interface CaptureGroupInsight extends SyntheticInsightFields {
    /**
     * We do not support capture group insight in setting based API or in
     * runtime mode. Capture group should always have data provided by BE.
     */
    type: InsightExecutionType.Backend
    viewType: InsightType.CaptureGroup
    title: string

    /** Capture group regexp query string */
    query: string

    /**
     * List of repositories that are used to collect data by query regexp field
     */
    repositories: string[]
    step: Duration
    filters: InsightFilters
}
