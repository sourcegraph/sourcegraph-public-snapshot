import { InsightsDashboardScope } from '../../types'

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
export const parseDashboardScope = (grants?: {
    global?: boolean
    users?: string[]
    organizations?: string[]
}): InsightsDashboardScope => {
    if (grants?.global) {
        return InsightsDashboardScope.Global
    }
    if (grants?.organizations?.length) {
        return InsightsDashboardScope.Organization
    }
    return InsightsDashboardScope.Personal
}
