import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { SiteAdminAreaRoute } from './SiteAdminArea'

export const siteAdminAreaRoutes: readonly SiteAdminAreaRoute[] = [
    {
        path: '/',
        render: lazyComponent(() => import('./analytics/AnalyticsOverviewPage'), 'AnalyticsOverviewPage'),
        exact: true,
    },
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
        render: lazyComponent(() => import('./analytics/AnalyticsCodeInsightsPage'), 'AnalyticsCodeInsightsPage'),
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
        render: lazyComponent(() => import('./UserManagement'), 'UsersManagement'),
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
        path: '/background-jobs',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminBackgroundJobsPage'), 'SiteAdminBackgroundJobsPage'),
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
    {
        path: '/slow-requests',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminSlowRequestsPage'), 'SiteAdminSlowRequestsPage'),
    },
    {
        path: '/webhooks/:id/edit',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminWebhookUpdatePage'), 'SiteAdminWebhookUpdatePage'),
    },
]
