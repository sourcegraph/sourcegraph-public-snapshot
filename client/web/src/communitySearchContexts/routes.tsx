import type { RouteObject } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { LegacyRoute } from '../LegacyRouteContext'

const KubernetesCommunitySearchContextPage = lazyComponent(
    () => import('./Kubernetes'),
    'KubernetesCommunitySearchContextPage'
)
const StackstormCommunitySearchContextPage = lazyComponent(
    () => import('./StackStorm'),
    'StackStormCommunitySearchContextPage'
)
const TemporalCommunitySearchContextPage = lazyComponent(
    () => import('./Temporal'),
    'TemporalCommunitySearchContextPage'
)
const O3deCommunitySearchContextPage = lazyComponent(() => import('./o3de'), 'O3deCommunitySearchContextPage')
const ChakraUICommunitySearchContextPage = lazyComponent(
    () => import('./chakraui'),
    'ChakraUICommunitySearchContextPage'
)
const StanfordCommunitySearchContextPage = lazyComponent(
    () => import('./Stanford'),
    'StanfordCommunitySearchContextPage'
)
const CncfCommunitySearchContextPage = lazyComponent(() => import('./cncf'), 'CncfCommunitySearchContextPage')
const JuliaCommunitySearchContextPage = lazyComponent(() => import('./Julia'), 'JuliaCommunitySearchContextPage')
const BackstageCommunitySearchContextPage = lazyComponent(
    () => import('./Backstage'),
    'BackstageCommunitySearchContextPage'
)

// Hack! Hardcode these routes into cmd/frontend/internal/app/ui/router.go
export const communitySearchContextsRoutes: readonly RouteObject[] = [
    {
        path: '/kubernetes',
        element: (
            <LegacyRoute
                render={props => <KubernetesCommunitySearchContextPage {...props} />}
                condition={({ isSourcegraphDotCom }) => isSourcegraphDotCom}
            />
        ),
    },
    {
        path: '/stackstorm',
        element: (
            <LegacyRoute
                render={props => <StackstormCommunitySearchContextPage {...props} />}
                condition={({ isSourcegraphDotCom }) => isSourcegraphDotCom}
            />
        ),
    },
    {
        path: '/temporal',
        element: (
            <LegacyRoute
                render={props => <TemporalCommunitySearchContextPage {...props} />}
                condition={({ isSourcegraphDotCom }) => isSourcegraphDotCom}
            />
        ),
    },
    {
        path: '/o3de',
        element: (
            <LegacyRoute
                render={props => <O3deCommunitySearchContextPage {...props} />}
                condition={({ isSourcegraphDotCom }) => isSourcegraphDotCom}
            />
        ),
    },
    {
        path: '/chakraui',
        element: (
            <LegacyRoute
                render={props => <ChakraUICommunitySearchContextPage {...props} />}
                condition={({ isSourcegraphDotCom }) => isSourcegraphDotCom}
            />
        ),
    },
    {
        path: '/stanford',
        element: (
            <LegacyRoute
                render={props => <StanfordCommunitySearchContextPage {...props} />}
                condition={({ isSourcegraphDotCom }) => isSourcegraphDotCom}
            />
        ),
    },
    {
        path: '/cncf',
        element: (
            <LegacyRoute
                render={props => <CncfCommunitySearchContextPage {...props} />}
                condition={({ isSourcegraphDotCom }) => isSourcegraphDotCom}
            />
        ),
    },
    {
        path: '/julia',
        element: (
            <LegacyRoute
                render={props => <JuliaCommunitySearchContextPage {...props} />}
                condition={({ isSourcegraphDotCom }) => isSourcegraphDotCom}
            />
        ),
    },
    {
        path: '/backstage',
        element: (
            <LegacyRoute
                render={props => <BackstageCommunitySearchContextPage {...props} />}
                condition={({ isSourcegraphDotCom }) => isSourcegraphDotCom}
            />
        ),
    },
]
