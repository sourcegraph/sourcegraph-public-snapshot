import { type InsightDashboard, type CustomInsightDashboard, isVirtualDashboard } from '../../../../../../core/types'

/**
 * Only dashboards that are stored in user/org/global settings can be edited.
 *
 * Besides, these settings based (configurable dashboards) we have virtual and built-in dashboards
 * they can't be edited in any way (add/remove insights, delete)
 */
export const isDashboardConfigurable = (dashboard: InsightDashboard | undefined): dashboard is CustomInsightDashboard =>
    !!dashboard && !isVirtualDashboard(dashboard)
