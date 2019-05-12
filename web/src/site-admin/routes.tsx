import React from 'react'
import { hot } from 'react-hot-loader'
import { Redirect } from 'react-router'
import { eventLogger } from '../tracking/eventLogger'
import { asyncComponent, asyncComponent2 } from '../util/asyncComponent'
import { SiteAdminAreaRoute } from './SiteAdminArea'

const SiteAdminAddExternalServicesPage = asyncComponent(
    () => import('./SiteAdminAddExternalServicesPage'),
    'SiteAdminAddExternalServicesPage'
)

export const siteAdminAreaRoutes: ReadonlyArray<SiteAdminAreaRoute> = [
    {
        // Render empty page if no page selected
        path: '',
        render: asyncComponent(async () => {
            const x = await import('./SiteAdminOverviewPage')
            return Promise.resolve({ SiteAdminOverviewPage: hot(module)(x.SiteAdminOverviewPage) })
        }, 'SiteAdminOverviewPage'),
        exact: true,
    },
    {
        path: '/configuration',
        exact: true,
        render: asyncComponent(
            () => import('./SiteAdminConfigurationPage'),
            'SiteAdminConfigurationPage',
            require.resolveWeak('./SiteAdminConfigurationPage')
        ),
    },
    {
        path: '/global-settings',
        exact: true,
        render: asyncComponent(
            () => import('./SiteAdminSettingsPage'),
            'SiteAdminSettingsPage',
            require.resolveWeak('./SiteAdminSettingsPage')
        ),
    },
    {
        path: '/external-services',
        render: asyncComponent(
            () => import('./SiteAdminExternalServicesPage'),
            'SiteAdminExternalServicesPage',
            require.resolveWeak('./SiteAdminExternalServicesPage')
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
        render: props => <SiteAdminAddExternalServicesPage {...props} eventLogger={eventLogger} />,
        exact: true,
    },
    {
        path: '/external-services/:id',
        render: asyncComponent(
            () => import('./SiteAdminExternalServicePage'),
            'SiteAdminExternalServicePage',
            require.resolveWeak('./SiteAdminExternalServicePage')
        ),
        exact: true,
    },
    {
        path: '/repositories',
        render: asyncComponent(
            () => import('./SiteAdminRepositoriesPage'),
            'SiteAdminRepositoriesPage',
            require.resolveWeak('./SiteAdminRepositoriesPage')
        ),
        exact: true,
    },
    {
        path: '/organizations',
        render: asyncComponent(
            () => import('./SiteAdminOrgsPage'),
            'SiteAdminOrgsPage',
            require.resolveWeak('./SiteAdminOrgsPage')
        ),
        exact: true,
    },
    {
        path: '/users',
        exact: true,
        render: asyncComponent(
            () => import('./SiteAdminAllUsersPage'),
            'SiteAdminAllUsersPage',
            require.resolveWeak('./SiteAdminAllUsersPage')
        ),
    },
    {
        path: '/users/new',
        render: asyncComponent(
            () => import('./SiteAdminCreateUserPage'),
            'SiteAdminCreateUserPage',
            require.resolveWeak('./SiteAdminCreateUserPage')
        ),
        exact: true,
    },
    {
        path: '/tokens',
        exact: true,
        render: asyncComponent(
            () => import('./SiteAdminTokensPage'),
            'SiteAdminTokensPage',
            require.resolveWeak('./SiteAdminTokensPage')
        ),
    },
    {
        path: '/usage-statistics',
        exact: true,
        render: asyncComponent(
            () => import('./SiteAdminUsageStatisticsPage'),
            'SiteAdminUsageStatisticsPage',
            require.resolveWeak('./SiteAdminUsageStatisticsPage')
        ),
    },
    {
        path: '/updates',
        render: asyncComponent(
            () => import('./SiteAdminUpdatesPage'),
            'SiteAdminUpdatesPage',
            require.resolveWeak('./SiteAdminUpdatesPage')
        ),
        exact: true,
    },
    {
        path: '/pings',
        render: asyncComponent(
            () => import('./SiteAdminPingsPage'),
            'SiteAdminPingsPage',
            require.resolveWeak('./SiteAdminPingsPage')
        ),
        exact: true,
    },
    {
        path: '/surveys',
        exact: true,
        render: asyncComponent(
            () => import('./SiteAdminSurveyResponsesPage'),
            'SiteAdminSurveyResponsesPage',
            require.resolveWeak('./SiteAdminSurveyResponsesPage')
        ),
    },
]
