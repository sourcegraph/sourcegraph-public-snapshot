import {
    InsightDashboard,
    isGlobalDashboard,
    isPersonalDashboard,
} from '../../../../../../core/types'
import { CustomDashboardWithOwner } from '../types';

/**
 * Get formatted dashboard title for the dashboard select option.
 */
export const getDashboardTitle = (dashboard: InsightDashboard): string => dashboard.title

/**
 * Get formatted dashboard owner name. Used for list option badge element.
 */
export const getDashboardOwnerName = (dashboard: InsightDashboard | CustomDashboardWithOwner): string => {
    if (isPersonalDashboard(dashboard)) {
        return 'Private'
    }

    if (isGlobalDashboard(dashboard)) {
        return 'Global'
    }

    return ''
}
