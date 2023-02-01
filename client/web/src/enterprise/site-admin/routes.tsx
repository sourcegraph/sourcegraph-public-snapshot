import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { siteAdminAreaRoutes } from '../../site-admin/routes'
import { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
import { Navigate, useLocation } from 'react-router-dom-v5-compat'

const SiteAdminProductSubscriptionPage = lazyComponent(
    () => import('./productSubscription/SiteAdminProductSubscriptionPage'),
    'SiteAdminProductSubscriptionPage'
)
const SiteAdminProductCustomersPage = lazyComponent(
    () => import('./dotcom/customers/SiteAdminCustomersPage'),
    'SiteAdminProductCustomersPage'
)
const SiteAdminCreateProductSubscriptionPage = lazyComponent(
    () => import('./dotcom/productSubscriptions/SiteAdminCreateProductSubscriptionPage'),
    'SiteAdminCreateProductSubscriptionPage'
)
const DotComSiteAdminProductSubscriptionPage = lazyComponent(
    () => import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionPage'),
    'SiteAdminProductSubscriptionPage'
)
const SiteAdminProductSubscriptionsPage = lazyComponent(
    () => import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionsPage'),
    'SiteAdminProductSubscriptionsPage'
)
const SiteAdminProductLicensesPage = lazyComponent(
    () => import('./dotcom/productSubscriptions/SiteAdminProductLicensesPage'),
    'SiteAdminProductLicensesPage'
)
const SiteAdminAuthenticationProvidersPage = lazyComponent(
    () => import('./SiteAdminAuthenticationProvidersPage'),
    'SiteAdminAuthenticationProvidersPage'
)
const SiteAdminExternalAccountsPage = lazyComponent(
    () => import('./SiteAdminExternalAccountsPage'),
    'SiteAdminExternalAccountsPage'
)
const BatchChangesSiteConfigSettingsArea = lazyComponent(
    () => import('../batches/settings/BatchChangesSiteConfigSettingsArea'),
    'BatchChangesSiteConfigSettingsArea'
)
const BatchSpecsPage = lazyComponent(() => import('../batches/BatchSpecsPage'), 'BatchSpecsPage')
const WebhookLogPage = lazyComponent(() => import('../../site-admin/webhooks/WebhookLogPage'), 'WebhookLogPage')
const CodeIntelConfigurationPage = lazyComponent(
    () => import('../codeintel/configuration/pages/CodeIntelConfigurationPage'),
    'CodeIntelConfigurationPage'
)
const CodeIntelConfigurationPolicyPage = lazyComponent(
    () => import('../codeintel/configuration/pages/CodeIntelConfigurationPolicyPage'),
    'CodeIntelConfigurationPolicyPage'
)
const CodeIntelInferenceConfigurationPage = lazyComponent(
    () => import('../codeintel/configuration/pages/CodeIntelInferenceConfigurationPage'),
    'CodeIntelInferenceConfigurationPage'
)
const SiteAdminLsifUploadPage = lazyComponent(() => import('./SiteAdminLsifUploadPage'), 'SiteAdminLsifUploadPage')
const ExecutorsSiteAdminArea = lazyComponent(
    () => import('../executors/ExecutorsSiteAdminArea'),
    'ExecutorsSiteAdminArea'
)

export const enterpriseSiteAdminAreaRoutes: readonly SiteAdminAreaRoute[] = (
    [
        ...siteAdminAreaRoutes,
        {
            path: '/license',
            render: () => <SiteAdminProductSubscriptionPage />,
        },
        {
            path: '/dotcom/customers',
            render: () => <SiteAdminProductCustomersPage />,
            condition: () => SHOW_BUSINESS_FEATURES,
        },
        {
            path: '/dotcom/product/subscriptions/new',
            render: props => <SiteAdminCreateProductSubscriptionPage {...props} />,
            condition: () => SHOW_BUSINESS_FEATURES,
        },
        {
            path: '/dotcom/product/subscriptions/:subscriptionUUID',
            render: () => <DotComSiteAdminProductSubscriptionPage />,
            condition: () => SHOW_BUSINESS_FEATURES,
        },
        {
            path: '/dotcom/product/subscriptions',
            render: () => <SiteAdminProductSubscriptionsPage />,
            condition: () => SHOW_BUSINESS_FEATURES,
        },
        {
            path: '/dotcom/product/licenses',
            render: () => <SiteAdminProductLicensesPage />,
            condition: () => SHOW_BUSINESS_FEATURES,
        },
        {
            path: '/auth/providers',
            render: () => <SiteAdminAuthenticationProvidersPage />,
        },
        {
            path: '/auth/external-accounts',
            render: () => <SiteAdminExternalAccountsPage />,
        },
        {
            path: '/batch-changes',
            render: () => <BatchChangesSiteConfigSettingsArea />,
            condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        },
        {
            path: '/batch-changes/specs',
            render: props => <BatchSpecsPage {...props} />,
            condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
                batchChangesEnabled && batchChangesExecutionEnabled,
        },
        {
            path: '/batch-changes/webhook-logs',
            render: () => <WebhookLogPage />,
            condition: ({ batchChangesEnabled, batchChangesWebhookLogsEnabled }) =>
                batchChangesEnabled && batchChangesWebhookLogsEnabled,
        },

        // Code intelligence redirect
        {
            path: '/code-intelligence',
            render: () => <NavigateToCodeGraph />,
        },
        {
            path: '/code-intelligence/*',
            render: () => <NavigateToCodeGraph />,
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
            render: props => <CodeIntelConfigurationPage {...props} />,
        },
        {
            path: '/code-graph/configuration/:id',
            render: props => <CodeIntelConfigurationPolicyPage {...props} />,
        },
        {
            path: '/code-graph/inference-configuration',
            render: props => <CodeIntelInferenceConfigurationPage {...props} />,
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
            render: () => <SiteAdminLsifUploadPage />,
        },

        // Executor routes
        {
            path: '/executors',
            render: () => <ExecutorsSiteAdminArea />,
            condition: () => Boolean(window.context?.executorsEnabled),
        },
        {
            path: '/executors/*',
            render: () => <ExecutorsSiteAdminArea />,
            condition: () => Boolean(window.context?.executorsEnabled),
        },
    ] as readonly (SiteAdminAreaRoute | undefined)[]
).filter(Boolean) as readonly SiteAdminAreaRoute[]

function NavigateToCodeGraph() {
    const location = useLocation()
    return <Navigate to={location.pathname.replace('/code-intelligence', '/code-graph')} />
}
