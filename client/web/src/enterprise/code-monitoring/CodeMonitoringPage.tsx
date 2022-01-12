import classNames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useMemo, useEffect, useState } from 'react'
import { of } from 'rxjs'
import { catchError, map, startWith } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageHeader, LoadingSpinner, Button } from '@sourcegraph/wildcard'

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
    authenticatedUser: AuthenticatedUser | null
    fetchUserCodeMonitors?: typeof _fetchUserCodeMonitors
    toggleCodeMonitorEnabled?: typeof _toggleCodeMonitorEnabled

    // For testing purposes only
    testForceTab?: 'list' | 'getting-started'
}

export const CodeMonitoringPage: React.FunctionComponent<CodeMonitoringPageProps> = ({
    settingsCascade,
    authenticatedUser,
    fetchUserCodeMonitors = _fetchUserCodeMonitors,
    toggleCodeMonitorEnabled = _toggleCodeMonitorEnabled,
    isLightTheme,
    testForceTab,
}) => {
    useEffect(() => eventLogger.logViewEvent('CodeMonitoringPage'), [])

    const LOADING = 'loading' as const

    const userHasCodeMonitors = useObservable(
        useMemo(
            () =>
                authenticatedUser
                    ? fetchUserCodeMonitors({
                          id: authenticatedUser.id,
                          first: 1,
                          after: null,
                      }).pipe(
                          map(monitors => monitors.nodes.length > 0),
                          startWith(LOADING),
                          catchError(error => [asError(error)])
                      )
                    : of(false),
            [authenticatedUser, fetchUserCodeMonitors]
        )
    )

    const [currentTab, setCurrentTab] = useState<'list' | 'getting-started'>('list')

    // If user has no code monitors, default to the getting started tab after loading
    useEffect(() => {
        if (userHasCodeMonitors === false) {
            setCurrentTab('getting-started')
        }
    }, [userHasCodeMonitors])

    // Force tab for testing
    useEffect(() => {
        if (testForceTab && testForceTab !== currentTab) {
            setCurrentTab(testForceTab)
        }
    }, [currentTab, testForceTab])

    const showList = userHasCodeMonitors !== 'loading' && !isErrorLike(userHasCodeMonitors) && currentTab === 'list'

    return (
        <div className="code-monitoring-page" data-testid="code-monitoring-page">
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
                    !isErrorLike(userHasCodeMonitors) &&
                    authenticatedUser && (
                        <Button to="/code-monitoring/new" variant="primary" as={Link}>
                            <PlusIcon className="icon-inline" />
                            Create
                        </Button>
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

            {userHasCodeMonitors === 'loading' ? (
                <LoadingSpinner inline={false} />
            ) : (
                <div className="d-flex flex-column">
                    <div className="code-monitoring-page-tabs mb-4">
                        <div className="nav nav-tabs">
                            <div className="nav-item">
                                {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                                <a
                                    href=""
                                    onClick={event => {
                                        event.preventDefault()
                                        setCurrentTab('list')
                                    }}
                                    className={classNames('nav-link', currentTab === 'list' && 'active')}
                                    role="button"
                                >
                                    <span className="text-content" data-tab-content="Code monitors">
                                        Code monitors
                                    </span>
                                </a>
                            </div>
                            <div className="nav-item">
                                {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                                <a
                                    href=""
                                    onClick={event => {
                                        event.preventDefault()
                                        setCurrentTab('getting-started')
                                    }}
                                    className={classNames('nav-link', currentTab === 'getting-started' && 'active')}
                                    role="button"
                                >
                                    <span className="text-content" data-tab-content="Getting started">
                                        Getting started
                                    </span>
                                </a>
                            </div>
                        </div>
                    </div>

                    {currentTab === 'getting-started' && (
                        <CodeMonitoringGettingStarted isLightTheme={isLightTheme} isSignedIn={!!authenticatedUser} />
                    )}

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
