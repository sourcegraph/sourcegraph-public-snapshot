import React from 'react'
import { Redirect } from 'react-router'
import { eventLogger } from '../tracking/eventLogger'
import { asyncComponent } from '../util/asyncComponent'
import { SiteAdminAreaRoute } from './SiteAdminArea'

const SiteAdminAddExternalServicesPage = asyncComponent(
    () => import('./SiteAdminAddExternalServicesPage'),
    'SiteAdminAddExternalServicesPage'
)
const SiteAdminAllUsersPage = asyncComponent(() => import('./SiteAdminAllUsersPage'), 'SiteAdminAllUsersPage')
const SiteAdminConfigurationPage = asyncComponent(
    () => import('./SiteAdminConfigurationPage'),
    'SiteAdminConfigurationPage'
)
const SiteAdminCreateUserPage = asyncComponent(() => import('./SiteAdminCreateUserPage'), 'SiteAdminCreateUserPage')
const SiteAdminExternalServicePage = asyncComponent(
    () => import('./SiteAdminExternalServicePage'),
    'SiteAdminExternalServicePage'
)
const SiteAdminExternalServicesPage = asyncComponent(
    () => import('./SiteAdminExternalServicesPage'),
    'SiteAdminExternalServicesPage'
)
const SiteAdminOrgsPage = asyncComponent(() => import('./SiteAdminOrgsPage'), 'SiteAdminOrgsPage')
const SiteAdminOverviewPage = asyncComponent(() => import('./SiteAdminOverviewPage'), 'SiteAdminOverviewPage')
const SiteAdminPingsPage = asyncComponent(() => import('./SiteAdminPingsPage'), 'SiteAdminPingsPage')
const SiteAdminRepositoriesPage = asyncComponent(
    () => import('./SiteAdminRepositoriesPage'),
    'SiteAdminRepositoriesPage'
)
const SiteAdminSettingsPage = asyncComponent(() => import('./SiteAdminSettingsPage'), 'SiteAdminSettingsPage')
const SiteAdminSurveyResponsesPage = asyncComponent(
    () => import('./SiteAdminSurveyResponsesPage'),
    'SiteAdminSurveyResponsesPage'
)
const SiteAdminTokensPage = asyncComponent(() => import('./SiteAdminTokensPage'), 'SiteAdminTokensPage')
const SiteAdminUpdatesPage = asyncComponent(() => import('./SiteAdminUpdatesPage'), 'SiteAdminUpdatesPage')
const SiteAdminUsageStatisticsPage = asyncComponent(
    () => import('./SiteAdminUsageStatisticsPage'),
    'SiteAdminUsageStatisticsPage'
)

export const siteAdminAreaRoutes: ReadonlyArray<SiteAdminAreaRoute> = [
    {
        // Render empty page if no page selected
        path: '',
        render: props => <SiteAdminOverviewPage {...props} />,
        exact: true,
    },
    {
        path: '/configuration',
        exact: true,
        render: props => <SiteAdminConfigurationPage {...props} />,
    },
    {
        path: '/global-settings',
        exact: true,
        render: props => <SiteAdminSettingsPage {...props} />,
    },
    {
        path: '/external-services',
        render: props => <SiteAdminExternalServicesPage {...props} />,
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
        render: props => <SiteAdminExternalServicePage {...props} />,
        exact: true,
    },
    {
        path: '/repositories',
        render: props => <SiteAdminRepositoriesPage {...props} />,
        exact: true,
    },
    {
        path: '/organizations',
        render: props => <SiteAdminOrgsPage {...props} />,
        exact: true,
    },
    {
        path: '/users',
        exact: true,
        render: props => <SiteAdminAllUsersPage {...props} />,
    },
    {
        path: '/users/new',
        render: props => <SiteAdminCreateUserPage {...props} />,
        exact: true,
    },
    {
        path: '/tokens',
        exact: true,
        render: props => <SiteAdminTokensPage {...props} />,
    },
    {
        path: '/usage-statistics',
        exact: true,
        render: props => <SiteAdminUsageStatisticsPage {...props} />,
    },
    {
        path: '/updates',
        render: props => <SiteAdminUpdatesPage {...props} />,
        exact: true,
    },
    {
        path: '/pings',
        render: props => <SiteAdminPingsPage {...props} />,
        exact: true,
    },
    {
        path: '/surveys',
        exact: true,
        render: props => <SiteAdminSurveyResponsesPage {...props} />,
    },
]
