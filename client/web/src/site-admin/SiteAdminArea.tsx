import React, { useLayoutEffect, useMemo, useRef } from 'react'

import classNames from 'classnames'
import { isEqual } from 'lodash'
import ChartLineVariantIcon from 'mdi-react/ChartLineVariantIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch, useLocation } from 'react-router'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import * as GQL from '@sourcegraph/shared/src/schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { PageHeader, LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { BatchChangesProps } from '../batches'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { HeroPage } from '../components/HeroPage'
import { Page } from '../components/Page'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { RouteDescriptor } from '../util/contributions'

import { overviewGroup } from './sidebaritems'
import { SiteAdminSidebar, SiteAdminSideBarGroup, SiteAdminSideBarGroups } from './SiteAdminSidebar'

import styles from './SiteAdminArea.module.scss'

const NotFoundPage: React.ComponentType<React.PropsWithChildren<{}>> = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested site admin page was not found."
    />
)

const NotSiteAdminPage: React.ComponentType<React.PropsWithChildren<{}>> = () => (
    <HeroPage icon={MapSearchIcon} title="403: Forbidden" subtitle="Only site admins are allowed here." />
)

export interface SiteAdminAreaRouteContext
    extends PlatformContextProps,
        SettingsCascadeProps,
        ActivationProps,
        BatchChangesProps,
        TelemetryProps {
    site: Pick<GQL.ISite, '__typename' | 'id'>
    authenticatedUser: AuthenticatedUser
    isLightTheme: boolean
    isSourcegraphDotCom: boolean

    /** This property is only used by {@link SiteAdminOverviewPage}. */
    overviewComponents: readonly React.ComponentType<React.PropsWithChildren<{}>>[]
}

export interface SiteAdminAreaRoute extends RouteDescriptor<SiteAdminAreaRouteContext> {}

interface SiteAdminAreaProps
    extends RouteComponentProps<{}>,
        PlatformContextProps,
        SettingsCascadeProps,
        ActivationProps,
        BatchChangesProps,
        TelemetryProps {
    routes: readonly SiteAdminAreaRoute[]
    sideBarGroups: SiteAdminSideBarGroups
    overviewComponents: readonly React.ComponentType<React.PropsWithChildren<unknown>>[]
    authenticatedUser: AuthenticatedUser
    isLightTheme: boolean
    isSourcegraphDotCom: boolean
}

export const analyticsGroup: SiteAdminSideBarGroup = {
    header: {
        label: 'Analytics',
        icon: ChartLineVariantIcon,
    },
    items: [
        {
            label: 'Overview',
            to: '/site-admin/',
            exact: true,
        },
        {
            label: 'Search',
            to: '/site-admin/analytics/search',
        },
        {
            label: 'Code intel',
            to: '/site-admin/analytics/code-intel',
        },
        {
            label: 'Users',
            to: '/site-admin/analytics/users',
        },
        {
            label: 'Batch changes',
            to: '/site-admin/analytics/batch-changes',
        },
        {
            label: 'Notebooks',
            to: '/site-admin/analytics/notebooks',
        },
        {
            label: 'Extensions',
            to: '/site-admin/analytics/extensions',
        },
        {
            label: 'Feedback survey',
            to: '/site-admin/surveys',
        },
        {
            label: 'Code insights (soon)',
            to: '/site-admin/analytics/code-insights',
        },
    ],
}

export const analyticsRoutes: readonly SiteAdminAreaRoute[] = [
    {
        path: '/analytics/search',
        render: lazyComponent(() => import('./analytics/AnalyticsSearchPage'), 'AnalyticsSearchPage'),
        exact: true,
    },
    {
        path: '/analytics/code-intel',
        render: lazyComponent(() => import('./analytics/AnalyticsCodeIntelPage'), 'AnalyticsCodeIntelPage'),
        exact: true,
    },
    {
        path: '/analytics/extensions',
        render: lazyComponent(() => import('./analytics/AnalyticsExtensionsPage'), 'AnalyticsExtensionsPage'),
        exact: true,
    },
    {
        path: '/analytics/users',
        render: lazyComponent(() => import('./analytics/AnalyticsUsersPage'), 'AnalyticsUsersPage'),
        exact: true,
    },
    {
        path: '/analytics/code-insights',
        render: lazyComponent(() => import('./analytics/AnalyticsComingSoonPage'), 'AnalyticsComingSoonPage'),
        exact: true,
    },
    {
        path: '/analytics/batch-changes',
        render: lazyComponent(() => import('./analytics/AnalyticsBatchChangesPage'), 'AnalyticsBatchChangesPage'),
        exact: true,
    },
    {
        path: '/analytics/notebooks',
        render: lazyComponent(() => import('./analytics/AnalyticsNotebooksPage'), 'AnalyticsNotebooksPage'),
        exact: true,
    },
    {
        path: '/analytics/extensions',
        render: lazyComponent(() => import('./analytics/AnalyticsComingSoonPage'), 'AnalyticsComingSoonPage'),
        exact: true,
    },
    {
        path: '/',
        render: lazyComponent(() => import('./analytics/AnalyticsOverviewPage'), 'AnalyticsOverviewPage'),
        exact: true,
    },
]

const AuthenticatedSiteAdminArea: React.FunctionComponent<React.PropsWithChildren<SiteAdminAreaProps>> = props => {
    const { pathname } = useLocation()

    const reference = useRef<HTMLDivElement>(null)

    useLayoutEffect(() => {
        if (reference.current) {
            reference.current.scrollIntoView()
        }
    }, [pathname])

    const [isAdminAnalyticsDisabled] = useFeatureFlag('admin-analytics-disabled', false)

    const adminSideBarGroups = useMemo(() => {
        if (isAdminAnalyticsDisabled) {
            return props.sideBarGroups
        }
        return [analyticsGroup, ...props.sideBarGroups.filter(group => !isEqual(group, overviewGroup))]
    }, [isAdminAnalyticsDisabled, props.sideBarGroups])

    const routes = useMemo(() => (!isAdminAnalyticsDisabled ? [...analyticsRoutes, ...props.routes] : props.routes), [
        isAdminAnalyticsDisabled,
        props.routes,
    ])

    // If not site admin, redirect to sign in.
    if (!props.authenticatedUser.siteAdmin) {
        return <NotSiteAdminPage />
    }

    const context: SiteAdminAreaRouteContext = {
        authenticatedUser: props.authenticatedUser,
        platformContext: props.platformContext,
        settingsCascade: props.settingsCascade,
        isLightTheme: props.isLightTheme,
        isSourcegraphDotCom: props.isSourcegraphDotCom,
        batchChangesEnabled: props.batchChangesEnabled,
        batchChangesExecutionEnabled: props.batchChangesExecutionEnabled,
        batchChangesWebhookLogsEnabled: props.batchChangesWebhookLogsEnabled,
        activation: props.activation,
        site: { __typename: 'Site' as const, id: window.context.siteGQLID },
        overviewComponents: props.overviewComponents,
        telemetryService: props.telemetryService,
    }

    return (
        <Page>
            <PageHeader path={[{ text: 'Admin' }]} />
            <div className="d-flex my-3" ref={reference}>
                <SiteAdminSidebar
                    className={classNames('flex-0 mr-3', styles.sidebar)}
                    groups={adminSideBarGroups}
                    isSourcegraphDotCom={props.isSourcegraphDotCom}
                    batchChangesEnabled={props.batchChangesEnabled}
                    batchChangesExecutionEnabled={props.batchChangesExecutionEnabled}
                    batchChangesWebhookLogsEnabled={props.batchChangesWebhookLogsEnabled}
                />
                <div className="flex-bounded">
                    <ErrorBoundary location={props.location}>
                        <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
                            <Switch>
                                {routes.map(
                                    ({ render, path, exact, condition = () => true }) =>
                                        condition(context) && (
                                            <Route
                                                // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                                key="hardcoded-key"
                                                path={props.match.url + path}
                                                exact={exact}
                                                render={routeComponentProps =>
                                                    render({ ...context, ...routeComponentProps })
                                                }
                                            />
                                        )
                                )}
                                <Route component={NotFoundPage} />
                            </Switch>
                        </React.Suspense>
                    </ErrorBoundary>
                </div>
            </div>
        </Page>
    )
}

/**
 * Renders a layout of a sidebar and a content area to display site admin information.
 */
export const SiteAdminArea = withAuthenticatedUser(AuthenticatedSiteAdminArea)
