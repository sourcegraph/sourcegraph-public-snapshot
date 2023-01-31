import { Redirect } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { siteAdminAreaRoutes } from '../../site-admin/routes'
import { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
import type { ExecutorsSiteAdminAreaProps } from '../executors/ExecutorsSiteAdminArea'

export const enterpriseSiteAdminAreaRoutes: readonly SiteAdminAreaRoute[] = (
    [
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

        // Code intelligence redirect
        {
            path: '/code-intelligence',
            exact: false,
            render: props => <Redirect to={props.location.pathname.replace('/code-intelligence/', '/code-graph/')} />,
        },

        // Precise index routes
        {
            path: '/code-graph/indexes',
            render: lazyComponent(
                () => import('../codeintel/indexes/pages/CodeIntelPreciseIndexesPage'),
                'CodeIntelPreciseIndexesPage'
            ),
            exact: true,
        },
        {
            path: '/code-graph/indexes/:id',
            render: lazyComponent(
                () => import('../codeintel/indexes/pages/CodeIntelPreciseIndexPage'),
                'CodeIntelPreciseIndexPage'
            ),
            exact: true,
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
        {
            path: '/code-graph/inference-configuration',
            render: lazyComponent(
                () => import('../codeintel/configuration/pages/CodeIntelInferenceConfigurationPage'),
                'CodeIntelInferenceConfigurationPage'
            ),
            exact: true,
        },

        // Legacy routes
        {
            path: '/code-graph/uploads/:id',
            render: props => (
                <Redirect
                    to={`../indexes/${btoa(
                        `PreciseIndex:"U:${(atob(props.match.params.id).match(/(\d+)/) ?? [''])[0]}"`
                    )}`}
                />
            ),
            exact: true,
        },
        {
            path: '/lsif-uploads/:id',
            render: lazyComponent(() => import('./SiteAdminLsifUploadPage'), 'SiteAdminLsifUploadPage'),
            exact: true,
        },

        // Executor routes
        {
            path: '/executors',
            render: lazyComponent<ExecutorsSiteAdminAreaProps, 'ExecutorsSiteAdminArea'>(
                () => import('../executors/ExecutorsSiteAdminArea'),
                'ExecutorsSiteAdminArea'
            ),
            condition: () => Boolean(window.context?.executorsEnabled),
        },
    ] as readonly (SiteAdminAreaRoute | undefined)[]
).filter(Boolean) as readonly SiteAdminAreaRoute[]
