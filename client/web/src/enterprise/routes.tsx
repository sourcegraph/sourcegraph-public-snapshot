import React from 'react'
import { Redirect } from 'react-router'

import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { isCodeInsightsEnabled } from '../insights/utils/is-code-insights-enabled'
import { LayoutRouteProps, routes } from '../routes'
import { lazyComponent } from '../util/lazyComponent'

export enum EnterprisePageRoutes {
    SUBSCRIPTIONS_NEW = '/subscriptions/new',
    OLD_SUBSCRIPTIONS_NEW = '/user/subscriptions/new',
    BATCH_CHANGES = '/batch-changes',
    STATS = '/stats',
    CODE_MONITORING = '/code-monitoring',
    INSIGHTS = '/insights',
    CONTEXTS = '/contexts',
    CREATE_CONTEXT = '/contexts/new',
    EDIT_CONTEXT = '/contexts/:spec+/edit',
    CONTEXT = '/contexts/:spec+',
}

const isSearchContextsManagementEnabled = (settingsCascade: SettingsCascadeOrError): boolean =>
    !isErrorLike(settingsCascade.final) &&
    settingsCascade.final?.experimentalFeatures?.showSearchContext !== false &&
    settingsCascade.final?.experimentalFeatures?.showSearchContextManagement !== false

export const enterpriseRoutes: readonly LayoutRouteProps<any>[] = [
    {
        // Allow unauthenticated viewers to view the "new subscription" page to price out a subscription (instead
        // of just dumping them on a sign-in page).
        path: EnterprisePageRoutes.SUBSCRIPTIONS_NEW,
        exact: true,
        render: lazyComponent(
            () => import('./user/productSubscriptions/NewProductSubscriptionPageOrRedirectUser'),
            'NewProductSubscriptionPageOrRedirectUser'
        ),
    },
    {
        // Redirect from old /user/subscriptions/new -> /subscriptions/new.
        path: EnterprisePageRoutes.OLD_SUBSCRIPTIONS_NEW,
        exact: true,
        render: () => <Redirect to="/subscriptions/new" />,
    },
    {
        path: EnterprisePageRoutes.BATCH_CHANGES,
        render: lazyComponent(() => import('./batches/global/GlobalBatchChangesArea'), 'GlobalBatchChangesArea'),
        // We also render this route on sourcegraph.com as a precaution in case anyone
        // follows an in-app link to /batch-changes from sourcegraph.com; the component
        // will just redirect the visitor to the marketing page
        condition: ({ batchChangesEnabled, isSourcegraphDotCom }) => batchChangesEnabled || isSourcegraphDotCom,
    },
    {
        path: EnterprisePageRoutes.STATS,
        render: lazyComponent(() => import('./search/stats/SearchStatsPage'), 'SearchStatsPage'),
    },
    {
        path: EnterprisePageRoutes.CODE_MONITORING,
        render: lazyComponent(
            () => import('./code-monitoring/global/GlobalCodeMonitoringArea'),
            'GlobalCodeMonitoringArea'
        ),
    },
    {
        path: EnterprisePageRoutes.INSIGHTS,
        render: lazyComponent(() => import('./insights/InsightsRouter'), 'InsightsRouter'),
        condition: props => isCodeInsightsEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.CONTEXTS,
        render: lazyComponent(() => import('./searchContexts/SearchContextsListPage'), 'SearchContextsListPage'),
        exact: true,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.CREATE_CONTEXT,
        render: lazyComponent(() => import('./searchContexts/CreateSearchContextPage'), 'CreateSearchContextPage'),
        exact: true,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.EDIT_CONTEXT,
        render: lazyComponent(() => import('./searchContexts/EditSearchContextPage'), 'EditSearchContextPage'),
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.CONTEXT,
        render: lazyComponent(() => import('./searchContexts/SearchContextPage'), 'SearchContextPage'),
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    ...routes,
]
