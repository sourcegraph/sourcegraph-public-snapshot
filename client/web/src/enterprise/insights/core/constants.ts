import { InsightsDashboardType, VirtualInsightsDashboard } from './types'

/**
 * Special virtual dashboard - "All Insights". This dashboard doesn't
 * exist in settings or in BE database.
 */
export const ALL_INSIGHTS_DASHBOARD: VirtualInsightsDashboard = {
    id: 'all',
    type: InsightsDashboardType.Virtual,
    title: 'All Insights',
}
