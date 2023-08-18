import {
    type CustomInsightDashboard,
    type InsightDashboard,
    isGlobalDashboard,
    isOrganizationDashboard,
    isPersonalDashboard,
    isVirtualDashboard,
} from '../../../../../core'

export const getDashboardOwnerName = (dashboard?: InsightDashboard): string => {
    if (!dashboard || isVirtualDashboard(dashboard)) {
        return ''
    }

    if (isPersonalDashboard(dashboard)) {
        return 'Private'
    }

    if (isGlobalDashboard(dashboard)) {
        return 'Global'
    }

    return dashboard.owners.map(owner => owner.title).join(' ')
}

interface DashboardOrganizationGroup {
    id: string
    name: string
    dashboards: CustomInsightDashboard[]
}

/**
 * Returns organization dashboards grouped by dashboard owner id
 */
export const getDashboardOrganizationsGroups = (dashboards: CustomInsightDashboard[]): DashboardOrganizationGroup[] => {
    const groupsDictionary = dashboards
        .filter(isOrganizationDashboard)
        .reduce<Record<string, DashboardOrganizationGroup>>((store, dashboard) => {
            for (const owner of dashboard.owners) {
                if (!store[owner.id]) {
                    store[owner.id] = {
                        id: owner.id,
                        name: owner.title,
                        dashboards: [],
                    }
                }

                store[owner.id].dashboards.push(dashboard)
            }

            return store
        }, {})

    return Object.values(groupsDictionary)
}
