import { Navigate, type RouteObject } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { LegacyRoute } from '../LegacyRouteContext'
import { PageRoutes } from '../routes.constants'

import { CodyIgnoreProvider } from './useCodyIgnore'

const CodyChatPage = lazyComponent(() => import('./chat/CodyChatPage'), 'CodyChatPage')
const CodySwitchAccountPage = lazyComponent(
    () => import('./switch-account/CodySwitchAccountPage'),
    'CodySwitchAccountPage'
)
const CodyDashboardPage = lazyComponent(() => import('./dashboard/CodyDashboardPage'), 'CodyDashboardPage')

/**
 * Use {@link codyProRoutes} for Cody PLG routes.
 */
export const codyRoutes: RouteObject[] = [
    {
        path: PageRoutes.CodyRedirectToMarketingOrDashboard,
        element: (
            <LegacyRoute
                render={({ isSourcegraphDotCom }) => (
                    <Navigate
                        to={isSourcegraphDotCom ? 'https://sourcegraph.com/cody' : PageRoutes.CodyDashboard}
                        replace={true}
                    />
                )}
                condition={() => window.context?.codyEnabledOnInstance}
            />
        ),
    },
    {
        path: PageRoutes.CodySwitchAccount,
        element: (
            <LegacyRoute
                render={props => (
                    <CodySwitchAccountPage {...props} telemetryRecorder={props.platformContext.telemetryRecorder} />
                )}
                condition={() => window.context?.codyEnabledOnInstance}
            />
        ),
    },
    {
        path: `${PageRoutes.CodyChat}/*`,
        element: (
            <LegacyRoute
                render={props => (
                    <CodyIgnoreProvider isSourcegraphDotCom={props.isSourcegraphDotCom}>
                        <CodyChatPage
                            {...props}
                            context={window.context}
                            telemetryRecorder={props.platformContext.telemetryRecorder}
                        />
                    </CodyIgnoreProvider>
                )}
                condition={() => window.context?.codyEnabledOnInstance}
            />
        ),
    },
    {
        path: PageRoutes.CodyDashboard,
        element: (
            <LegacyRoute
                render={props => <CodyDashboardPage {...props} />}
                condition={() => window.context?.codyEnabledOnInstance}
            />
        ),
    },
]
