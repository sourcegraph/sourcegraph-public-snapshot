import { useMemo } from 'react'
import { useLocation } from 'react-router'

import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import {
    SettingsOrgSubject,
    SettingsUserSubject,
    SettingsSubject,
    SettingsCascadeOrError,
} from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

export interface UseNamespacesResult {
    userNamespace: SettingsUserSubject
    namespaces: (SettingsUserSubject | SettingsOrgSubject)[]
    defaultSelectedNamespace: SettingsUserSubject | SettingsOrgSubject
}

/**
 * Custom hook to extract namespaces from the provided `settingsCascade` and determine the
 * appropriate default namespace to select for the user.
 *
 * @param settingsCascade The current user's `Settings`.
 */
export const useNamespaces = (settingsCascade: SettingsCascadeOrError<Settings>): UseNamespacesResult => {
    const location = useLocation()

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

    const organizationNamespaces = useMemo(
        () => rawNamespaces.filter((namespace): namespace is SettingsOrgSubject => namespace.__typename === 'Org'),
        [rawNamespaces]
    )

    const namespaces: (SettingsUserSubject | SettingsOrgSubject)[] = useMemo(
        () => [userNamespace, ...organizationNamespaces],
        [userNamespace, organizationNamespaces]
    )

    // Check if there's a namespace parameter in the URL.
    const defaultNamespace = new URLSearchParams(location.search).get('namespace')

    // The default namespace selected from the dropdown should match whatever was in the
    // URL parameter, or else default to the user's namespace.
    const defaultSelectedNamespace = useMemo(() => {
        if (defaultNamespace) {
            const lowerCaseDefaultNamespace = defaultNamespace.toLowerCase()
            return (
                namespaces.find(
                    namespace =>
                        namespace.displayName?.toLowerCase() === lowerCaseDefaultNamespace ||
                        (namespace.__typename === 'User' &&
                            namespace.username.toLowerCase() === lowerCaseDefaultNamespace) ||
                        (namespace.__typename === 'Org' && namespace.name.toLowerCase() === lowerCaseDefaultNamespace)
                ) || userNamespace
            )
        }
        return userNamespace
    }, [namespaces, defaultNamespace, userNamespace])

    return {
        userNamespace,
        namespaces,
        defaultSelectedNamespace,
    }
}
