import { Redirect } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { siteAdminAreaRoutes } from '../../site-admin/routes'
import { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
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
        render: lazyComponent(() => import('../batches/BatchSpecsPage'), 'BatchSpecsPage'),
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

    // Code graph upload routes
    {
        path: '/code-intelligence',
        exact: false,
        render: props => <Redirect to={props.location.pathname.replace('/code-intelligence/', '/code-graph/')} />,
    },
    {
        path: '/code-graph/uploads',
        render: lazyComponent(() => import('../codeintel/uploads/pages/CodeIntelUploadsPage'), 'CodeIntelUploadsPage'),
        exact: true,
    },

    {
        path: '/code-graph/uploads/:id',
        render: lazyComponent(() => import('../codeintel/uploads/pages/CodeIntelUploadPage'), 'CodeIntelUploadPage'),
        exact: true,
    },

    // Auto-indexing routes
    {
        path: '/code-graph/indexes',
        render: lazyComponent(() => import('../codeintel/indexes/pages/CodeIntelIndexesPage'), 'CodeIntelIndexesPage'),
        exact: true,
        condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
    },
    {
        path: '/code-graph/indexes/:id',
        render: lazyComponent(() => import('../codeintel/indexes/pages/CodeIntelIndexPage'), 'CodeIntelIndexPage'),
        exact: true,
        condition: () => Boolean(window.context?.codeIntelAutoIndexingEnabled),
    },

    // Lockfile indexes & dependency search routes
    {
        path: '/code-graph/lockfiles',
        render: lazyComponent(
            () => import('../codeintel/lockfiles/pages/CodeIntelLockfilesPage'),
            'CodeIntelLockfilesPage'
        ),
        exact: true,
        condition: () => Boolean(window.context?.codeIntelLockfileIndexingEnabled),
    },
    {
        path: '/code-graph/lockfiles/:id',
        render: lazyComponent(
            () => import('../codeintel/lockfiles/pages/CodeIntelLockfilePage'),
            'CodeIntelLockfilePage'
        ),
        exact: true,
        condition: () => Boolean(window.context?.codeIntelLockfileIndexingEnabled),
    },

    // Code graph configuration
    {
        path: '/code-graph/configuration',
        render: lazyComponent(
            () => import('../codeintel/configuration/pages/CodeIntelConfigurationPage'),
            'CodeIntelConfigurationPage'
        ),
        exact: true,
    },
    {
        path: '/code-graph/configuration/:id',
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
