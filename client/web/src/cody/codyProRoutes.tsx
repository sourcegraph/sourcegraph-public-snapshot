import { Navigate, type RouteObject } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { type LegacyLayoutRouteContext, LegacyRoute } from '../LegacyRouteContext'

import { QueryClientProvider } from './management/api/react-query/QueryClientProvider'
import { useUserCodySubscription } from './subscription/useUserCodySubscription'
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

const routeComponents = {
    [CodyProRoutes.NewProSubscription]: lazyComponent(
        () => import('./management/subscription/new/NewCodyProSubscriptionPage'),
        'NewCodyProSubscriptionPage'
    ),
    [CodyProRoutes.Manage]: lazyComponent(() => import('./management/CodyManagementPage'), 'CodyManagementPage'),
    [CodyProRoutes.Subscription]: lazyComponent(
        () => import('./subscription/CodySubscriptionPage'),
        'CodySubscriptionPage'
    ),
    [CodyProRoutes.SubscriptionManage]: lazyComponent(
        () => import('./management/subscription/manage/CodySubscriptionManagePage'),
        'CodySubscriptionManagePage'
    ),
    [CodyProRoutes.ManageTeam]: lazyComponent(() => import('./team/CodyManageTeamPage'), 'CodyManageTeamPage'),
}

interface CodyProPageProps extends Pick<LegacyLayoutRouteContext, 'authenticatedUser' | 'telemetryRecorder'> {
    path: CodyProRoutes
}

/**
 * Renders the appropriate Cody Pro page component for the given route path.
 *
 * This is to more easily isolate the Cody Pro-specific functionality (which
 * only applies to non-Enterprise users) from the rest of the Sourcegraph UI.
 */
const CodyProPage: React.FC<CodyProPageProps> = props => {
    const { data } = useUserCodySubscription()
    if (!data) {
        return null
    }

    if (!data.currentUser) {
        // Cody plan is not available if the user is not authenticated. Redirecting to the sign-in page.
        return <Navigate to={`/sign-in?returnTo=${CodyProRoutes.Manage}`} replace={true} />
    }

    if (!data.currentUser.codySubscription) {
        logger.error('Cody subscription data is not available.')
        return null
    }

    const Component = routeComponents[props.path]
    return (
        <QueryClientProvider>
            <Component
                authenticatedUser={props.authenticatedUser}
                telemetryRecorder={props.telemetryRecorder}
                codySubscription={data.currentUser.codySubscription}
            />
        </QueryClientProvider>
    )
}
