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
 * Generally available Cody Pro routes.
 */
const stableRoutes = new Set([CodyProRoutes.Manage, CodyProRoutes.Subscription])

/**
 * Determines if a given Cody Pro route should be rendered.
 * If the embedded Cody Pro UI is enabled, all routes including experimental are rendered.
 * Otherwise, only the generally available routes are rendered.
 */
const isRouteEnabled = (path: CodyProRoutes): boolean => (isEmbeddedCodyProUIEnabled() ? true : stableRoutes.has(path))

const CodyProPage = lazyComponent(() => import('./CodyProPage'), 'CodyProPage')

export const codyProRoutes: RouteObject[] = Object.values(CodyProRoutes).map(path => ({
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
            condition={({ isSourcegraphDotCom, licenseFeatures }) =>
                isSourcegraphDotCom && licenseFeatures.isCodyEnabled && isRouteEnabled(path)
            }
        />
    ),
}))
