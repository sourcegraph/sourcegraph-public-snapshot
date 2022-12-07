import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { SiteAdminAreaRoute } from './SiteAdminArea'

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
        render: lazyComponent(() => import('./SiteAdminExternalServicesArea'), 'SiteAdminExternalServicesArea'),
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
        render: lazyComponent(
            () => import('./SiteAdminAllUsersPage/FeatureFlaggedUsersPage'),
            'FeatureFlaggedUsersPage'
        ),
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
    {
        path: '/migrations',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminMigrationsPage'), 'SiteAdminMigrationsPage'),
    },
    {
        path: '/outbound-requests',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminOutboundRequestsPage'), 'SiteAdminOutboundRequestsPage'),
    },
    {
        path: '/feature-flags',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminFeatureFlagsPage'), 'SiteAdminFeatureFlagsPage'),
    },
    {
        path: '/feature-flags/configuration/:name',
        exact: true,
        render: lazyComponent(
            () => import('./SiteAdminFeatureFlagConfigurationPage'),
            'SiteAdminFeatureFlagConfigurationPage'
        ),
    },
    {
        path: '/webhooks',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminWebhooksPage'), 'SiteAdminWebhooksPage'),
    },
    {
        path: '/webhooks/create',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminWebhookCreatePage'), 'SiteAdminWebhookCreatePage'),
    },
    {
        path: '/webhooks/:id',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminWebhookPage'), 'SiteAdminWebhookPage'),
    },
]
