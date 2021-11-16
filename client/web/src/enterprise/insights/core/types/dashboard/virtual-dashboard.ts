import { InsightDashboard, InsightsDashboardScope, InsightsDashboardType } from './core'

/**
 * Special 'virtual' dashboard that includes all insights from the personal and organization and global
 * level dashboards. Virtual dashboard doesn't exist in settings but lives only in a runtime.
 */
export interface VirtualInsightsDashboard extends InsightDashboard {
    type: InsightsDashboardType.Virtual
    scope: InsightsDashboardScope.Personal
    id: string
    insightIds?: string[]
}

/**
 * One of virtual dashboard id that contains all available for a user insights
 */
export const ALL_INSIGHTS_DASHBOARD_ID = 'all'
