import { RepoContainerRoute } from '../../repo/RepoContainer'
import { RepoRevContainerRoute } from '../../repo/RepoRevContainer'
import { repoContainerRoutes, repoRevContainerRoutes } from '../../repo/routes'
import { lazyComponent } from '../../util/lazyComponent'

export const enterpriseRepoContainerRoutes: ReadonlyArray<RepoContainerRoute> = [
    ...repoContainerRoutes,
    {
        path: '/-/threads',
        render: lazyComponent(() => import('../threads/repository/RepositoryThreadsArea'), 'RepositoryThreadsArea'),
    },
    {
        path: '/-/labels',
        render: lazyComponent(() => import('../labels/RepositoryLabelsPage'), 'RepositoryLabelsPage'),
    },
]

export const enterpriseRepoRevContainerRoutes: ReadonlyArray<RepoRevContainerRoute> = repoRevContainerRoutes
