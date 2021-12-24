import { useMemo } from 'react'

import { isErrorLike } from '@sourcegraph/common'
import {
    SettingsOrgSubject,
    SettingsUserSubject,
    SettingsSubject,
    SettingsCascadeOrError,
} from '@sourcegraph/shared/src/settings/settings'

import { Scalars } from '../../../graphql-operations'
import { Settings } from '../../../schema/settings.schema'

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

    const organizationNamespaces = useMemo(
        () => rawNamespaces.filter((namespace): namespace is SettingsOrgSubject => namespace.__typename === 'Org'),
        [rawNamespaces]
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
    }
}
