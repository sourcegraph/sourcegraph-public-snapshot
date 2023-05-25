import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RedirectRoute } from '../../components/RedirectRoute'
import { RepoContainerRoute } from '../../repo/RepoContainer'
import { repoContainerRoutes } from '../../repo/repoContainerRoutes'

const RepositoryCodeIntelArea = lazyComponent(
    () => import('../codeintel/repo/RepositoryCodeIntelArea'),
    'RepositoryCodeIntelArea'
)

const CodyRepoArea = lazyComponent(() => import('../cody/repo/CodyRepoArea'), 'CodyRepoArea')

const RepositoryBatchChangesArea = lazyComponent(
    () => import('../batches/repo/RepositoryBatchChangesArea'),
    'RepositoryBatchChangesArea'
)

const RepositoryOwnPage = lazyComponent(() => import('../own/RepositoryOwnPage'), 'RepositoryOwnPage')

export const enterpriseRepoContainerRoutes: readonly RepoContainerRoute[] = [
    ...repoContainerRoutes,

    {
        path: '/-/code-intelligence/*',
        render: () => (
            <RedirectRoute
                getRedirectURL={({ location }) => location.pathname.replace('/code-intelligence', '/code-graph')}
            />
        ),
    },
    {
        path: '/-/code-graph/*',
        render: context => <RepositoryCodeIntelArea {...context} />,
    },
    {
        path: '/-/embeddings/*',
        render: context => <CodyRepoArea {...context} />,
    },
    {
        path: '/-/batch-changes',
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        render: context => <RepositoryBatchChangesArea {...context} />,
    },
    {
        path: '/-/own',
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom,
        render: context => <RepositoryOwnPage {...context} />,
    },
]
