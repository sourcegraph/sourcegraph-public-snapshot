import React from 'react'
import { Redirect } from 'react-router'
import { eventLogger } from '../tracking/eventLogger'
import { SiteAdminAreaRoute } from './SiteAdminArea'
const SiteAdminAddExternalServicesPage = React.lazy(async () => ({
    default: (await import('./SiteAdminAddExternalServicesPage')).SiteAdminAddExternalServicesPage,
}))
const SiteAdminAllUsersPage = React.lazy(async () => ({
    default: (await import('./SiteAdminAllUsersPage')).SiteAdminAllUsersPage,
}))
const SiteAdminConfigurationPage = React.lazy(async () => ({
    default: (await import('./SiteAdminConfigurationPage')).SiteAdminConfigurationPage,
}))
const SiteAdminCreateUserPage = React.lazy(async () => ({
    default: (await import('./SiteAdminCreateUserPage')).SiteAdminCreateUserPage,
}))
const SiteAdminExternalServicePage = React.lazy(async () => ({
    default: (await import('./SiteAdminExternalServicePage')).SiteAdminExternalServicePage,
}))
const SiteAdminExternalServicesPage = React.lazy(async () => ({
    default: (await import('./SiteAdminExternalServicesPage')).SiteAdminExternalServicesPage,
}))
const SiteAdminOrgsPage = React.lazy(async () => ({ default: (await import('./SiteAdminOrgsPage')).SiteAdminOrgsPage }))
const SiteAdminOverviewPage = React.lazy(async () => ({
    default: (await import('./SiteAdminOverviewPage')).SiteAdminOverviewPage,
}))
const SiteAdminPingsPage = React.lazy(async () => ({
    default: (await import('./SiteAdminPingsPage')).SiteAdminPingsPage,
}))
const SiteAdminRepositoriesPage = React.lazy(async () => ({
    default: (await import('./SiteAdminRepositoriesPage')).SiteAdminRepositoriesPage,
}))
const SiteAdminSettingsPage = React.lazy(async () => ({
    default: (await import('./SiteAdminSettingsPage')).SiteAdminSettingsPage,
}))
const SiteAdminSurveyResponsesPage = React.lazy(async () => ({
    default: (await import('./SiteAdminSurveyResponsesPage')).SiteAdminSurveyResponsesPage,
}))
const SiteAdminTokensPage = React.lazy(async () => ({
    default: (await import('./SiteAdminTokensPage')).SiteAdminTokensPage,
}))
const SiteAdminUpdatesPage = React.lazy(async () => ({
    default: (await import('./SiteAdminUpdatesPage')).SiteAdminUpdatesPage,
}))
const SiteAdminUsageStatisticsPage = React.lazy(async () => ({
    default: (await import('./SiteAdminUsageStatisticsPage')).SiteAdminUsageStatisticsPage,
}))

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
