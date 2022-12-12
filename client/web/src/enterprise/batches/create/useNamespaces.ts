import { useMemo } from 'react'

import { SettingsOrgSubject, SettingsUserSubject } from '@sourcegraph/shared/src/settings/settings'

import { AuthenticatedUser } from '../../../auth'
import { Scalars } from '../../../graphql-operations'

type UserNamespace = Omit<SettingsUserSubject, 'viewerCanAdminister'>
type OrgNamespace = Omit<SettingsOrgSubject, 'viewerCanAdminister'>

export interface UseNamespacesResult {
    userNamespace: UserNamespace
    namespaces: (UserNamespace | OrgNamespace)[]
    defaultSelectedNamespace: UserNamespace | OrgNamespace
}

/**
 * Custom hook to extract namespaces from the provided `settingsCascade` and determine the
 * appropriate default namespace to select for the user.
 *
 * @param authenticatedUser     The currently signed-in user
 * @param initialNamespaceID    The id of the initial namespace to select.
 */
export const useNamespaces = (
    authenticatedUser: AuthenticatedUser | null,
    initialNamespaceID?: Scalars['ID']
): UseNamespacesResult => {
    if (!authenticatedUser) {
        throw new Error('No user found')
    }

    const { organizations, ...userDetails } = authenticatedUser

    const organizationNamespaces = organizations.nodes
    const userNamespace = userDetails

    const namespaces = useMemo(
        () => [userNamespace, ...organizationNamespaces],
        [userNamespace, organizationNamespaces]
    )

    // The default namespace selected from the dropdown should match whatever the initial
    // namespace was, or else default to the user's namespace.
    const defaultSelectedNamespace = useMemo(() => {
        if (initialNamespaceID) {
            return namespaces.find(namespace => namespace.id === initialNamespaceID) || userNamespace
        }
        return userNamespace
    }, [namespaces, initialNamespaceID, userNamespace])

    return {
        userNamespace,
        namespaces,
        defaultSelectedNamespace,
    }
}
