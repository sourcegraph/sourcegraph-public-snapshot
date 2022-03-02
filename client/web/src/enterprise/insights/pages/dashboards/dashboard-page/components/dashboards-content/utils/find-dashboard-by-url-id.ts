import { InsightDashboard, isVirtualDashboard } from '../../../../../../core/types'

/**
 * Returns dashboard configurations by URL query param - dashboardID.
 *
 * @param dashboards - list of all reachable dashboards
 * @param dashboardID - possible dashboard id from the URL query param.
 */
export function findDashboardByUrlId(
    dashboards: InsightDashboard[],
    dashboardID: string
): InsightDashboard | undefined {
    return dashboards.find(dashboard => {
        if (isVirtualDashboard(dashboard)) {
            return dashboard.id === dashboardID.toLowerCase()
        }

        return dashboard.id === dashboardID || dashboard.title.toLowerCase() === dashboardID?.toLowerCase()
    })
}
