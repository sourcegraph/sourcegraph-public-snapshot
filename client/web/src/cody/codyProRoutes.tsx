import type { RouteObject } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { LegacyRoute } from '../LegacyRouteContext'

import { isEmbeddedCodyProUIEnabled } from './util'

export enum CodyProRoutes {
    // The checkout form for a new Cody Pro subscription.
    NewProSubscription = '/cody/manage/subscription/new',
    // The Manage page is labeled as the "Dashboard" page.
    Manage = '/cody/manage',
    // The Subscriptions page is a comparison of different Cody product tiers.
    Subscription = '/cody/subscription',
    SubscriptionManage = '/cody/subscription/manage',

    ManageTeam = '/cody/team/manage',
    // Accepts an invite to join a Cody team, then redirects to the Cody team page.
    AcceptInvite = '/cody/invites/accept',
}

/**
 * Defines the routes to be rendered in the Cody Pro UI.
 * If the embedded Cody Pro UI is enabled, all available routes are rendered.
 * Otherwise, only the Manage and Subscription routes are rendered.
 */
const routes = isEmbeddedCodyProUIEnabled()
    ? Object.values(CodyProRoutes)
    : [CodyProRoutes.Manage, CodyProRoutes.Subscription]

const CodyProPage = lazyComponent(() => import('./CodyProPage'), 'CodyProPage')

export const codyProRoutes: RouteObject[] = routes.map(path => ({
    path,
    element: (
        <LegacyRoute
            render={props => (
                <CodyProPage
                    path={path}
                    authenticatedUser={props.authenticatedUser}
                    telemetryRecorder={props.platformContext.telemetryRecorder}
                />
            )}
            condition={({ isSourcegraphDotCom }) => isSourcegraphDotCom}
        />
    ),
}))
