import { InsightsPermissionGrantsInput } from '../../../../../graphql-operations'
import { DashboardCreateInput } from '../code-insights-backend-types'

/**
 * Helper function to parse a grants object from a given type and visibility.
 * TODO: Remove this function when settings api is deprecated
 *
 * @param input {object} - A DashboardCreateInput object
 * @param input.type {('personal'|'organization'|'global')} - The type of the dashboard
 * @param input.visibility {string} - Usually the user or organization id
 * @returns - A properly formatted grants object
 */
export const createDashboardGrants = (input: DashboardCreateInput): InsightsPermissionGrantsInput => {
    const { type, visibility } = input
    const grants: InsightsPermissionGrantsInput = {
        users: [],
        organizations: [],
        global: false,
    }

    if (type === 'personal') {
        grants.users = [visibility]
    }

    if (type === 'organization') {
        grants.organizations = [visibility]
    }

    if (type === 'global') {
        grants.global = true
    }

    return grants
}
