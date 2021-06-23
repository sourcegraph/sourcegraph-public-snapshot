import PlusIcon from 'mdi-react/PlusIcon'
import React, { useMemo, useEffect } from 'react'
import { NavLink, Redirect } from 'react-router-dom'
import { catchError, map, startWith } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CodeMonitoringLogo } from '../../code-monitoring/CodeMonitoringLogo'
import { PageTitle } from '../../components/PageTitle'
import { Settings } from '../../schema/settings.schema'
import { eventLogger } from '../../tracking/eventLogger'

import {
    fetchUserCodeMonitors as _fetchUserCodeMonitors,
    toggleCodeMonitorEnabled as _toggleCodeMonitorEnabled,
} from './backend'
import { CodeMonitoringGettingStarted } from './CodeMonitoringGettingStarted'
import { CodeMonitorList } from './CodeMonitorList'

export interface CodeMonitoringPageProps extends SettingsCascadeProps<Settings>, ThemeProps {
    authenticatedUser: AuthenticatedUser
    fetchUserCodeMonitors?: typeof _fetchUserCodeMonitors
    toggleCodeMonitorEnabled?: typeof _toggleCodeMonitorEnabled
    showGettingStarted?: boolean
}

export const CodeMonitoringPage: React.FunctionComponent<CodeMonitoringPageProps> = ({
    settingsCascade,
    authenticatedUser,
    fetchUserCodeMonitors = _fetchUserCodeMonitors,
    toggleCodeMonitorEnabled = _toggleCodeMonitorEnabled,
    showGettingStarted = false,
    isLightTheme,
}) => {
    useEffect(() => eventLogger.logViewEvent('CodeMonitoringPage'), [])

    const LOADING = 'loading' as const

    const userHasCodeMonitors = useObservable(
        useMemo(
            () =>
                fetchUserCodeMonitors({
                    id: authenticatedUser.id,
                    first: 1,
                    after: null,
                }).pipe(
                    map(monitors => monitors.nodes.length > 0),
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [authenticatedUser.id, fetchUserCodeMonitors]
        )
    )

    // If user has no code monitors, redirect to the getting started page
    if (!showGettingStarted && userHasCodeMonitors === false) {
        return <Redirect to="/code-monitoring/getting-started" />
    }

    const showList =
        userHasCodeMonitors &&
        userHasCodeMonitors !== 'loading' &&
        !isErrorLike(userHasCodeMonitors) &&
        !showGettingStarted

    return (
        <div className="code-monitoring-page">
            <PageTitle title="Code Monitoring" />
            <PageHeader
                path={[
                    {
                        icon: CodeMonitoringLogo,
                        text: 'Code monitoring',
                    },
                ]}
                actions={
                    userHasCodeMonitors &&
                    userHasCodeMonitors !== 'loading' &&
                    !isErrorLike(userHasCodeMonitors) && (
                        <Link to="/code-monitoring/new" className="btn btn-primary">
                            <PlusIcon className="icon-inline" />
                            Create
                        </Link>
                    )
                }
                description={
                    userHasCodeMonitors &&
                    userHasCodeMonitors !== 'loading' &&
                    !isErrorLike(userHasCodeMonitors) && (
                        <>
                            Watch your code for changes and trigger actions to get notifications, send webhooks, and
                            more.
                        </>
                    )
                }
                className="mb-3"
            />
            {userHasCodeMonitors === 'loading' && <LoadingSpinner />}

            {(showGettingStarted || showList) && (
                <div className="d-flex flex-column">
                    <div className="code-monitoring-page-tabs mb-4">
                        <div className="nav nav-tabs">
                            <div className="nav-item">
                                <NavLink
                                    to="/code-monitoring"
                                    className="nav-link"
                                    activeClassName="active"
                                    exact={true}
                                >
                                    <span className="text-content" data-tab-content="Code monitors">
                                        Code monitors
                                    </span>
                                </NavLink>
                            </div>
                            <div className="nav-item">
                                <NavLink
                                    to="/code-monitoring/getting-started"
                                    className="nav-link"
                                    activeClassName="active"
                                    exact={true}
                                >
                                    <span className="text-content" data-tab-content="Getting started">
                                        Getting started
                                    </span>
                                </NavLink>
                            </div>
                        </div>
                    </div>

                    {showGettingStarted && <CodeMonitoringGettingStarted isLightTheme={isLightTheme} />}

                    {showList && (
                        <CodeMonitorList
                            settingsCascade={settingsCascade}
                            authenticatedUser={authenticatedUser}
                            fetchUserCodeMonitors={fetchUserCodeMonitors}
                            toggleCodeMonitorEnabled={toggleCodeMonitorEnabled}
                        />
                    )}
                </div>
            )}
        </div>
    )
}
