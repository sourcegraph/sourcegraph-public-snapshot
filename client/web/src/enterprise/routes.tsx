import React from 'react'
import { Redirect } from 'react-router'

import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { isCodeInsightsEnabled } from '../insights/utils/is-code-insights-enabled'
import { LayoutRouteProps, routes } from '../routes'
import { lazyComponent } from '../util/lazyComponent'

const isSearchContextsManagementEnabled = (settingsCascade: SettingsCascadeOrError): boolean =>
    !isErrorLike(settingsCascade.final) &&
    settingsCascade.final?.experimentalFeatures?.showSearchContext !== false &&
    settingsCascade.final?.experimentalFeatures?.showSearchContextManagement !== false

export const enterpriseRoutes: readonly LayoutRouteProps<any>[] = [
    {
        // Allow unauthenticated viewers to view the "new subscription" page to price out a subscription (instead
        // of just dumping them on a sign-in page).
        path: '/subscriptions/new',
        exact: true,
        render: lazyComponent(
            () => import('./user/productSubscriptions/NewProductSubscriptionPageOrRedirectUser'),
            'NewProductSubscriptionPageOrRedirectUser'
        ),
    },
    {
        // Redirect from old /user/subscriptions/new -> /subscriptions/new.
        path: '/user/subscriptions/new',
        exact: true,
        render: () => <Redirect to="/subscriptions/new" />,
    },
    {
        path: '/batch-changes',
        render: lazyComponent(() => import('./batches/global/GlobalBatchChangesArea'), 'GlobalBatchChangesArea'),
        // We also render this route on sourcegraph.com as a precaution in case anyone
        // follows an in-app link to /batch-changes from sourcegraph.com; the component
        // will just redirect the visitor to the marketing page
        condition: ({ batchChangesEnabled, isSourcegraphDotCom }) => batchChangesEnabled || isSourcegraphDotCom,
    },
    {
        path: '/stats',
        render: lazyComponent(() => import('./search/stats/SearchStatsPage'), 'SearchStatsPage'),
    },
    {
        path: '/code-monitoring',
        render: lazyComponent(
            () => import('./code-monitoring/global/GlobalCodeMonitoringArea'),
            'GlobalCodeMonitoringArea'
        ),
    },
    {
        path: '/insights',
        render: lazyComponent(() => import('./insights/InsightsRouter'), 'InsightsRouter'),
        condition: props => isCodeInsightsEnabled(props.settingsCascade),
    },
    {
        path: '/contexts',
        render: lazyComponent(() => import('./searchContexts/SearchContextsListPage'), 'SearchContextsListPage'),
        exact: true,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: '/contexts/new',
        render: lazyComponent(() => import('./searchContexts/CreateSearchContextPage'), 'CreateSearchContextPage'),
        exact: true,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: '/contexts/:spec+/edit',
        render: lazyComponent(() => import('./searchContexts/EditSearchContextPage'), 'EditSearchContextPage'),
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: '/contexts/:spec+',
        render: lazyComponent(() => import('./searchContexts/SearchContextPage'), 'SearchContextPage'),
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    ...routes,
]
