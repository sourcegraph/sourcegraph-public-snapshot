import { InsightsDashboardType, RealInsightDashboard } from '../../../../../../core/types'

/**
 * Get formatted dashboard title for the dashboard select option.
 */
export const getDashboardTitle = (dashboard: RealInsightDashboard): string => {
    const { builtIn } = dashboard

    if (builtIn) {
        if (dashboard.type === InsightsDashboardType.Global) {
            return 'Global Insights'
        }

        return `${dashboard.owner!.name}'s Insights`
    }

    return dashboard.title
}

/**
 * Get formatted dashboard owner name. Used for list option badge element.
 */
export const getDashboardOwnerName = (dashboard: RealInsightDashboard): string => {
    const { type } = dashboard

    if (type === InsightsDashboardType.Personal || dashboard.grants?.users?.length) {
        return 'Private'
    }

    if (type === InsightsDashboardType.Global || dashboard.grants?.global) {
        return 'Global'
    }

    return dashboard.owner?.name || dashboard.grants?.organizations?.[0] || 'Unknown'
}
