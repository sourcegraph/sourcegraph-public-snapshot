import React from 'react'
import { Redirect } from 'react-router'
import { LayoutRouteProps, routes } from '../routes'
import { lazyComponent } from '../util/lazyComponent'

export const enterpriseRoutes: ReadonlyArray<LayoutRouteProps> = [
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
        path: '/tasks',
        render: lazyComponent(() => import('./tasks/global/TasksArea'), 'TasksArea'),
    },
    {
        path: '/changesets',
        render: lazyComponent(() => import('./changesets/global/ChangesetsArea'), 'ChangesetsArea'),
    },
    {
        path: '/threads',
        render: lazyComponent(() => import('./threads/global/ThreadsArea'), 'ThreadsArea'),
    },
    {
        path: '/checks',
        render: lazyComponent(() => import('./checks/global/ChecksArea'), 'ChecksArea'),
    },
    {
        path: '/changes',
        render: lazyComponent(() => import('./changes/global/ChangesArea'), 'ChangesArea'),
    },
    {
        path: '/activity',
        render: lazyComponent(() => import('./activity/global/ActivityArea'), 'ActivityArea'),
    },
    ...routes,
]
