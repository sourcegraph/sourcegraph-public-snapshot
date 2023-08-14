import { useMemo } from 'react'

import type { OrgSettingFields, UserSettingFields } from '@sourcegraph/shared/src/graphql-operations'

import type { AuthenticatedUser } from '../../../auth'
import type { Scalars } from '../../../graphql-operations'

export interface UseNamespacesResult {
    userNamespace: UserSettingFields
    namespaces: (UserSettingFields | OrgSettingFields)[]
    defaultSelectedNamespace: UserSettingFields | OrgSettingFields
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
    ) as UseNamespacesResult['namespaces']

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
