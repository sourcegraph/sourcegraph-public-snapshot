import { RepoContainerRoute } from '../../repo/RepoContainer'
import { RepoRevContainerRoute } from '../../repo/RepoRevContainer'
import { repoContainerRoutes, repoRevContainerRoutes } from '../../repo/routes'
import { lazyComponent } from '../../util/lazyComponent'

export const enterpriseRepoContainerRoutes: ReadonlyArray<RepoContainerRoute> = [
    ...repoContainerRoutes,
    {
        path: '/-/issues',
        render: lazyComponent(() => import('../issues/repository/RepositoryIssuesArea'), 'RepositoryIssuesArea'),
    },
    {
        path: '/-/changesets',
        render: lazyComponent(
            () => import('../threads/repository/RepositoryChangesetsArea'),
            'RepositoryChangesetsArea'
        ),
    },
]

export const enterpriseRepoRevContainerRoutes: ReadonlyArray<RepoRevContainerRoute> = repoRevContainerRoutes
