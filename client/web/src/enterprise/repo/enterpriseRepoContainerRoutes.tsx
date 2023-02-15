import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RedirectRoute } from '../../components/RedirectRoute'
import { RepoContainerRoute } from '../../repo/RepoContainer'
import { repoContainerRoutes } from '../../repo/repoContainerRoutes'

const RepositoryCodeIntelArea = lazyComponent(
    () => import('../codeintel/repo/RepositoryCodeIntelArea'),
    'RepositoryCodeIntelArea'
)

const RepositoryBatchChangesArea = lazyComponent(
    () => import('../batches/repo/RepositoryBatchChangesArea'),
    'RepositoryBatchChangesArea'
)

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
        path: '/-/batch-changes',
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        render: context => <RepositoryBatchChangesArea {...context} />,
    },
]
