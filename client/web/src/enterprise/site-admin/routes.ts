import { siteAdminAreaRoutes } from '../../site-admin/routes'
import { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
import { lazyComponent } from '../../util/lazyComponent'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
import type { ExecutorsListPageProps } from '../executors/ExecutorsListPage'

export const enterpriseSiteAdminAreaRoutes: readonly SiteAdminAreaRoute[] = [
    ...siteAdminAreaRoutes,
    {
        path: '/license',
        render: lazyComponent(
            () => import('./productSubscription/SiteAdminProductSubscriptionPage'),
            'SiteAdminProductSubscriptionPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/customers',
        render: lazyComponent(
            () => import('./dotcom/customers/SiteAdminCustomersPage'),
            'SiteAdminProductCustomersPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions/new',
        render: lazyComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminCreateProductSubscriptionPage'),
            'SiteAdminCreateProductSubscriptionPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions/:subscriptionUUID',
        render: lazyComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionPage'),
            'SiteAdminProductSubscriptionPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions',
        render: lazyComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionsPage'),
            'SiteAdminProductSubscriptionsPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
        exact: true,
    },
    {
        path: '/dotcom/product/licenses',
        render: lazyComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminProductLicensesPage'),
            'SiteAdminProductLicensesPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
        exact: true,
    },
    {
        path: '/auth/providers',
        render: lazyComponent(
            () => import('./SiteAdminAuthenticationProvidersPage'),
            'SiteAdminAuthenticationProvidersPage'
        ),
        exact: true,
    },
    {
        path: '/auth/external-accounts',
        render: lazyComponent(() => import('./SiteAdminExternalAccountsPage'), 'SiteAdminExternalAccountsPage'),
        exact: true,
    },
    {
        path: '/registry/extensions',
        render: lazyComponent(() => import('./SiteAdminRegistryExtensionsPage'), 'SiteAdminRegistryExtensionsPage'),
        exact: true,
    },

    {
        path: '/batch-changes',
        exact: true,
        render: lazyComponent(
            () => import('../batches/settings/BatchChangesSiteConfigSettingsArea'),
            'BatchChangesSiteConfigSettingsArea'
        ),
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
    },
    {
        path: '/batch-changes/specs',
        exact: true,
        render: lazyComponent(() => import('../batches/settings/BatchSpecsPage'), 'BatchSpecsPage'),
        condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
            batchChangesEnabled && batchChangesExecutionEnabled,
    },
    {
        path: '/batch-changes/webhook-logs',
        exact: true,
        render: lazyComponent(() => import('../../site-admin/webhooks/WebhookLogPage'), 'WebhookLogPage'),
        condition: ({ batchChangesEnabled, batchChangesWebhookLogsEnabled }) =>
            batchChangesEnabled && batchChangesWebhookLogsEnabled,
    },

    // Code intelligence upload routes
    {
        path: '/code-intelligence/uploads',
        render: lazyComponent(() => import('../codeintel/uploads/pages/CodeIntelUploadsPage'), 'CodeIntelUploadsPage'),
        exact: true,
    },
    {
        path: '/code-intelligence/uploads/:id',
        render: lazyComponent(() => import('../codeintel/uploads/pages/CodeIntelUploadPage'), 'CodeIntelUploadPage'),
        exact: true,
    },

    // Auto-indexing routes
    {
        path: '/code-intelligence/indexes',
        render: lazyComponent(() => import('../codeintel/indexes/pages/CodeIntelIndexesPage'), 'CodeIntelIndexesPage'),
        exact: true,
        condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
    },
    {
        path: '/code-intelligence/indexes/:id',
        render: lazyComponent(() => import('../codeintel/indexes/pages/CodeIntelIndexPage'), 'CodeIntelIndexPage'),
        exact: true,
        condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
    },

    // Code intelligence configuration
    {
        path: '/code-intelligence/configuration',
        render: lazyComponent(
            () => import('../codeintel/configuration/pages/CodeIntelConfigurationPage'),
            'CodeIntelConfigurationPage'
        ),
        exact: true,
    },
    {
        path: '/code-intelligence/configuration/:id',
        render: lazyComponent(
            () => import('../codeintel/configuration/pages/CodeIntelConfigurationPolicyPage'),
            'CodeIntelConfigurationPolicyPage'
        ),
        exact: true,
    },

    // Legacy routes
    {
        path: '/lsif-uploads/:id',
        render: lazyComponent(() => import('./SiteAdminLsifUploadPage'), 'SiteAdminLsifUploadPage'),
        exact: true,
    },

    // Executor routes
    {
        path: '/executors',
        render: lazyComponent<ExecutorsListPageProps, 'ExecutorsListPage'>(
            () => import('../executors/ExecutorsListPage'),
            'ExecutorsListPage'
        ),
        exact: true,
        condition: () => Boolean(window.context?.executorsEnabled),
    },
    // Organization routes
    {
        path: '/organizations/early-access-orgs-code',
        render: lazyComponent(() => import('../organizations/EarlyAccessOrgsCodeForm'), 'EarlyAccessOrgsCodeForm'),
        exact: true,
    },
]
