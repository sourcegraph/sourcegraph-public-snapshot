import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { LayoutRouteProps } from '../routes'

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
export const communitySearchContextsRoutes: readonly LayoutRouteProps<any>[] = [
    {
        isV6: false,
        path: '/kubernetes',
        render: props => <KubernetesCommunitySearchContextPage {...props} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        isV6: false,
        path: '/stackstorm',
        render: props => <StackstormCommunitySearchContextPage {...props} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        isV6: false,
        path: '/temporal',
        render: props => <TemporalCommunitySearchContextPage {...props} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        isV6: false,
        path: '/o3de',
        render: props => <O3deCommunitySearchContextPage {...props} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        isV6: false,
        path: '/chakraui',
        render: props => <ChakraUICommunitySearchContextPage {...props} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        isV6: false,
        path: '/stanford',
        render: props => <StanfordCommunitySearchContextPage {...props} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        isV6: false,
        path: '/cncf',
        render: props => <CncfCommunitySearchContextPage {...props} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        isV6: false,
        path: '/julia',
        render: props => <JuliaCommunitySearchContextPage {...props} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        isV6: false,
        path: '/backstage',
        render: props => <BackstageCommunitySearchContextPage {...props} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
]
