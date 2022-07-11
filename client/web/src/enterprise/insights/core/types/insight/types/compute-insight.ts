import { Duration } from 'date-fns'

import { GroupByField, Maybe } from '../../../../../../graphql-operations'
import { BaseInsight, InsightExecutionType, InsightFilters, InsightType } from '../common'

// Based off of SearchBasedInsight
// We do not extend SearchBasedInsight becuase this is going to be a temporary solution
export interface ComputeInsight extends BaseInsight {
    repositories: string[]
    filters: InsightFilters
    series: ComputeBasedInsightSeries[]
    step: Duration

    executionType: InsightExecutionType.Backend
    type: InsightType.Compute

    groupBy: Maybe<GroupByField>
}

export interface ComputeBasedInsightSeries {
    id: string
    name: string
    query: string
    stroke?: string
}
