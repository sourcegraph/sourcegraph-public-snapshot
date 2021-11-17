import {
    InsightDashboard,
    isGlobalDashboard,
    isPersonalDashboard,
    isVirtualDashboard,
    RealInsightDashboard,
} from '../../../../../../core/types'
import { isBuiltInInsightDashboard } from '../../../../../../core/types/dashboard/real-dashboard'

/**
 * Get formatted dashboard title for the dashboard select option.
 */
export const getDashboardTitle = (dashboard: InsightDashboard): string => {
    if (isVirtualDashboard(dashboard)) {
        return dashboard.title
    }

    if (isBuiltInInsightDashboard(dashboard)) {
        if (isGlobalDashboard(dashboard)) {
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
    if (isPersonalDashboard(dashboard)) {
        return 'Private'
    }

    if (isGlobalDashboard(dashboard)) {
        return 'Global'
    }

    return dashboard.owner?.name ?? dashboard.grants?.organizations?.[0] ?? 'Unknown'
}
