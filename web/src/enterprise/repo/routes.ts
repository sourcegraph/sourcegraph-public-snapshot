import { RepoContainerRoute } from '../../repo/RepoContainer'
import { RepoRevContainerRoute } from '../../repo/RepoRevContainer'
import { repoContainerRoutes, repoRevContainerRoutes } from '../../repo/routes'
import { lazyComponent } from '../../util/lazyComponent'

export const enterpriseRepoContainerRoutes: readonly RepoContainerRoute[] = [
    ...repoContainerRoutes,
    {
        path: '/-/threads',
        render: lazyComponent(() => import('../threads/repository/RepositoryThreadsArea'), 'RepositoryThreadsArea'),
    },
    {
        path: '/-/labels',
        render: lazyComponent(() => import('../labels/list/RepositoryLabelsPage'), 'RepositoryLabelsPage'),
    },
]

export const enterpriseRepoRevContainerRoutes: readonly RepoRevContainerRoute[] = repoRevContainerRoutes
