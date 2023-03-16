import { Navigate, useLocation } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { otherSiteAdminRoutes, UsersManagement } from '../../site-admin/routes'
import { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
import { BatchSpecsPageProps } from '../batches/BatchSpecsPage'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
import { SiteAdminRolesPageProps } from '../rbac/SiteAdminRolesPage'

import { RoleAssignmentModalProps } from './UserManagement/components/RoleAssignmentModal'

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
const BatchSpecsPage = lazyComponent<BatchSpecsPageProps, 'BatchSpecsPage'>(
    () => import('../batches/BatchSpecsPage'),
    'BatchSpecsPage'
)
const WebhookLogPage = lazyComponent(() => import('../../site-admin/webhooks/WebhookLogPage'), 'WebhookLogPage')
const AdminCodeIntelArea = lazyComponent(() => import('../codeintel/admin/AdminCodeIntelArea'), 'AdminCodeIntelArea')
const SiteAdminLsifUploadPage = lazyComponent(() => import('./SiteAdminLsifUploadPage'), 'SiteAdminLsifUploadPage')
const ExecutorsSiteAdminArea = lazyComponent(
    () => import('../executors/ExecutorsSiteAdminArea'),
    'ExecutorsSiteAdminArea'
)

const SiteAdminRolesPage = lazyComponent<SiteAdminRolesPageProps, 'SiteAdminRolesPage'>(
    () => import('../rbac/SiteAdminRolesPage'),
    'SiteAdminRolesPage'
)

const RoleAssignmentModal = lazyComponent<RoleAssignmentModalProps, 'RoleAssignmentModal'>(
    () => import('./UserManagement/components/RoleAssignmentModal'),
    'RoleAssignmentModal'
)

const CodeInsightsJobsPage = lazyComponent(() => import('../insights/admin-ui/CodeInsightsJobs'), 'CodeInsightsJobs')

const SiteAdminCodyPage = lazyComponent(() => import('./cody/SiteAdminCodyPage'), 'SiteAdminCodyPage')

export const enterpriseSiteAdminAreaRoutes: readonly SiteAdminAreaRoute[] = (
    [
        ...otherSiteAdminRoutes,
        {
            path: '/users',
            render: () => (
                <UsersManagement
                    isEnterprise={true}
                    renderAssignmentModal={(onCancel, onSuccess, user) => (
                        <RoleAssignmentModal onCancel={onCancel} onSuccess={onSuccess} user={user} />
                    )}
                />
            ),
        },
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
            render: () => <BatchSpecsPage />,
            condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
                batchChangesEnabled && batchChangesExecutionEnabled,
        },
        {
            path: '/batch-changes/webhook-logs',
            render: () => <WebhookLogPage />,
            condition: ({ batchChangesEnabled, batchChangesWebhookLogsEnabled }) =>
                batchChangesEnabled && batchChangesWebhookLogsEnabled,
        },

        // Enterprise maintenance area

        {
            exact: true,
            path: '/code-insights-jobs',
            render: () => <CodeInsightsJobsPage />,
        },

        // Code intelligence redirect
        {
            path: '/code-intelligence/*',
            render: () => <NavigateToCodeGraph />,
        },
        // Code graph routes
        {
            path: '/code-graph/*',
            render: props => <AdminCodeIntelArea {...props} />,
        },
        {
            path: '/lsif-uploads/:id',
            render: () => <SiteAdminLsifUploadPage />,
        },

        // Executor routes
        {
            path: '/executors/*',
            render: () => <ExecutorsSiteAdminArea />,
            condition: () => Boolean(window.context?.executorsEnabled),
        },

        // Cody configuration
        {
            path: '/cody',
            render: props => <SiteAdminCodyPage {...props} />,
            condition: () => Boolean(window.context?.embeddingsEnabled),
        },
        // rbac-related routes
        {
            path: '/roles',
            exact: true,
            render: props => <SiteAdminRolesPage {...props} />,
        },
    ] as readonly (SiteAdminAreaRoute | undefined)[]
).filter(Boolean) as readonly SiteAdminAreaRoute[]

function NavigateToCodeGraph(): JSX.Element {
    const location = useLocation()
    return <Navigate to={location.pathname.replace('/code-intelligence', '/code-graph')} />
}
