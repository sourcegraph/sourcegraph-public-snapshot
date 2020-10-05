import React, { useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { isSettingsValid, SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { TelemetryProps } from '../../../shared/src/telemetry/telemetryService'
import { Settings } from '../schema/settings.schema'

interface Props extends RouteComponentProps<{ name: string }>, SettingsCascadeProps<Settings>, TelemetryProps {}

/**
 * An experimental repository owner (user or organization) profile page, only shown on
 * Sourcegraph.com currently. It lets you search across all of the repository owner's repositories
 * (which is defined currently as all repositories in the repogroup with the same name as the
 * organization).
 */
export const NamespaceProfilePage: React.FunctionComponent<Props> = ({ match, settingsCascade, ...props }) => {
    const repositoryOwnerName = match.params.name || 'cncf'

    useEffect(() => props.telemetryService.logViewEvent(`RepositoryOwnerProfilePage:${repositoryOwnerName}`), [
        repositoryOwnerName,
        props.telemetryService,
    ])

    /* const repogroup = useMemo(
        () =>
            (isSettingsValid<Settings>(settingsCascade) &&
                settingsCascade.final['search.repositoryGroups']?.[organizationName]) ||
            null,
        [organizationName, settingsCascade]
    ) */
    // const repositories

    return <div className="container">Repository owner {repositoryOwnerName}</div>
}
