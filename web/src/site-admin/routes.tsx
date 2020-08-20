import React from 'react'
import { Redirect } from 'react-router'
import { lazyComponent } from '../util/lazyComponent'
import { SiteAdminAreaRoute } from './SiteAdminArea'
import { codeHostExternalServices, nonCodeHostExternalServices } from '../components/externalServices/externalServices'

const ExternalServicesPage = lazyComponent(
    () => import('../components/externalServices/ExternalServicesPage'),
    'ExternalServicesPage'
)
const AddExternalServicesPage = lazyComponent(
    () => import('../components/externalServices/AddExternalServicesPage'),
    'AddExternalServicesPage'
)
const ExternalServicePage = lazyComponent(
    () => import('../components/externalServices/ExternalServicePage'),
    'ExternalServicePage'
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
        render: props => (
            <ExternalServicesPage
                {...props}
                routingPrefix="/site-admin"
                afterDeleteRoute="/site-admin/repositories?repositoriesUpdated"
            />
        ),
        exact: true,
    },
    {
        path: '/external-services/add',
        render: () => <Redirect to="new" />,
        exact: true,
    },
    {
        path: '/external-services/new',
        render: props => (
            <AddExternalServicesPage
                {...props}
                routingPrefix="/site-admin"
                afterCreateRoute="/site-admin/repositories?repositoriesUpdated"
                codeHostExternalServices={codeHostExternalServices}
                nonCodeHostExternalServices={nonCodeHostExternalServices}
            />
        ),
        exact: true,
    },
    {
        path: '/external-services/:id',
        render: props => (
            <ExternalServicePage {...props} afterUpdateRoute="/site-admin/repositories?repositoriesUpdated" />
        ),
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
