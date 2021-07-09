import { InsightDashboard, isOrganizationDashboard, RealInsightDashboard } from '../../../../../../core/types';

interface DashboardOrganizationGroup {
    id: string
    name: string
    dashboards: RealInsightDashboard[]
}

/**
 * Returns organization dashboards grouped by dashboard owner id
 */
export const getGroupedOrganizationDashboards = (dashboards: InsightDashboard[]): DashboardOrganizationGroup[] => {
    const groupsDictionary = dashboards
        .filter(isOrganizationDashboard)
        .reduce<Record<string, DashboardOrganizationGroup>>((store, dashboard) => {
            if (!store[dashboard.owner.id]) {
                store[dashboard.owner.id] = {
                    id: dashboard.owner.id,
                    name: dashboard.owner.name,
                    dashboards: [],
                }
            }

            store[dashboard.owner.id].dashboards.push(dashboard)

            return store
        }, {})

    return Object.values(groupsDictionary)
}

