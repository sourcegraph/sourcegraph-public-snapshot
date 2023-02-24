import { GroupByField } from '@sourcegraph/shared/src/graphql-operations'

import { BaseInsight, InsightFilters, InsightType } from '../common'

import { SearchBasedInsightSeries } from './search-insight'

export interface ComputeInsight extends BaseInsight {
    type: InsightType.Compute
    repositories: string[]
    filters: InsightFilters
    series: SearchBasedInsightSeries[]
    groupBy: GroupByField
}
