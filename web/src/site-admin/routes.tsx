import React from 'react'
import { Redirect } from 'react-router'
import { eventLogger } from '../tracking/eventLogger'
import { lazyComponent } from '../util/lazyComponent'
import { SiteAdminAreaRoute } from './SiteAdminArea'

const SiteAdminAddExternalServicesPage = lazyComponent(
    () => import('./SiteAdminAddExternalServicesPage'),
    'SiteAdminAddExternalServicesPage'
)

export const siteAdminAreaRoutes: readonly SiteAdminAreaRoute[] = [
    {
        // Render empty page if no page selected
        path: '',
        render: lazyComponent(() => import('./overview/SiteAdminOverviewPage'), 'SiteAdminOverviewPage'),
        exact: true,
    },
    {
        path: '/configuration',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminConfigurationPage'), 'SiteAdminConfigurationPage'),
    },
    {
        path: '/global-settings',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminSettingsPage'), 'SiteAdminSettingsPage'),
    },
    {
        path: '/external-services',
        render: lazyComponent(() => import('./SiteAdminExternalServicesPage'), 'SiteAdminExternalServicesPage'),
        exact: true,
    },
    {
        path: '/external-services/add',
        render: () => <Redirect to="new" />,
        exact: true,
    },
    {
        path: '/external-services/new',
        render: props => <SiteAdminAddExternalServicesPage {...props} eventLogger={eventLogger} />,
        exact: true,
    },
    {
        path: '/external-services/:id',
        render: lazyComponent(() => import('./SiteAdminExternalServicePage'), 'SiteAdminExternalServicePage'),
        exact: true,
    },
    {
        path: '/repositories',
        render: lazyComponent(() => import('./SiteAdminRepositoriesPage'), 'SiteAdminRepositoriesPage'),
        exact: true,
    },
    {
        path: '/organizations',
        render: lazyComponent(() => import('./SiteAdminOrgsPage'), 'SiteAdminOrgsPage'),
        exact: true,
    },
    {
        path: '/users',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminAllUsersPage'), 'SiteAdminAllUsersPage'),
    },
    {
        path: '/users/new',
        render: lazyComponent(() => import('./SiteAdminCreateUserPage'), 'SiteAdminCreateUserPage'),
        exact: true,
    },
    {
        path: '/tokens',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminTokensPage'), 'SiteAdminTokensPage'),
    },
    {
        path: '/usage-statistics',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminUsageStatisticsPage'), 'SiteAdminUsageStatisticsPage'),
    },
    {
        path: '/updates',
        render: lazyComponent(() => import('./SiteAdminUpdatesPage'), 'SiteAdminUpdatesPage'),
        exact: true,
    },
    {
        path: '/pings',
        render: lazyComponent(() => import('./SiteAdminPingsPage'), 'SiteAdminPingsPage'),
        exact: true,
    },
    {
        path: '/report-bug',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminReportBugPage'), 'SiteAdminReportBugPage'),
    },
    {
        path: '/surveys',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminSurveyResponsesPage'), 'SiteAdminSurveyResponsesPage'),
    },
]
