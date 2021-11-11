import { InsightsDashboardType } from '../../types'

/**
 * Helper function to parse the dashboard type from the grants object.
 * TODO: Remove this function when settings api is deprecated
 *
 * @param grants {object} - A grants object from an insight dashboard
 * @param grants.global {boolean}
 * @param grants.users {string[]}
 * @param grants.organizations {string[]}
 * @returns - The type of the dashboard
 */
export const parseDashboardType = (grants?: {
    global?: boolean
    users?: string[]
    organizations?: string[]
}): InsightsDashboardType.Personal | InsightsDashboardType.Organization | InsightsDashboardType.Global => {
    if (grants?.global) {
        return InsightsDashboardType.Global
    }
    if (grants?.organizations?.length) {
        return InsightsDashboardType.Organization
    }
    return InsightsDashboardType.Personal
}
