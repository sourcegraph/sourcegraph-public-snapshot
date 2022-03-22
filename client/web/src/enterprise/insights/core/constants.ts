import { InsightsDashboardType, VirtualInsightsDashboard } from './types'

/**
 * One of virtual dashboard id that contains all available for a user insights
 */
export const ALL_INSIGHTS_DASHBOARD_ID = 'all'

/**
 * Special virtual dashboard - "All Insights". This dashboard doesn't
 * exist in settings or in BE database.
 */
export const ALL_INSIGHTS_DASHBOARD: VirtualInsightsDashboard = {
    id: 'all',
    type: InsightsDashboardType.Virtual,
    title: 'All Insights',
}
