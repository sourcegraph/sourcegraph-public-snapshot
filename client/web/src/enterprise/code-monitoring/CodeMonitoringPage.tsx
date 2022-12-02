import React, { useMemo, useEffect, useState, useLayoutEffect } from 'react'

import { mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { of } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    PageHeader,
    LoadingSpinner,
    useObservable,
    Button,
    Link,
    ProductStatusBadge,
    Icon,
} from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CodeMonitoringLogo } from '../../code-monitoring/CodeMonitoringLogo'
import { PageTitle } from '../../components/PageTitle'
import { useExperimentalFeatures } from '../../stores'
import { eventLogger } from '../../tracking/eventLogger'

import {
    fetchUserCodeMonitors as _fetchUserCodeMonitors,
    toggleCodeMonitorEnabled as _toggleCodeMonitorEnabled,
} from './backend'
import { CodeMonitoringGettingStarted } from './CodeMonitoringGettingStarted'
import { CodeMonitoringLogs } from './CodeMonitoringLogs'
import { CodeMonitorList } from './CodeMonitorList'

export interface CodeMonitoringPageProps extends SettingsCascadeProps<Settings>, ThemeProps {
    authenticatedUser: AuthenticatedUser | null
    fetchUserCodeMonitors?: typeof _fetchUserCodeMonitors
    toggleCodeMonitorEnabled?: typeof _toggleCodeMonitorEnabled

    // For testing purposes only
    testForceTab?: 'list' | 'getting-started' | 'logs'
}

export const CodeMonitoringPage: React.FunctionComponent<React.PropsWithChildren<CodeMonitoringPageProps>> = ({
    authenticatedUser,
    fetchUserCodeMonitors = _fetchUserCodeMonitors,
    toggleCodeMonitorEnabled = _toggleCodeMonitorEnabled,
    isLightTheme,
    testForceTab,
}) => {
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
                          catchError(error => [asError(error)])
                      )
                    : of(false),
            [authenticatedUser, fetchUserCodeMonitors]
        )
    )

    const [currentTab, setCurrentTab] = useState<'list' | 'getting-started' | 'logs' | null>(null)

    // Select the appropriate tab after loading:
    // - If the user has code monitors, show the list tab
    // - If the user has no code monitors, show the getting started tab
    useLayoutEffect(() => {
        if (userHasCodeMonitors === true) {
            setCurrentTab('list')
        } else if (userHasCodeMonitors === false) {
            setCurrentTab('getting-started')
        }
    }, [userHasCodeMonitors])

    // Force tab for testing
    useLayoutEffect(() => {
        if (testForceTab && testForceTab !== currentTab) {
            setCurrentTab(testForceTab)
        }
    }, [currentTab, testForceTab])

    // Log page view of selected tab
    useEffect(() => {
        if (userHasCodeMonitors !== undefined) {
            switch (currentTab) {
                case 'getting-started':
                    eventLogger.logPageView('CodeMonitoringGettingStartedPage')
                    break
                case 'logs':
                    eventLogger.logPageView('CodeMonitoringLogsPage')
                    break
                case 'list':
                    eventLogger.logPageView('CodeMonitoringPage')
            }
        }
    }, [currentTab, userHasCodeMonitors])

    const showList = userHasCodeMonitors !== undefined && !isErrorLike(userHasCodeMonitors) && currentTab === 'list'

    const showLogsTab = useExperimentalFeatures(features => features.showCodeMonitoringLogs) && authenticatedUser

    return (
        <div className="code-monitoring-page" data-testid="code-monitoring-page">
            <PageTitle title="Code Monitoring" />
            <PageHeader
                actions={
                    authenticatedUser && (
                        <Button to="/code-monitoring/new" variant="primary" as={Link}>
                            <Icon aria-hidden={true} svgPath={mdiPlus} /> Create a code monitor
                        </Button>
                    )
                }
                description={
                    <>Watch your code for changes and trigger actions to get notifications, send webhooks, and more.</>
                }
                className="mb-3"
            >
                <PageHeader.Heading as="h2" styleAs="h1">
                    <PageHeader.Breadcrumb icon={CodeMonitoringLogo}>Code monitoring</PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>

            {userHasCodeMonitors === undefined ? (
                <LoadingSpinner inline={false} />
            ) : (
                <div className="d-flex flex-column">
                    <div className="code-monitoring-page-tabs mb-4">
                        <div className="nav nav-tabs">
                            <div className="nav-item">
                                {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                                <Link
                                    to=""
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
                                </Link>
                            </div>
                            <div className="nav-item">
                                {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                                <Link
                                    to=""
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
                                </Link>
                            </div>
                            {showLogsTab && (
                                <div className="nav-item">
                                    {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                                    <Link
                                        to=""
                                        onClick={event => {
                                            event.preventDefault()
                                            setCurrentTab('logs')
                                        }}
                                        className={classNames('nav-link flex-row', currentTab === 'logs' && 'active')}
                                        role="button"
                                    >
                                        <span className="text-content" data-tab-content="Logs">
                                            Logs
                                        </span>
                                        <ProductStatusBadge status="beta" className="ml-2" />
                                    </Link>
                                </div>
                            )}
                        </div>
                    </div>

                    {currentTab === 'getting-started' && (
                        <CodeMonitoringGettingStarted isLightTheme={isLightTheme} isSignedIn={!!authenticatedUser} />
                    )}

                    {currentTab === 'logs' && <CodeMonitoringLogs />}

                    {showList && (
                        <CodeMonitorList
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
