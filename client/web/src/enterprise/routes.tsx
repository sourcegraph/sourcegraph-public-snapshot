import { Redirect } from 'react-router'

import { isErrorLike } from '@sourcegraph/common'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { isCodeInsightsEnabled } from '../insights/utils/is-code-insights-enabled'
import { LayoutRouteProps, routes } from '../routes'
import { EnterprisePageRoutes } from '../routes.constants'

const isSearchContextsManagementEnabled = (settingsCascade: SettingsCascadeOrError): boolean =>
    !isErrorLike(settingsCascade.final) &&
    settingsCascade.final?.experimentalFeatures?.showSearchContext !== false &&
    settingsCascade.final?.experimentalFeatures?.showSearchContextManagement !== false

export const enterpriseRoutes: readonly LayoutRouteProps<any>[] = [
    {
        // Allow unauthenticated viewers to view the "new subscription" page to price out a subscription (instead
        // of just dumping them on a sign-in page).
        path: EnterprisePageRoutes.SubscriptionsNew,
        exact: true,
        render: lazyComponent(
            () => import('./user/productSubscriptions/NewProductSubscriptionPageOrRedirectUser'),
            'NewProductSubscriptionPageOrRedirectUser'
        ),
    },
    {
        // Redirect from old /user/subscriptions/new -> /subscriptions/new.
        path: EnterprisePageRoutes.OldSubscriptionsNew,
        exact: true,
        render: () => <Redirect to="/subscriptions/new" />,
    },
    {
        path: EnterprisePageRoutes.BatchChanges,
        render: lazyComponent(() => import('./batches/global/GlobalBatchChangesArea'), 'GlobalBatchChangesArea'),
        // We also render this route on sourcegraph.com as a precaution in case anyone
        // follows an in-app link to /batch-changes from sourcegraph.com; the component
        // will just redirect the visitor to the marketing page
        condition: ({ batchChangesEnabled, isSourcegraphDotCom }) => batchChangesEnabled || isSourcegraphDotCom,
    },
    {
        path: EnterprisePageRoutes.Stats,
        render: lazyComponent(() => import('./search/stats/SearchStatsPage'), 'SearchStatsPage'),
    },
    {
        path: EnterprisePageRoutes.CodeMonitoring,
        render: lazyComponent(
            () => import('./code-monitoring/global/GlobalCodeMonitoringArea'),
            'GlobalCodeMonitoringArea'
        ),
    },
    {
        path: EnterprisePageRoutes.Insights,
        render: lazyComponent(() => import('./insights/CodeInsightsRouter'), 'CodeInsightsRouter'),
        condition: props => isCodeInsightsEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.Contexts,
        render: lazyComponent(() => import('./searchContexts/SearchContextsListPage'), 'SearchContextsListPage'),
        exact: true,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.CreateContext,
        render: lazyComponent(() => import('./searchContexts/CreateSearchContextPage'), 'CreateSearchContextPage'),
        exact: true,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.EditContext,
        render: lazyComponent(() => import('./searchContexts/EditSearchContextPage'), 'EditSearchContextPage'),
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.Context,
        render: lazyComponent(() => import('./searchContexts/SearchContextPage'), 'SearchContextPage'),
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    ...routes,
]
