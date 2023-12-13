import React, { useMemo, useEffect, useState, useLayoutEffect } from 'react'

import { mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { of } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import type { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { type SettingsCascadeProps, useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    PageHeader,
    LoadingSpinner,
    useObservable,
    Button,
    Link,
    ProductStatusBadge,
    Icon,
} from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { CodeMonitoringLogo } from '../../code-monitoring/CodeMonitoringLogo'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'

import {
    fetchUserCodeMonitors as _fetchUserCodeMonitors,
    fetchCodeMonitors as _fetchCodeMonitors,
    toggleCodeMonitorEnabled as _toggleCodeMonitorEnabled,
} from './backend'
import { CodeMonitoringGettingStarted } from './CodeMonitoringGettingStarted'
import { CodeMonitoringLogs } from './CodeMonitoringLogs'
import { CodeMonitorList } from './CodeMonitorList'

export interface CodeMonitoringPageProps extends SettingsCascadeProps<Settings>, TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
    fetchUserCodeMonitors?: typeof _fetchUserCodeMonitors
    fetchCodeMonitors?: typeof _fetchCodeMonitors
    toggleCodeMonitorEnabled?: typeof _toggleCodeMonitorEnabled
    isCodyApp: boolean
    // For testing purposes only
    testForceTab?: 'list' | 'getting-started' | 'logs'
}

export const CodeMonitoringPage: React.FunctionComponent<React.PropsWithChildren<CodeMonitoringPageProps>> = ({
    authenticatedUser,
    fetchUserCodeMonitors = _fetchUserCodeMonitors,
    fetchCodeMonitors = _fetchCodeMonitors,
    toggleCodeMonitorEnabled = _toggleCodeMonitorEnabled,
    testForceTab,
    isCodyApp,
    telemetryRecorder,
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
                    telemetryRecorder.recordEvent('codeMonitoringGettingStarted', 'viewed')
                    eventLogger.logPageView('CodeMonitoringGettingStartedPage')
                    break
                case 'logs':
                    telemetryRecorder.recordEvent('codeMonitoringLogs', 'viewed')
                    eventLogger.logPageView('CodeMonitoringLogsPage')
                    break
                case 'list':
                    telemetryRecorder.recordEvent('codeMonitoring', 'viewed')
                    eventLogger.logPageView('CodeMonitoringPage')
            }
        }
    }, [currentTab, userHasCodeMonitors, window.context.telemetryRecorder])

    const showList = userHasCodeMonitors !== undefined && !isErrorLike(userHasCodeMonitors) && currentTab === 'list'

    const showLogsTab =
        useExperimentalFeatures(features => features.showCodeMonitoringLogs) && authenticatedUser && !isCodyApp

    return (
        <div className="code-monitoring-page" data-testid="code-monitoring-page">
            <PageTitle title="Code Monitoring" />
            <PageHeader
                actions={
                    authenticatedUser &&
                    !isCodyApp && (
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
                            {!isCodyApp && (
                                <div className="nav-item">
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
                            )}
                            <div className="nav-item">
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
                        <CodeMonitoringGettingStarted
                            authenticatedUser={authenticatedUser}
                            isCodyApp={isCodyApp}
                            telemetryRecorder={telemetryRecorder}
                        />
                    )}

                    {currentTab === 'logs' && <CodeMonitoringLogs />}

                    {showList && (
                        <CodeMonitorList
                            authenticatedUser={authenticatedUser}
                            fetchUserCodeMonitors={fetchUserCodeMonitors}
                            fetchCodeMonitors={fetchCodeMonitors}
                            toggleCodeMonitorEnabled={toggleCodeMonitorEnabled}
                            telemetryRecorder={telemetryRecorder}
                        />
                    )}
                </div>
            )}
        </div>
    )
}
