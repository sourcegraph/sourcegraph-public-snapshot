import { useMemo } from 'react'

import { ApolloError } from '@apollo/client'

import { useQuery } from '@sourcegraph/http-client'
import { SettingsOrgSubject, SettingsUserSubject } from '@sourcegraph/shared/src/settings/settings'
import { useObservable } from '@sourcegraph/wildcard'

import { authenticatedUser } from '../../../auth'
import { Scalars, GetUserOrganizationsResult, GetUserOrganizationsVariables } from '../../../graphql-operations'

import { GET_USER_ORGANIZATIONS } from './backend'

export interface UseNamespacesResult {
    userNamespace: SettingsUserSubject
    namespaces: (SettingsUserSubject | SettingsOrgSubject)[]
    defaultSelectedNamespace: SettingsUserSubject | SettingsOrgSubject
    loading: boolean
    error: ApolloError | undefined
}

/**
 * Custom hook to extract namespaces from the provided `settingsCascade` and determine the
 * appropriate default namespace to select for the user.
 *
 * @param initialNamespaceID The id of the initial namespace to select.
 */
export const useNamespaces = (initialNamespaceID?: Scalars['ID']): UseNamespacesResult => {
    const user = useObservable(authenticatedUser)

    if (!user) {
        throw new Error('No user found')
    }

    const localUserNamespace = useMemo(
        () => ({
            __typename: user.__typename,
            id: user.id,
            username: user?.username,
            displayName: user.displayName,
            viewerCanAdminister: user.viewerCanAdminister,
        }),
        [user]
    )

    const { loading, data, error } = useQuery<GetUserOrganizationsResult, GetUserOrganizationsVariables>(
        GET_USER_ORGANIZATIONS,
        {
            variables: {
                userId: user.id,
            },
            fetchPolicy: 'cache-and-network',
        }
    )

    if (data?.node?.__typename !== 'User') {
        throw new Error('No user found')
    }

    const userNamespace: SettingsUserSubject = useMemo(() => {
        const userData = data?.node
        if (!loading && userData?.__typename === 'User') {
            return userData
        }
        return localUserNamespace
    }, [data, loading, localUserNamespace])

    const organizationNamespaces: SettingsOrgSubject[] = useMemo(() => {
        const userData = data?.node
        if (!loading && userData?.__typename === 'User') {
            return userData.organizations.nodes
        }
        return []
    }, [data, loading])

    const namespaces: (SettingsUserSubject | SettingsOrgSubject)[] = useMemo(
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
        loading,
        error,
    }
}
