import React from 'react'
import { Redirect } from 'react-router'
import { eventLogger } from '../tracking/eventLogger'
import { asyncComponent } from '../util/asyncComponent'
import { SiteAdminAreaRoute } from './SiteAdminArea'

const SiteAdminAddExternalServicesPage = asyncComponent(
    () => import('./SiteAdminAddExternalServicesPage'),
    'SiteAdminAddExternalServicesPage'
)

export const siteAdminAreaRoutes: ReadonlyArray<SiteAdminAreaRoute> = [
    {
        // Render empty page if no page selected
        path: '',
        render: asyncComponent(() => import('./SiteAdminOverviewPage'), 'SiteAdminOverviewPage'),
        exact: true,
    },
    {
        path: '/configuration',
        exact: true,
        render: asyncComponent(() => import('./SiteAdminConfigurationPage'), 'SiteAdminConfigurationPage'),
    },
    {
        path: '/global-settings',
        exact: true,
        render: asyncComponent(() => import('./SiteAdminSettingsPage'), 'SiteAdminSettingsPage'),
    },
    {
        path: '/external-services',
        render: asyncComponent(() => import('./SiteAdminExternalServicesPage'), 'SiteAdminExternalServicesPage'),
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
        render: asyncComponent(() => import('./SiteAdminExternalServicePage'), 'SiteAdminExternalServicePage'),
        exact: true,
    },
    {
        path: '/repositories',
        render: asyncComponent(() => import('./SiteAdminRepositoriesPage'), 'SiteAdminRepositoriesPage'),
        exact: true,
    },
    {
        path: '/organizations',
        render: asyncComponent(() => import('./SiteAdminOrgsPage'), 'SiteAdminOrgsPage'),
        exact: true,
    },
    {
        path: '/users',
        exact: true,
        render: asyncComponent(() => import('./SiteAdminAllUsersPage'), 'SiteAdminAllUsersPage'),
    },
    {
        path: '/users/new',
        render: asyncComponent(() => import('./SiteAdminCreateUserPage'), 'SiteAdminCreateUserPage'),
        exact: true,
    },
    {
        path: '/tokens',
        exact: true,
        render: asyncComponent(() => import('./SiteAdminTokensPage'), 'SiteAdminTokensPage'),
    },
    {
        path: '/usage-statistics',
        exact: true,
        render: asyncComponent(() => import('./SiteAdminUsageStatisticsPage'), 'SiteAdminUsageStatisticsPage'),
    },
    {
        path: '/updates',
        render: asyncComponent(() => import('./SiteAdminUpdatesPage'), 'SiteAdminUpdatesPage'),
        exact: true,
    },
    {
        path: '/pings',
        render: asyncComponent(() => import('./SiteAdminPingsPage'), 'SiteAdminPingsPage'),
        exact: true,
    },
    {
        path: '/surveys',
        exact: true,
        render: asyncComponent(() => import('./SiteAdminSurveyResponsesPage'), 'SiteAdminSurveyResponsesPage'),
    },
]
