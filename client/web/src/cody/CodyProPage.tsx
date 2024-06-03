import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { LegacyLayoutRouteContext } from '../LegacyRouteContext'

import { CodyProRoutes } from './codyProRoutes'

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
    [CodyProRoutes.AcceptInvite]: lazyComponent(() => import('./invites/AcceptInvitePage'), 'CodyAcceptInvitePage'),
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
export const CodyProPage: React.FC<CodyProPageProps> = props => {
    const Component = routeComponents[props.path]
    return <Component authenticatedUser={props.authenticatedUser} telemetryRecorder={props.telemetryRecorder} />
}
