import { SeriesSortDirection, SeriesSortMode } from '../../../graphql-operations'

import { InsightsDashboardType, VirtualInsightsDashboard } from './types'
import { SeriesDisplayOptionsInputRequired } from './types/insight/common'

/**
 * Special virtual dashboard - "All Insights". This dashboard doesn't
 * exist in settings or in BE database.
 */
export const ALL_INSIGHTS_DASHBOARD: VirtualInsightsDashboard = {
    id: 'all',
    type: InsightsDashboardType.Virtual,
    title: 'All Insights',
}

// This constant should match the defaults set on the backend
// If this value is updated, make sure it matches the default in the backend
// limit: 20
// mode: ResultCount
// direction: Desc
export const DEFAULT_SERIES_DISPLAY_OPTIONS: SeriesDisplayOptionsInputRequired = {
    limit: 20,
    sortOptions: {
        direction: SeriesSortDirection.DESC,
        mode: SeriesSortMode.RESULT_COUNT,
    },
}
