import { InsightDashboard, isRealDashboard, CustomInsightDashboard } from '../../../../../../core/types'
import { isCustomInsightDashboard } from '../../../../../../core/types/dashboard/real-dashboard'

/**
 * Only dashboards that are stored in user/org/global settings can be edited.
 *
 * Besides these settings based (configurable dashboards) we have virtual and built-in dashboards
 * they can't be edited in any way (add/remove insights, delete)
 */
export const isDashboardConfigurable = (
    currentDashboard: InsightDashboard | undefined
): currentDashboard is CustomInsightDashboard =>
    !!currentDashboard && isRealDashboard(currentDashboard) && isCustomInsightDashboard(currentDashboard)
