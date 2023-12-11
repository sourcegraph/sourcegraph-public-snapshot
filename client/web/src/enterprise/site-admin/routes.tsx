import { Navigate, useLocation } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { FeedbackBadge } from '@sourcegraph/wildcard'

import { otherSiteAdminRoutes, UsersManagement } from '../../site-admin/routes'
import type { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
import type { BatchSpecsPageProps } from '../batches/BatchSpecsPage'
import { CodeIntelConfigurationPolicyPage } from '../codeintel/configuration/pages/CodeIntelConfigurationPolicyPage'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
import { OwnAnalyticsPage } from '../own/admin-ui/OwnAnalyticsPage'
import type { SiteAdminRolesPageProps } from '../rbac/SiteAdminRolesPage'

import type { RoleAssignmentModalProps } from './UserManagement/components/RoleAssignmentModal'

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
const SiteAdminLicenseKeyLookupPage = lazyComponent(
    () => import('./dotcom/productSubscriptions/SiteAdminLicenseKeyLookupPage'),
    'SiteAdminLicenseKeyLookupPage'
)
const SiteAdminAuthenticationProvidersPage = lazyComponent(
    () => import('./SiteAdminAuthenticationProvidersPage'),
    'SiteAdminAuthenticationProvidersPage'
)
const SiteAdminExternalAccountsPage = lazyComponent(
    () => import('./SiteAdminExternalAccountsPage'),
    'SiteAdminExternalAccountsPage'
)
const BatchChangesSiteConfigSettingsPage = lazyComponent(
    () => import('../batches/settings/BatchChangesSiteConfigSettingsPage'),
    'BatchChangesSiteConfigSettingsPage'
)
const BatchChangesCreateGitHubAppPage = lazyComponent(
    () => import('../batches/settings/BatchChangesCreateGitHubAppPage'),
    'BatchChangesCreateGitHubAppPage'
)
const GitHubAppPage = lazyComponent(() => import('../../components/gitHubApps/GitHubAppPage'), 'GitHubAppPage')
const BatchSpecsPage = lazyComponent<BatchSpecsPageProps, 'BatchSpecsPage'>(
    () => import('../batches/BatchSpecsPage'),
    'BatchSpecsPage'
)
const AdminCodeIntelArea = lazyComponent(() => import('../codeintel/admin/AdminCodeIntelArea'), 'AdminCodeIntelArea')
const SiteAdminPreciseIndexPage = lazyComponent(
    () => import('./SiteAdminPreciseIndexPage'),
    'SiteAdminPreciseIndexPage'
)
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
const OwnStatusPage = lazyComponent(() => import('../own/admin-ui/OwnStatusPage'), 'OwnStatusPage')

const SiteAdminCodyPage = lazyComponent(() => import('./cody/SiteAdminCodyPage'), 'SiteAdminCodyPage')
const CodyConfigurationPage = lazyComponent(
    () => import('../cody/configuration/pages/CodyConfigurationPage'),
    'CodyConfigurationPage'
)

const codyIsEnabled = (): boolean => Boolean(window.context?.codyEnabled && window.context?.embeddingsEnabled)

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
            render: () => <SiteAdminLicenseKeyLookupPage />,
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
            render: () => <BatchChangesSiteConfigSettingsPage />,
            condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        },
        {
            path: '/batch-changes/github-apps/new',
            render: () => <BatchChangesCreateGitHubAppPage />,
            condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        },
        {
            path: '/batch-changes/github-apps/:appID',
            render: props => (
                <GitHubAppPage
                    headerParentBreadcrumb={{ to: '/site-admin/batch-changes', text: 'Batch Changes settings' }}
                    headerAnnotation={<FeedbackBadge status="beta" feedback={{ mailto: 'support@sourcegraph.com' }} />}
                    telemetryService={props.telemetryService}
                    telemetryRecorder={props.telemetryRecorder}
                />
            ),
            condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        },
        {
            path: '/batch-changes/specs',
            render: () => <BatchSpecsPage />,
            condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
                batchChangesEnabled && batchChangesExecutionEnabled,
        },
        // Old batch changes webhooks logs page redirects to new incoming webhooks page.
        // The old page components and documentation are still available in the codebase
        // but should be fully removed in the next release.
        {
            path: '/batch-changes/webhook-logs',
            render: () => <Navigate to="/site-admin/webhooks/incoming" replace={true} />,
            condition: ({ batchChangesEnabled, batchChangesWebhookLogsEnabled }) =>
                batchChangesEnabled && batchChangesWebhookLogsEnabled,
        },

        // Enterprise maintenance area

        {
            exact: true,
            path: '/code-insights-jobs',
            render: () => <CodeInsightsJobsPage />,
            condition: ({ codeInsightsEnabled }) => codeInsightsEnabled,
        },
        {
            exact: true,
            path: '/own-signal-page',
            render: () => <OwnStatusPage />,
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
            render: () => <SiteAdminPreciseIndexPage />,
        },

        // Executor routes
        {
            path: '/executors/*',
            render: () => <ExecutorsSiteAdminArea />,
            condition: () => Boolean(window.context?.executorsEnabled),
        },

        // Cody configuration
        {
            exact: true,
            path: '/cody',
            render: () => <Navigate to="/site-admin/embeddings" />,
            condition: codyIsEnabled,
        },
        {
            exact: true,
            path: '/embeddings',
            render: props => <SiteAdminCodyPage {...props} />,
            condition: codyIsEnabled,
        },
        {
            exact: true,
            path: '/embeddings/configuration',
            render: props => <CodyConfigurationPage {...props} />,
            condition: codyIsEnabled,
        },
        {
            path: '/embeddings/configuration/:id',
            render: props => <CodeIntelConfigurationPolicyPage {...props} domain="embeddings" />,
            condition: codyIsEnabled,
        },

        // rbac-related routes
        {
            path: '/roles',
            exact: true,
            render: props => <SiteAdminRolesPage {...props} />,
        },

        // Own analytics
        {
            exact: true,
            path: '/analytics/own',
            render: () => <OwnAnalyticsPage />,
        },
    ] as readonly (SiteAdminAreaRoute | undefined)[]
).filter(Boolean) as readonly SiteAdminAreaRoute[]

function NavigateToCodeGraph(): JSX.Element {
    const location = useLocation()
    return <Navigate to={location.pathname.replace('/code-intelligence', '/code-graph')} />
}
