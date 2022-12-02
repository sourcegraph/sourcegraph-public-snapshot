import { useMemo } from 'react'

import { ApolloError } from '@apollo/client'

import { isErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import {
    SettingsOrgSubject,
    SettingsUserSubject,
    SettingsSubject,
    SettingsCascadeOrError,
} from '@sourcegraph/shared/src/settings/settings'

import { Scalars, GetOrganizationsResult, GetOrganizationsVariables } from '../../../graphql-operations'

import { GET_ORGANIZATIONS, } from './backend'

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
 * @param settingsCascade The current user's `Settings`.
 * @param initialNamespaceID The id of the initial namespace to select.
 */
export const useNamespaces = (
    settingsCascade: SettingsCascadeOrError<Settings>,
    initialNamespaceID?: Scalars['ID']
): UseNamespacesResult => {
    // Gather all the available namespaces from the settings subjects.
    const rawNamespaces: SettingsSubject[] = useMemo(
        () =>
            (settingsCascade !== null &&
                !isErrorLike(settingsCascade) &&
                settingsCascade.subjects !== null &&
                settingsCascade.subjects.map(({ subject }) => subject).filter(subject => !isErrorLike(subject))) ||
            [],
        [settingsCascade]
    )

    const userNamespace = useMemo(
        () => rawNamespaces.find((namespace): namespace is SettingsUserSubject => namespace.__typename === 'User'),
        [rawNamespaces]
    )

    if (!userNamespace) {
        throw new Error('No user namespace found')
    }

    const { loading, data, error } = useQuery<GetOrganizationsResult, GetOrganizationsVariables>(GET_ORGANIZATIONS, {
        fetchPolicy: 'cache-and-network'
    })

    const organizationNamespaces: SettingsOrgSubject[] = useMemo(
        () => {
            if (!loading && data?.organizations) {
                return data.organizations.nodes
            }
            return []
        },
        [data, loading]
    )

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
